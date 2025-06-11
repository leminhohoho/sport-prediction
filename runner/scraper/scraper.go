package scraper

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/leminhohoho/sport-prediction/runner/helpers"
)

type Scraper struct {
	retries int
}

// Create a new scraper, randomness will determinte how unpredictable the scheduler will behave,
// retries specify the maximum number of retries the scraper will make if initial require is failed,
// errorHandler specify the error handler that will be used for this scraper
func NewScraper(retries int) *Scraper {
	return &Scraper{
		retries: retries,
	}
}

func (s *Scraper) initializeBrowserContext(ctx context.Context) (context.Context, context.CancelFunc, error) {
	// Create a context to carry all the options.
	allocCtx, allowCtxCancel := chromedp.NewRemoteAllocator(ctx, "ws://127.0.0.1:9222/devtools/browser")
	newCtx, newCtxCancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	return newCtx, func() { allowCtxCancel(); newCtxCancel() }, nil
}

func (s *Scraper) Scrape(ctx context.Context, url, targetPageReachedSelector string) (string, error) {
	chromeCtx, cancel, err := s.initializeBrowserContext(ctx)
	if err != nil {
		return "", err
	}
	defer cancel()

	var html string

	// Listen for signal send from chrome dev tool and handle it
	chromedp.ListenTarget(chromeCtx, func(event interface{}) {
		switch ev := event.(type) {
		// Handle signals those are sent from every request made
		// Fail every request those are not of document type (HTML) to save bandwidth
		case *fetch.EventRequestPaused:
			go func() {
				if ev.ResourceType == network.ResourceTypeDocument {
					if err = chromedp.Run(chromeCtx, fetch.ContinueRequest(ev.RequestID)); err != nil {
						log.Println(err.Error())
					}
				} else {
					if err = chromedp.Run(chromeCtx, fetch.FailRequest(ev.RequestID, network.ErrorReasonBlockedByClient)); err != nil {
						log.Println(err.Error())
					}
				}
			}()
		case *network.EventResponseReceived:
			if ev.Response.URL == url {
				fetch.Disable().Do(chromeCtx)
			}
		}
	})

	// Initiate a chromium instance and run.
	// NOTE: chromedp.Run() first time called will initiate a chromium instance,
	// therefore it is crucial to not pass a context with timeout for the first call
	// NOTE: Since this is connected to a remote chromium instance, the above note is not
	// applied here
	retries := s.retries

	err = chromedp.Run(
		chromeCtx,
		network.Enable(),
		// Listen for images request and fail it (to save bandwidth)
		fetch.Enable().WithPatterns([]*fetch.RequestPattern{
			{
				URLPattern:   "*",
				RequestStage: fetch.RequestStageRequest,
			},
		}), // List of actions that will be executed sequentially.
		// Disable webdriver detection.
		chromedp.Evaluate(`Object.defineProperty(navigator, 'webdriver', { get: () => false })`, nil),
		// Navigate the the target page.
		chromedp.Navigate(url),

		// --- IMPORTANT: PROCESS OF FETCHING THE TARGET PAGE --- //
		// We will first waiting for the initial page to load (which for the most case is a Cloudfare
		// portal), by waiting for the query selector which query the body to return something.
		// After that, we will wait for a unique query selector that only appear in the target page
		// before fetching the HTML
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.WaitReady(targetPageReachedSelector, chromedp.ByQuery),
		// Introduce a randomn time duration to mimic human response
		chromedp.ActionFunc(func(actionCtx context.Context) error {
			select {
			case <-actionCtx.Done():
				return actionCtx.Err()
			default:
				time.Sleep(helpers.GetRandomTime(time.Second*1, time.Second*3))
				return nil
			}
		}),
		// Extract the HTML content
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
		fetch.Disable(),
	)
	if err != nil {
		if retries > 0 {
			return s.Scrape(ctx, url, targetPageReachedSelector)
		}

		return "", err
	}

	return html, nil
}
