package scraper

import (
	"context"
	"fmt"
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
		// "http://100.106.3.17:8888",
	)
	content, err := scraper.Scrape(
		ctx,
		"https://www.soccerstats.com/pmatch.asp?league=england&stats=117-19-17-2025",
		"#insidetopdiv > table > tbody > tr > td:nth-child(1) > a > img",
	)
	if err != nil {
		t.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	team1Name := doc.Find(`#content > table:nth-child(10) > tbody > tr > td:nth-child(2) > div > div:nth-child(2) > table:nth-child(2) > tbody > tr:nth-child(2) > td:nth-child(1) > a > font`).
		Text()
	team2Name := doc.Find(`#content > table:nth-child(10) > tbody > tr > td:nth-child(2) > div > div:nth-child(2) > table:nth-child(2) > tbody > tr:nth-child(2) > td:nth-child(5) > a > font`).
		Text()
	team1Result := doc.Find(`#content > table:nth-child(10) > tbody > tr > td:nth-child(2) > div > div:nth-child(2) > table:nth-child(2) > tbody > tr:nth-child(2) > td:nth-child(2) > font > b`).
		Text()
	team2Result := doc.Find(`#content > table:nth-child(10) > tbody > tr > td:nth-child(2) > div > div:nth-child(2) > table:nth-child(2) > tbody > tr:nth-child(2) > td:nth-child(2) > font > b`).
		Text()
	team1BallPossesion := doc.Find(`#content > table:nth-child(10) > tbody > tr > td:nth-child(2) > div > div:nth-child(3) > table:nth-child(6) > tbody > tr > td > table:nth-child(3) > tbody > tr:nth-child(2) > td:nth-child(1) > font > b`).
		Text()
	team2BallPossesion := doc.Find(`#content > table:nth-child(10) > tbody > tr > td:nth-child(2) > div > div:nth-child(3) > table:nth-child(6) > tbody > tr > td > table:nth-child(3) > tbody > tr:nth-child(2) > td:nth-child(1) > font > b`).
		Text()
	team1TimeOfLeading := doc.Find(`#content > table:nth-child(10) > tbody > tr > td:nth-child(2) > div > div:nth-child(3) > table:nth-child(6) > tbody > tr > td > table:nth-child(5) > tbody > tr:nth-child(2) > td:nth-child(1) > font > b`).
		Text()
	team2TimeOfLeading := doc.Find(`#content > table:nth-child(10) > tbody > tr > td:nth-child(2) > div > div:nth-child(3) > table:nth-child(6) > tbody > tr > td > table:nth-child(5) > tbody > tr:nth-child(2) > td:nth-child(3) > font > b`).
		Text()

	fmt.Printf("Result: %s %s %s %s\n", team1Name, team1Result, team2Result, team2Name)
	fmt.Printf("Ball Possesion: %s %s %s %s\n", team1Name, team1BallPossesion, team2BallPossesion, team2Name)
	fmt.Printf("Time of leading: %s %s %s %s\n", team1Name, team1TimeOfLeading, team2TimeOfLeading, team2Name)
}
