package controller

import (
	"context"
	"fmt"
	"testing"

	"github.com/gocolly/colly"
	"github.com/leminhohoho/sport-prediction/runner/scheduler"
)

func TestController(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskScheduler := scheduler.NewScheduler(3, ctx, true)

	crawler := colly.NewCollector(colly.AllowedDomains("www.vlr.gg"))

	c := NewController(taskScheduler, nil, crawler)

	matches, err := c.ScrapeMatches(ctx)
	if err != nil {
		t.Fatal(err)
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
