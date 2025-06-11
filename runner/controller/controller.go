package controller

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/leminhohoho/sport-prediction/runner/helpers"
	"github.com/leminhohoho/sport-prediction/runner/models"
	"github.com/leminhohoho/sport-prediction/runner/patterns"
	"github.com/leminhohoho/sport-prediction/runner/scheduler"
	"github.com/leminhohoho/sport-prediction/runner/scraper"
	"github.com/leminhohoho/sport-prediction/runner/selectors"
)

type Controller struct {
	taskScheduler *scheduler.Scheduler
	db            *sql.DB
	scrapebot     *scraper.Scraper
	crawler       *colly.Collector
}

func NewController(
	taskScheduler *scheduler.Scheduler,
	db *sql.DB,
	scrapebot *scraper.Scraper,
	crawler *colly.Collector,
) *Controller {
	return &Controller{
		taskScheduler: taskScheduler,
		db:            db,
		scrapebot:     scrapebot,
		crawler:       crawler,
	}
}

func (c *Controller) CrawlMatches() ([]models.Matches, error) {
	page := 1
	layoutFrom := "Mon, January 2, 2006"
	var matches []models.Matches
	var crawlErr error

	c.crawler.OnScraped(func(res *colly.Response) {
		if page == 100 {
			log.Printf("Number of matches:%d\n", len(matches))
			return
		}

		page++
		fmt.Printf("Page: %d\n", page)
		c.crawler.Visit(fmt.Sprintf("https://www.vlr.gg/matches/results/?page=%d", page))
	})

	c.crawler.OnError(func(res *colly.Response, err error) {
		crawlErr = err
	})

	c.crawler.OnResponse(func(r *colly.Response) {
		html := string(r.Body)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			crawlErr = err
		}

		doc.Find(selectors.MatchDate).
			Each(func(i int, s *goquery.Selection) {
				if regexp.MustCompile(patterns.DatePattern).
					MatchString(strings.TrimSpace(s.Text())) {
					fmt.Println(strings.TrimSpace(s.Text()))
					date, err := time.Parse(layoutFrom, strings.TrimSpace(s.Text()))
					if err != nil {
						crawlErr = err
					}

					currentDayMatches := s.Next()
					currentDayMatches.Find("a[href].match-item").Each(func(j int, ns *goquery.Selection) {
						link, exist := ns.Attr("href")
						if exist {
							id := strings.Split(link, "/")[1]

							matches = append(matches, models.Matches{
								Date: date,
								Url:  "https://www.vlr.gg" + link,
								Id:   id,
							})
						}
					})
				}
			})
	})

	c.crawler.Visit(fmt.Sprintf("https://www.vlr.gg/matches/results/?page=%d", page))

	if crawlErr != nil {
		return nil, crawlErr
	}

	return matches, nil
}

func (c *Controller) ScrapeTournament(ctx context.Context, tournamentUrl string) error {
	tournamentUrlFragment := strings.Split(tournamentUrl, "/")
	tournamentIdStr := tournamentUrlFragment[2]
	tournamentId, err := strconv.Atoi(tournamentIdStr)
	if err != nil {
		return err
	}

	fmt.Println("Start scraping tournament")
	htmlContent, err := c.scrapebot.Scrape(ctx, "https://www.vlr.gg"+tournamentUrl, selectors.VLRggIndicator)
	if err != nil {
		return err
	}
	fmt.Println("Finished scraping tournament")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return err
	}

	tournamentName := strings.TrimSpace(doc.Find(selectors.TournamentName).Text())

	prizeStr := regexp.MustCompile(`[0-9,]+ USD`).FindString(doc.Find(selectors.TournamentPrize).Text())
	prizeStr = strings.Replace(prizeStr, ",", "", -1)
	prizeStr = strings.Replace(prizeStr, " USD", "", -1)
	prizePool, err := strconv.Atoi(prizeStr)
	if err != nil {
		return err
	}

	teamsContainer := doc.Find(selectors.TournamentTeamsContainer)
	numberOfTeams := teamsContainer.Children().Length()

	isBigEvent := helpers.IsBigEvent(tournamentName, prizePool)

	_, err = c.db.Exec(`INSERT OR IGNORE INTO tournaments(
		id,
		url,
		name,
		prize_pool,
		number_of_teams,
		tier
	)
	VALUES(?,?,?,?,?,?)
	`,
		tournamentId,
		"https://www.vlr.gg"+tournamentUrl,
		tournamentName,
		prizePool,
		numberOfTeams,
		isBigEvent,
	)
	if err != nil {
		return err
	}

	fmt.Printf("Tournament id: %d\n", tournamentId)
	fmt.Printf("Tournament name: %s\n", tournamentName)
	fmt.Printf("Prize: %d USD\n", prizePool)
	fmt.Printf("Number of teams: %d\n", numberOfTeams)

	return nil
}

