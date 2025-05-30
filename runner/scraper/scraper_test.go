package scraper

import (
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

func TestScraper(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	scraper := NewScraper(
		42,
		3,
		"http://100.106.3.17:8888",
	)

	content, err := scraper.Scrape(
		ctx,
		"https://www.hltv.org/matches/2382694/vitality-vs-falcons-iem-dallas-2025",
		"div.hltv-logo-container",
	)
	if err != nil {
		t.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	team1Name := doc.Find("body > div.bgPadding > div.widthControl > div:nth-child(2) > div.contentCol > div.match-page > div.standard-box.teamsBox > div:nth-child(1) > div > a > div").
		Text()
	if team1Name == "" {
		t.Errorf("team 1's name is not found\n")
	}
	team2Name := doc.Find("body > div.bgPadding > div.widthControl > div:nth-child(2) > div.contentCol > div.match-page > div.standard-box.teamsBox > div:nth-child(3) > div > a > div").
		Text()
	if team2Name == "" {
		t.Errorf("team 2's name is not found\n")
	}
	team1Result := doc.Find("body > div.bgPadding > div.widthControl > div:nth-child(2) > div.contentCol > div.match-page > div.standard-box.teamsBox > div:nth-child(1) > div > div").
		Text()
	if team1Result == "" {
		t.Errorf("team 1's result is not found\n")
	}
	team2Result := doc.Find("body > div.bgPadding > div.widthControl > div:nth-child(2) > div.contentCol > div.match-page > div.standard-box.teamsBox > div:nth-child(3) > div > div").
		Text()
	if team2Result == "" {
		t.Errorf("team 2's result is not found\n")
	}

	team1Name = strings.TrimSpace(team1Name)
	team2Name = strings.TrimSpace(team2Name)
	team1Result = strings.TrimSpace(team1Result)
	team2Result = strings.TrimSpace(team2Result)

	log.Println(team1Name)
	log.Println(team1Result)
	log.Println(team2Name)
	log.Println(team2Result)

}
