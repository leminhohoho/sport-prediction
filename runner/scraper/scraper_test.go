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
		3,
		// "http://100.106.3.17:8888",
	)
	content, err := scraper.Scrape(
		ctx,
		"https://www.vlr.gg/487861/leviat-n-academy-vs-shinden-challengers-league-2025-latam-south-ace-stage-2-gf",
		"body > header > nav > a.header-logo > img",
	)
	if err != nil {
		t.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	team1Name := doc.Find(`#wrapper > div.col-container > div.col.mod-3 > div.wf-card.match-header > div.match-header-vs > a.match-header-link.wf-link-hover.mod-1 > div > div.wf-title-med`).
		Text()
	team2Name := doc.Find(`#wrapper > div.col-container > div.col.mod-3 > div.wf-card.match-header > div.match-header-vs > a.match-header-link.wf-link-hover.mod-2 > div > div.wf-title-med`).
		Text()
	team1Player1 := doc.Find(`#wrapper > div.col-container > div.col.mod-3 > div:nth-child(6) > div > div.vm-stats-container > div.vm-stats-game.mod-active > div:nth-child(2) > div:nth-child(1) > table > tbody > tr:nth-child(1) > td.mod-player > div > a > div.text-of`).
		Text()
	team1Player2 := doc.Find(`#wrapper > div.col-container > div.col.mod-3 > div:nth-child(6) > div > div.vm-stats-container > div.vm-stats-game.mod-active > div:nth-child(2) > div:nth-child(1) > table > tbody > tr:nth-child(2) > td.mod-player > div > a > div.text-of`).
		Text()
	team1Player3 := doc.Find(`#wrapper > div.col-container > div.col.mod-3 > div:nth-child(6) > div > div.vm-stats-container > div.vm-stats-game.mod-active > div:nth-child(2) > div:nth-child(1) > table > tbody > tr:nth-child(3) > td.mod-player > div > a > div.text-of`).
		Text()
	team1Player4 := doc.Find(`#wrapper > div.col-container > div.col.mod-3 > div:nth-child(6) > div > div.vm-stats-container > div.vm-stats-game.mod-active > div:nth-child(2) > div:nth-child(1) > table > tbody > tr:nth-child(4) > td.mod-player > div > a > div.text-of`).
		Text()
	team1Player5 := doc.Find(`#wrapper > div.col-container > div.col.mod-3 > div:nth-child(6) > div > div.vm-stats-container > div.vm-stats-game.mod-active > div:nth-child(2) > div:nth-child(1) > table > tbody > tr:nth-child(5) > td.mod-player > div > a > div.text-of`).
		Text()

	team2Player1 := doc.Find(`#wrapper > div.col-container > div.col.mod-3 > div:nth-child(6) > div > div.vm-stats-container > div.vm-stats-game.mod-active > div:nth-child(2) > div:nth-child(2) > table > tbody > tr:nth-child(1) > td.mod-player > div > a > div.text-of`).
		Text()
	team2Player2 := doc.Find(`#wrapper > div.col-container > div.col.mod-3 > div:nth-child(6) > div > div.vm-stats-container > div.vm-stats-game.mod-active > div:nth-child(2) > div:nth-child(2) > table > tbody > tr:nth-child(2) > td.mod-player > div > a > div.text-of`).
		Text()
	team2Player3 := doc.Find(`#wrapper > div.col-container > div.col.mod-3 > div:nth-child(6) > div > div.vm-stats-container > div.vm-stats-game.mod-active > div:nth-child(2) > div:nth-child(2) > table > tbody > tr:nth-child(3) > td.mod-player > div > a > div.text-of`).
		Text()
	team2Player4 := doc.Find(`#wrapper > div.col-container > div.col.mod-3 > div:nth-child(6) > div > div.vm-stats-container > div.vm-stats-game.mod-active > div:nth-child(2) > div:nth-child(2) > table > tbody > tr:nth-child(4) > td.mod-player > div > a > div.text-of`).
		Text()
	team2Player5 := doc.Find(`#wrapper > div.col-container > div.col.mod-3 > div:nth-child(6) > div > div.vm-stats-container > div.vm-stats-game.mod-active > div:nth-child(2) > div:nth-child(2) > table > tbody > tr:nth-child(5) > td.mod-player > div > a > div.text-of`).
		Text()

	fmt.Printf(
		"%s:\n	%s\n	%s\n	%s\n	%s\n	%s\n",
		strings.TrimSpace(team1Name),
		strings.TrimSpace(team1Player1),
		strings.TrimSpace(team1Player2),
		strings.TrimSpace(team1Player3),
		strings.TrimSpace(team1Player4),
		strings.TrimSpace(team1Player5),
	)
	fmt.Printf(
		"%s:\n	%s\n	%s\n	%s\n	%s\n	%s\n",
		strings.TrimSpace(team2Name),
		strings.TrimSpace(team2Player1),
		strings.TrimSpace(team2Player2),
		strings.TrimSpace(team2Player3),
		strings.TrimSpace(team2Player4),
		strings.TrimSpace(team2Player5),
	)
}
