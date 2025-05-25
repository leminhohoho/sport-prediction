package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/leminhohoho/sport-prediction/runner/helpers"
)

type Scraper struct {
	randomness int
	retries    int
	// errorHandler models.ErrorHandler
}

// Create a new scraper, randomness will determinte how unpredictable the scheduler will behave,
// retries specify the maximum number of retries the scraper will make if initial require is failed,
// errorHandler specify the error handler that will be used for this scraper
func NewScraper(randomness int, retries int,

// errorHandler models.ErrorHandler
) *Scraper {
	return &Scraper{
		randomness: randomness,
		retries:    retries,
		// errorHandler: errorHandler,
	}
}

func (s *Scraper) Scrape(ctx context.Context, url, targetPageReachedSelector string) (string, error) {
	// Get user data directory for the headless chromium instance.
	// This is important for bypassing Cloudfare bot detection by using a user data
	// directory that has passed the captcha test by cloudfare.
	// Even though we can use the default user data directory, creating a new one is
	// encouraged for flexibility in switching user data directory to even the load
	// and not triggering cloudfare bot detection.
	// This folder need to be granted permission by using chmod before can be used.
	profileDir := os.Getenv("CHROME_USER_DATA_DIR")
	if profileDir == "" {
		return "", fmt.Errorf("user data dir not specified\n")
	}

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(profileDir),

		// Passing flags for the modifying the chromium instance behavior.
		// This is important for both minimizing resource used by the instance
		// And passing crucial credentials to mimic human.
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		// User agent is crucial for represent and actual browser that is making a request.
		chromedp.UserAgent(
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
		),
		// `AutomationControlled` is a feature that set `navigation.webdriver` to true, which tell
		// the server that this instance is controlled by automation tools, therefore we disable
		// it by using `disable-blinks-features` flag.
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		// Set `disable-dev-shm-usage` to true tell the chromium instance to use /tmp as temporary
		// storage, which is better when using with Docker.
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	// Create a context to carry all the options.
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	newCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	var html string

	// Initiate a chromium instance and run.
	// NOTE: chromedp.Run() first time called will initiate a chromium instance,
	// therefore it is crucial to not pass a context with timeout for the first call
	if err := chromedp.Run(
		newCtx,

		// List of actions that will be executed sequentially.
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
				time.Sleep(helpers.GetRandomTime(time.Second*2, time.Second*5))
				return nil
			}
		}),
		// Extract the HTML content
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
	); err != nil {
		return "", err
	}

	return html, nil
}