func (c *Controller) ScrapeMatches(ctx context.Context) ([]models.Matches, error) {
	var err error

	matches, err := c.CrawlMatches()
	if err != nil {
		return nil, err
	}

	var tournamentUrls []string
	var playerUrls []string
	var teamUrls []string

	// // NOTE: Temporary
	// matches = matches[:3]

	taskScheduler := scheduler.NewScheduler(3, ctx, false)

	var schedulerActions []scheduler.Action

	for i, match := range matches {
		if i%50 == 0 && i > 0 {
			schedulerActions = append(
				schedulerActions,
				scheduler.ActionFunc(func(ctx context.Context) error {
					fmt.Println("##################")
					fmt.Println("### BREAK TIME ###")
					fmt.Println("##################")

					return nil
				}),
				scheduler.Sleep(time.Second*30),
			)
		}
		schedulerActions = append(
			schedulerActions,
			scheduler.Sleep(time.Second*1),
		)

		schedulerActions = append(
			schedulerActions,
			scheduler.Async{A: scheduler.ActionFunc(func(ctx context.Context) error {
				overviewHTMLContent, err := c.scrapebot.Scrape(ctx, match.Url, selectors.VLRggIndicator)
				performanceHTMLContent, err := c.scrapebot.Scrape(
					ctx,
					match.Url+`/?tab=performance`,
					selectors.VLRggIndicator,
				)

				doc1, err := goquery.NewDocumentFromReader(strings.NewReader(overviewHTMLContent))
				if err != nil {
					return err
				}
				doc2, err := goquery.NewDocumentFromReader(strings.NewReader(performanceHTMLContent))
				if err != nil {
					return err
				}

				tournamentUrl, exists := doc1.Find(selectors.MatchTournamentAnchor).Attr("href")
				if exists {
					matches[i].TournamentId = strings.Split(tournamentUrl, "/")[2]
				}

				if !slices.Contains(tournamentUrls, tournamentUrl) {
					tournamentUrls = append(tournamentUrls, tournamentUrl)
				}

				team1Url, exists := doc1.Find(selectors.MatchTeam1Anchor).Attr("href")
				if exists {
					matches[i].Team1Id = strings.Split(team1Url, "/")[2]
				}

				if !slices.Contains(teamUrls, team1Url) {
					teamUrls = append(teamUrls, team1Url)
				}

				team2Url, exists := doc1.Find(selectors.MatchTeam2Anchor).Attr("href")
				if exists {
					matches[i].Team2Id = strings.Split(team2Url, "/")[2]
				}

				if !slices.Contains(teamUrls, team2Url) {
					teamUrls = append(teamUrls, team2Url)
				}

				matchHeader := strings.ToLower(doc1.Find(selectors.MatchHeader).Text())
				if strings.Contains(matchHeader, "grand final") {
					matches[i].Stage = "Grand final"
				} else if strings.Contains(matchHeader, "playoffs") || strings.Contains(matchHeader, "main event") {
					matches[i].Stage = "Playoff"
				} else {
					matches[i].Stage = "Group stage"
				}
				team1Result, err := strconv.Atoi(
					strings.TrimSpace(doc1.Find(selectors.Team1OverallResult).Text()),
				)
				team2Result, err := strconv.Atoi(
					strings.TrimSpace(doc1.Find(selectors.Team2OverallResult).Text()),
				)
				if err != nil {
					return err
				}

				if team1Result > team2Result {
					matches[i].TeamWon = matches[i].Team1Id
				} else {
					matches[i].TeamWon = matches[i].Team2Id
				}

				// NOTE: Query to update matches table
				matchQuery := fmt.Sprintf(
					`INSERT INTO matches(id, url, date, tournament_id, stage, team_1_id, team_2_id, team_won) VALUES(%s,"%s","%s",%s,"%s",%s,%s,%s)`,
					matches[i].Id,
					matches[i].Url,
					matches[i].Date.Format("2006-01-01"),
					matches[i].TournamentId,
					matches[i].Stage,
					matches[i].Team1Id,
					matches[i].Team2Id,
					matches[i].TeamWon,
				)

				log.Println(matchQuery)

				fmt.Printf("Match: %s\n", match.Url)
				helpers.Find2Docs(
					selectors.MatchMapGenericSelector,
					doc1,
					doc2,
					func(node1 *goquery.Selection, node2 *goquery.Selection) {
						var matchMap models.MatchesMaps

						matchMap.MatchId = matches[i].Id

						matchMap.MapName = strings.TrimSpace(
							strings.Replace(
								node1.Find(`div.vm-stats-game-header > div.map > div > span`).Text(),
								"PICK",
								"",
								-1,
							),
						)

						matchMap.Team1Score = strings.TrimSpace(
							node1.Find(`div.vm-stats-game-header > div:nth-child(1) > div.score`).Text(),
						)
						matchMap.Team2Score = strings.TrimSpace(
							node1.Find(`div.vm-stats-game-header > div.team.mod-right > div.score`).Text(),
						)

						node1.Find(`div.vm-stats-game-header > div:nth-child(1) > div:nth-child(2) > span`).
							First().
							Each(func(_ int, node *goquery.Selection) {
								val, _ := node.Attr("class")
								if strings.TrimSpace(val) == "mod-ct" {
									matchMap.TeamDefFirst = matches[i].Team1Id
									matchMap.TeamAtkFirst = matches[i].Team2Id
								} else {
									matchMap.TeamDefFirst = matches[i].Team2Id
									matchMap.TeamAtkFirst = matches[i].Team1Id
								}
							})

						var mapId int
						err = c.db.QueryRow(`SELECT id FROM maps WHERE name = ?`, matchMap.MapName).
							Scan(&mapId)
						if err != nil {
							log.Println(err.Error())
						}

						matchMapQuery := fmt.Sprintf(
							`INSERT INTO matches_maps(match_id,map_id,team_1_score,team_2_score,team_def_first,team_atk_first)
						VALUES(%s,%d,%s,%s,%s,%s)`,
							matchMap.MatchId,
							mapId,
							matchMap.Team1Score,
							matchMap.Team2Score,
							matchMap.TeamDefFirst,
							matchMap.TeamAtkFirst,
						)

						_, err = c.db.Exec(matchMapQuery)

						if err != nil {
							log.Println(err.Error())
						}

						players := [2]map[string]models.MatchesMapsTeamsPlayersPerformance{}

						players[0] = make(map[string]models.MatchesMapsTeamsPlayersPerformance)
						players[1] = make(map[string]models.MatchesMapsTeamsPlayersPerformance)

						// NOTE: The selector is relevant to vm-stat-game
						extractPlayerOverviewStat := func(
							teamId string,
							playerIdSelector string,
							playerAgentImageSelector string,
							playerRatingSelector string,
							playerACSSelector string,
							playerKillsSelector string,
							playerDeathsSelector string,
							playerAssistsSelector string,
							playerKastSelector string,
							playerADRSelector string,
							playerHSPercentageSelector string,
							playerFirstKillsSelector string,
							playerFirstDeathsSelector string,
						) models.MatchesMapsTeamsPlayersPerformance {
							agentName, _ := node1.Find(
								playerAgentImageSelector,
							).Attr("title")

							playerUrl, _ := node1.Find(playerIdSelector).Attr("href")

							var playerId string
							if len(strings.Split(playerUrl, "/")) == 1 {
								playerId = ""
							} else {
								playerId = strings.Split(playerUrl, "/")[2]
							}
							if !slices.Contains(playerUrls, playerUrl) {
								playerUrls = append(playerUrls, playerUrl)
							}
							if err != nil {
								log.Println(err.Error())
							}

							playerRating := node1.Find(playerRatingSelector).Text()
							playerACS := node1.Find(playerACSSelector).Text()
							playerKills := node1.Find(playerKillsSelector).Text()
							playerDeaths := node1.Find(playerDeathsSelector).Text()
							playerAssists := node1.Find(playerAssistsSelector).Text()
							playerKast := node1.Find(playerKastSelector).Text()
							playerADR := node1.Find(playerADRSelector).Text()
							playerHSPercentage := node1.Find(playerHSPercentageSelector).Text()
							playerFirstKills := node1.Find(playerFirstKillsSelector).Text()
							playerFirstDeaths := node1.Find(playerFirstDeathsSelector).Text()

							player := models.MatchesMapsTeamsPlayersPerformance{
								MatchId:      matches[i].Id,
								MapName:      matchMap.MapName,
								TeamId:       teamId,
								PlayerId:     playerId,
								AgentName:    agentName,
								Rating:       playerRating,
								Acs:          playerACS,
								Kills:        playerKills,
								Deaths:       playerDeaths,
								Assists:      playerAssists,
								Kast:         playerKast,
								Adr:          playerADR,
								HSPercentage: playerHSPercentage,
								FirstKills:   playerFirstKills,
								FirstDeaths:  playerFirstDeaths,
							}

							return player
						}

						addPerformanceStat := func(
							teamIndex int,
							playerName string,
							N2KillsSelector string,
							N3KillsSelector string,
							N4KillsSelector string,
							N5KillsSelector string,
							N1v1Selector string,
							N1v2Selector string,
							N1v3Selector string,
							N1v4Selector string,
							N1v5Selector string,
						) {
							n2Kills := strings.TrimSpace(
								node2.Find(N2KillsSelector).Children().Remove().End().Text(),
							)
							if n2Kills == "" {
								n2Kills = "0"
							}
							n3Kills := strings.TrimSpace(
								node2.Find(N3KillsSelector).Children().Remove().End().Text(),
							)
							if n3Kills == "" {
								n3Kills = "0"
							}
							n4Kills := strings.TrimSpace(
								node2.Find(N4KillsSelector).Children().Remove().End().Text(),
							)
							if n4Kills == "" {
								n4Kills = "0"
							}
							n5Kills := strings.TrimSpace(
								node2.Find(N5KillsSelector).Children().Remove().End().Text(),
							)
							if n5Kills == "" {
								n5Kills = "0"
							}
							n1v1 := strings.TrimSpace(
								node2.Find(N1v1Selector).Children().Remove().End().Text(),
							)
							if n1v1 == "" {
								n1v1 = "0"
							}
							n1v2 := strings.TrimSpace(
								node2.Find(N1v2Selector).Children().Remove().End().Text(),
							)
							if n1v2 == "" {
								n1v2 = "0"
							}
							n1v3 := strings.TrimSpace(
								node2.Find(N1v3Selector).Children().Remove().End().Text(),
							)
							if n1v3 == "" {
								n1v3 = "0"
							}
							n1v4 := strings.TrimSpace(
								node2.Find(N1v4Selector).Children().Remove().End().Text(),
							)
							if n1v4 == "" {
								n1v4 = "0"
							}
							n1v5 := strings.TrimSpace(
								node2.Find(N1v5Selector).Children().Remove().End().Text(),
							)
							if n1v5 == "" {
								n1v5 = "0"
							}

							player := players[teamIndex][playerName]
							player.N2Kills = n2Kills
							player.N3Kills = n3Kills
							player.N4Kills = n4Kills
							player.N5Kills = n5Kills
							player.N1v1 = n1v1
							player.N1v2 = n1v2
							player.N1v3 = n1v3
							player.N1v4 = n1v4
							player.N1v5 = n1v5

							players[teamIndex][playerName] = player
						}

						players[0][strings.TrimSpace(node1.Find(`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td.mod-player > div > a > div.text-of`).Text())] = extractPlayerOverviewStat(
							matches[i].Team1Id,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td.mod-player > div > a`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td.mod-agents > div > span > img`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td:nth-child(3) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td:nth-child(4) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td.mod-stat.mod-vlr-kills > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td.mod-stat.mod-vlr-deaths > span > span:nth-child(2) > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td.mod-stat.mod-vlr-assists > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td:nth-child(9) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td:nth-child(10) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td:nth-child(11) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td.mod-stat.mod-fb > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td.mod-stat.mod-fd > span > span.side.mod-both`,
						)

						players[0][strings.TrimSpace(node1.Find(`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td.mod-player > div > a > div.text-of`).Text())] = extractPlayerOverviewStat(
							matches[i].Team1Id,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td.mod-player > div > a`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td.mod-agents > div > span > img`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td:nth-child(3) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td:nth-child(4) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td.mod-stat.mod-vlr-kills > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td.mod-stat.mod-vlr-deaths > span > span:nth-child(2) > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td.mod-stat.mod-vlr-assists > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td:nth-child(9) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td:nth-child(10) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td:nth-child(11) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td.mod-stat.mod-fb > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td.mod-stat.mod-fd > span > span.side.mod-both`,
						)

						players[0][strings.TrimSpace(node1.Find(`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td.mod-player > div > a > div.text-of`).Text())] = extractPlayerOverviewStat(
							matches[i].Team1Id,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td.mod-player > div > a`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td.mod-agents > div > span > img`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td:nth-child(3) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td:nth-child(4) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td.mod-stat.mod-vlr-kills > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td.mod-stat.mod-vlr-deaths > span > span:nth-child(2) > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td.mod-stat.mod-vlr-assists > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td:nth-child(9) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td:nth-child(10) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td:nth-child(11) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td.mod-stat.mod-fb > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td.mod-stat.mod-fd > span > span.side.mod-both`,
						)

						players[0][strings.TrimSpace(node1.Find(`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td.mod-player > div > a > div.text-of`).Text())] = extractPlayerOverviewStat(
							matches[i].Team1Id,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td.mod-player > div > a`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td.mod-agents > div > span > img`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td:nth-child(3) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td:nth-child(4) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td.mod-stat.mod-vlr-kills > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td.mod-stat.mod-vlr-deaths > span > span:nth-child(2) > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td.mod-stat.mod-vlr-assists > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td:nth-child(9) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td:nth-child(10) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td:nth-child(11) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td.mod-stat.mod-fb > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td.mod-stat.mod-fd > span > span.side.mod-both`,
						)

						players[0][strings.TrimSpace(node1.Find(`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td.mod-player > div > a > div.text-of`).Text())] = extractPlayerOverviewStat(
							matches[i].Team1Id,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td.mod-player > div > a`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td.mod-agents > div > span > img`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td:nth-child(3) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td:nth-child(4) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td.mod-stat.mod-vlr-kills > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td.mod-stat.mod-vlr-deaths > span > span:nth-child(2) > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td.mod-stat.mod-vlr-assists > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td:nth-child(9) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td:nth-child(10) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td:nth-child(11) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td.mod-stat.mod-fb > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td.mod-stat.mod-fd > span > span.side.mod-both`,
						)

						players[1][strings.TrimSpace(node1.Find(`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td.mod-player > div > a > div.text-of`).Text())] = extractPlayerOverviewStat(
							matches[i].Team2Id,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td.mod-player > div > a`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td.mod-agents > div > span > img`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td:nth-child(3) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td:nth-child(4) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td.mod-stat.mod-vlr-kills > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td.mod-stat.mod-vlr-deaths > span > span:nth-child(2) > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td.mod-stat.mod-vlr-assists > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td:nth-child(9) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td:nth-child(10) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td:nth-child(11) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td.mod-stat.mod-fb > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td.mod-stat.mod-fd > span > span.side.mod-both`,
						)

						players[1][strings.TrimSpace(node1.Find(`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td.mod-player > div > a > div.text-of`).Text())] = extractPlayerOverviewStat(
							matches[i].Team2Id,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td.mod-player > div > a`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td.mod-agents > div > span > img`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(3) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(4) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td.mod-stat.mod-vlr-kills > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td.mod-stat.mod-vlr-deaths > span > span:nth-child(2) > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td.mod-stat.mod-vlr-assists > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(9) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(10) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(11) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td.mod-stat.mod-fb > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td.mod-stat.mod-fd > span > span.side.mod-both`,
						)

						players[1][strings.TrimSpace(node1.Find(`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td.mod-player > div > a > div.text-of`).Text())] = extractPlayerOverviewStat(
							matches[i].Team2Id,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td.mod-player > div > a`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td.mod-agents > div > span > img`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(3) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(4) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td.mod-stat.mod-vlr-kills > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td.mod-stat.mod-vlr-deaths > span > span:nth-child(2) > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td.mod-stat.mod-vlr-assists > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(9) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(10) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(11) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td.mod-stat.mod-fb > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td.mod-stat.mod-fd > span > span.side.mod-both`,
						)

						players[1][strings.TrimSpace(node1.Find(`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td.mod-player > div > a > div.text-of`).Text())] = extractPlayerOverviewStat(
							matches[i].Team2Id,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td.mod-player > div > a`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td.mod-agents > div > span > img`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(3) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(4) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td.mod-stat.mod-vlr-kills > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td.mod-stat.mod-vlr-deaths > span > span:nth-child(2) > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td.mod-stat.mod-vlr-assists > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(9) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(10) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(11) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td.mod-stat.mod-fb > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td.mod-stat.mod-fd > span > span.side.mod-both`,
						)

						players[1][strings.TrimSpace(node1.Find(`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td.mod-player > div > a > div.text-of`).Text())] = extractPlayerOverviewStat(
							matches[i].Team2Id,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td.mod-player > div > a`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td.mod-agents > div > span > img`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(3) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(4) > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td.mod-stat.mod-vlr-kills > span > span.side.mod-side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td.mod-stat.mod-vlr-deaths > span > span:nth-child(2) > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td.mod-stat.mod-vlr-assists > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(9) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(10) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(11) > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td.mod-stat.mod-fb > span > span.side.mod-both`,
							`div:nth-child(4) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td.mod-stat.mod-fd > span > span.side.mod-both`,
						)

						addPerformanceStat(
							0,
							strings.TrimSpace(
								node2.Find(`div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(1) > div > div`).
									Children().
									Remove().
									End().
									Text(),
							),
							`div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(3) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(4) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(5) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(6) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(7) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(8) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(9) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(10) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(2) > td:nth-child(11) > div`,
						)

						addPerformanceStat(
							0,
							strings.TrimSpace(
								node2.Find(`div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(1) > div > div`).
									Children().
									Remove().
									End().
									Text(),
							),
							`div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(3) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(4) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(5) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(6) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(7) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(8) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(9) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(10) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(3) > td:nth-child(11) > div`,
						)

						addPerformanceStat(
							0,
							strings.TrimSpace(
								node2.Find(`div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(1) > div > div`).
									Children().
									Remove().
									End().
									Text(),
							),
							`div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(3) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(4) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(5) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(6) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(7) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(8) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(9) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(10) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(4) > td:nth-child(11) > div`,
						)

						addPerformanceStat(
							0,
							strings.TrimSpace(
								node2.Find(`div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(1) > div > div`).
									Children().
									Remove().
									End().
									Text(),
							),
							`div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(3) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(4) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(5) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(6) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(7) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(8) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(9) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(10) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(5) > td:nth-child(11) > div`,
						)

						addPerformanceStat(
							0,
							strings.TrimSpace(
								node2.Find(`div:nth-child(2) > table > tbody > tr:nth-child(6) > td:nth-child(1) > div > div`).
									Children().
									Remove().
									End().
									Text(),
							),
							`div:nth-child(2) > table > tbody > tr:nth-child(6) > td:nth-child(3) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(6) > td:nth-child(4) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(6) > td:nth-child(5) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(6) > td:nth-child(6) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(6) > td:nth-child(7) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(6) > td:nth-child(8) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(6) > td:nth-child(9) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(6) > td:nth-child(10) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(6) > td:nth-child(11) > div`,
						)

						addPerformanceStat(
							1,
							strings.TrimSpace(
								node2.Find(`div:nth-child(2) > table > tbody > tr:nth-child(7) > td:nth-child(1) > div > div`).
									Children().
									Remove().
									End().
									Text(),
							),
							`div:nth-child(2) > table > tbody > tr:nth-child(7) > td:nth-child(3) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(7) > td:nth-child(4) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(7) > td:nth-child(5) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(7) > td:nth-child(6) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(7) > td:nth-child(7) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(7) > td:nth-child(8) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(7) > td:nth-child(9) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(7) > td:nth-child(10) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(7) > td:nth-child(11) > div`,
						)

						addPerformanceStat(
							1,
							strings.TrimSpace(
								node2.Find(`div:nth-child(2) > table > tbody > tr:nth-child(8) > td:nth-child(1) > div > div`).
									Children().
									Remove().
									End().
									Text(),
							),
							`div:nth-child(2) > table > tbody > tr:nth-child(8) > td:nth-child(3) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(8) > td:nth-child(4) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(8) > td:nth-child(5) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(8) > td:nth-child(6) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(8) > td:nth-child(7) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(8) > td:nth-child(8) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(8) > td:nth-child(9) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(8) > td:nth-child(10) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(8) > td:nth-child(11) > div`,
						)
						addPerformanceStat(
							1,
							strings.TrimSpace(
								node2.Find(`div:nth-child(2) > table > tbody > tr:nth-child(9) > td:nth-child(1) > div > div`).
									Children().
									Remove().
									End().
									Text(),
							),
							`div:nth-child(2) > table > tbody > tr:nth-child(9) > td:nth-child(3) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(9) > td:nth-child(4) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(9) > td:nth-child(5) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(9) > td:nth-child(6) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(9) > td:nth-child(7) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(9) > td:nth-child(8) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(9) > td:nth-child(9) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(9) > td:nth-child(10) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(9) > td:nth-child(11) > div`,
						)
						addPerformanceStat(
							1,
							strings.TrimSpace(
								node2.Find(`div:nth-child(2) > table > tbody > tr:nth-child(10) > td:nth-child(1) > div > div`).
									Children().
									Remove().
									End().
									Text(),
							),
							`div:nth-child(2) > table > tbody > tr:nth-child(10) > td:nth-child(3) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(10) > td:nth-child(4) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(10) > td:nth-child(5) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(10) > td:nth-child(6) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(10) > td:nth-child(7) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(10) > td:nth-child(8) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(10) > td:nth-child(9) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(10) > td:nth-child(10) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(10) > td:nth-child(11) > div`,
						)
						addPerformanceStat(
							1,
							strings.TrimSpace(
								node2.Find(`div:nth-child(2) > table > tbody > tr:nth-child(11) > td:nth-child(1) > div > div`).
									Children().
									Remove().
									End().
									Text(),
							),
							`div:nth-child(2) > table > tbody > tr:nth-child(11) > td:nth-child(3) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(11) > td:nth-child(4) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(11) > td:nth-child(5) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(11) > td:nth-child(6) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(11) > td:nth-child(7) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(11) > td:nth-child(8) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(11) > td:nth-child(9) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(11) > td:nth-child(10) > div`,
							`div:nth-child(2) > table > tbody > tr:nth-child(11) > td:nth-child(11) > div`,
						)

						fmt.Printf("Map: %s\n", matchMap.MapName)
						for teamIndex := 0; teamIndex < 2; teamIndex++ {
							for playerName, playerPerformance := range players[teamIndex] {
								var agentId int
								err = c.db.QueryRow(`SELECT id FROM agents WHERE name = ?`, playerPerformance.AgentName).
									Scan(&agentId)
								if err != nil {
									log.Println(err.Error())
								}

								playerPerformanceQuery := fmt.Sprintf(`
									INSERT INTO matches_maps_teams_players_performance(
										match_id,
										map_id,
										team_id,
										player_id,
										agent_id,
										rating,
										acs,
										kills,
										deaths,
										assists,
										kast,
										adr,
										hs_percentage,
										first_kills,
										first_deaths,
										"2k",
										"3k",
										"4k",
										"5k",
										"1v1",
										"1v2",
										"1v3",
										"1v4",
										"1v5"
									)
									VALUES(%s,%d,%s,%s,%d,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)`,
									playerPerformance.MatchId,
									mapId,
									playerPerformance.TeamId,
									playerPerformance.PlayerId,
									agentId,
									playerPerformance.Rating,
									playerPerformance.Acs,
									playerPerformance.Kills,
									playerPerformance.Deaths,
									playerPerformance.Assists,
									strings.Replace(playerPerformance.Kast, "%", "", -1),
									playerPerformance.Adr,
									strings.Replace(playerPerformance.HSPercentage, "%", "", -1),
									playerPerformance.FirstKills,
									playerPerformance.FirstDeaths,
									playerPerformance.N2Kills,
									playerPerformance.N3Kills,
									playerPerformance.N4Kills,
									playerPerformance.N5Kills,
									playerPerformance.N1v1,
									playerPerformance.N1v2,
									playerPerformance.N1v3,
									playerPerformance.N1v4,
									playerPerformance.N1v5,
								)

								_, err := c.db.Exec(playerPerformanceQuery)
								if err != nil {
									log.Println(err.Error())
								}

								fmt.Printf("Team %d player %s:\n", teamIndex+1, playerName)
								fmt.Println("	Agent name: " + playerPerformance.AgentName)
								fmt.Println("	Player Id: " + playerPerformance.PlayerId)
								fmt.Println("	Rating: " + playerPerformance.Rating)
								fmt.Println("	ACS: " + playerPerformance.Acs)
								fmt.Printf(
									"	K/D/A: %s/%s/%s\n",
									playerPerformance.Kills,
									playerPerformance.Deaths,
									playerPerformance.Assists,
								)
								fmt.Println("	Kast: " + playerPerformance.Kast)
								fmt.Println("	Adr: " + playerPerformance.Adr)
								fmt.Println("	HS Percentage: " + playerPerformance.HSPercentage)
								fmt.Println("	First Kills: " + playerPerformance.FirstKills)
								fmt.Println("	First Deaths: " + playerPerformance.FirstDeaths)
								fmt.Println("	Number of 2 kills: " + playerPerformance.N2Kills)
								fmt.Println("	Number of 3 kills: " + playerPerformance.N3Kills)
								fmt.Println("	Number of 4 kills: " + playerPerformance.N4Kills)
								fmt.Println("	Number of 5 kills: " + playerPerformance.N5Kills)
								fmt.Println("	Number of 1v1: " + playerPerformance.N1v1)
								fmt.Println("	Number of 1v2: " + playerPerformance.N1v2)
								fmt.Println("	Number of 1v3: " + playerPerformance.N1v3)
								fmt.Println("	Number of 1v4: " + playerPerformance.N1v4)
								fmt.Println("	Number of 1v5: " + playerPerformance.N1v5)
							}
						}

					},
				)

				_, err = c.db.Exec(matchQuery)
				if err != nil {
					log.Println(err.Error())
					return err
				}

				return nil
			})},
		)
	}

	if err = taskScheduler.Run(schedulerActions...); err != nil {
		return nil, err
	}

	fmt.Println("List of tournaments:")
	for _, tournamentUrl := range tournamentUrls {
		fmt.Printf("	%s\n", tournamentUrl)
		err := c.ScrapeTournament(ctx, tournamentUrl)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("List of players:")
	for _, playerUrl := range playerUrls {
		fmt.Printf("	%s\n", playerUrl)
		if len(strings.Split(playerUrl, "/")) < 4 {
			continue
		}
		playerId := strings.Split(playerUrl, "/")[2]
		playerName := strings.Split(playerUrl, "/")[3]
		_, err = c.db.Exec(
			fmt.Sprintf(
				`INSERT INTO players(id,url, name) VALUES(%s,"%s","%s")`,
				playerId,
				"https://www.vlr.gg"+playerUrl,
				playerName,
			),
		)
		if err != nil {
			log.Println(err.Error())
		}
	}

	fmt.Println("List of teams:")
	for _, teamUrl := range teamUrls {
		fmt.Printf("	%s\n", teamUrl)
		if len(strings.Split(teamUrl, "/")) < 4 {
			continue
		}
		teamId := strings.Split(teamUrl, "/")[2]
		teamName := strings.Split(teamUrl, "/")[3]
		_, err = c.db.Exec(
			fmt.Sprintf(
				`INSERT INTO teams(id,url, name) VALUES(%s,"%s","%s")`,
				teamId,
				"https://www.vlr.gg"+teamUrl,
				teamName,
			),
		)
		if err != nil {
			log.Println(err.Error())
		}
	}

	return matches, nil
}
