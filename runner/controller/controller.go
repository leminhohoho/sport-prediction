package controller

import (
	"database/sql"

	"github.com/gocolly/colly"
	"github.com/leminhohoho/sport-prediction/runner/scheduler"
	"github.com/leminhohoho/sport-prediction/runner/scraper"
)

// The controller that control all aspect of the web scraper
// This is web the actual scraper code is written
type Controller struct {
	taskScheduler *scheduler.Scheduler
	db            *sql.DB
	crawler       *colly.Collector
	scrapers      *scraper.Scraper
}

func NewController(
	taskScheduler *scheduler.Scheduler,
	db *sql.DB,
	crawler *colly.Collector,
	scrapers *scraper.Scraper,
) *Controller {
	return &Controller{
		taskScheduler: taskScheduler,
		db:            db,
		crawler:       crawler,
		scrapers:      scrapers,
	}
}

func (c *Controller) Start() error {
	c.taskScheduler.Run()
}
