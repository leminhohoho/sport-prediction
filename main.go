package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/gocolly/colly"
	"github.com/leminhohoho/sport-prediction/runner/controller"
	"github.com/leminhohoho/sport-prediction/runner/scheduler"
	"github.com/leminhohoho/sport-prediction/runner/scraper"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskScheduler := scheduler.NewScheduler(3, ctx, true)

	crawler := colly.NewCollector(colly.AllowedDomains("www.vlr.gg"))
	scrapebot := scraper.NewScraper(10)

	db, err := sql.Open("sqlite3", "vlr2.db")
	if err != nil {
		log.Fatal(err)
	}

	c := controller.NewController(taskScheduler, db, scrapebot, crawler)

	matches, err := c.ScrapeMatches(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, match := range matches {
		fmt.Println(match.Id)
		fmt.Println(match.Url)
		fmt.Println(match.Date.Format("2006-01-02"))
		fmt.Println(match.TournamentId)
		fmt.Println(match.Team1Id)
		fmt.Println(match.Team2Id)
		fmt.Println(match.Stage)
		fmt.Println(match.TeamWon)
	}
}
