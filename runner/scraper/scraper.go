package scraper

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"time"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/leminhohoho/sport-prediction/runner/helpers"
)

type Scraper struct {
	randomness int
	retries    int
	proxies    []string
}

// Create a new scraper, randomness will determinte how unpredictable the scheduler will behave,
// retries specify the maximum number of retries the scraper will make if initial require is failed,
// errorHandler specify the error handler that will be used for this scraper
func NewScraper(randomness int, retries int, proxies ...string) *Scraper {
	return &Scraper{
		randomness: randomness,
		retries:    retries,
		proxies:    proxies,
	}
}

func (s *Scraper) initializeBrowserContext(ctx context.Context) (context.Context, context.CancelFunc, error) {
	// Get user data directory for the headless chromium instance.
	// This is important for bypassing Cloudfare bot detection by using a user data
	// directory that has passed the captcha test by cloudfare.
	// Even though we can use the default user data directory, creating a new one is
	// encouraged for flexibility in switching user data directory to even the load
	// and not triggering cloudfare bot detection.
	// This folder need to be granted permission by using chmod before can be used.
	profileDir := os.Getenv("CHROME_USER_DATA_DIR")
	if profileDir == "" {
		return nil, nil, fmt.Errorf("user data dir not specified\n")
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
		// chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.IgnoreCertErrors,
	)

	if len(s.proxies) > 0 {
		proxy := s.proxies[rand.IntN(len(s.proxies))]
		opts = append(opts, chromedp.ProxyServer(proxy))
	}

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
	var buf []byte

	// Listen for signal send from chrome dev tool and handle it
	chromedp.ListenTarget(chromeCtx, func(event interface{}) {
		switch ev := event.(type) {
		// Handle signals those are sent from every request made
		// Fail every request those are not of document type (HTML) to save bandwidth
		case *fetch.EventRequestPaused:
			go func() {
				fmt.Println(ev.Request.URL)
				if ev.ResourceType == network.ResourceTypeDocument {
					_ = chromedp.Run(chromeCtx, fetch.ContinueRequest(ev.RequestID))
				} else {
					fmt.Printf("Fail request %s\n", ev.RequestID)
					_ = chromedp.Run(chromeCtx, fetch.FailRequest(ev.RequestID, network.ErrorReasonBlockedByClient))
				}
			}()
		case *network.EventResponseReceived:
			res := ev.Response
			if ev.Response.URL == url {
				fetch.Disable().Do(chromeCtx)
			}
			fmt.Printf("Received response: URL=%s, Status=%d, Headers=%v",
				res.URL, res.Status, res.Headers)
		}
	})

	// Initiate a chromium instance and run.
	// NOTE: chromedp.Run() first time called will initiate a chromium instance,
	// therefore it is crucial to not pass a context with timeout for the first call
	if err = chromedp.Run(
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
				time.Sleep(helpers.GetRandomTime(time.Second*2, time.Second*5))
				return nil
			}
		}),
		// Extract the HTML content
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
		chromedp.FullScreenshot(&buf, 100),
		fetch.Disable(),
	); err != nil {
		return "", err
	}

	if err := os.WriteFile("fullScreenshot.png", buf, 0o644); err != nil {
		log.Fatal(err)
	}

	return html, nil
}
