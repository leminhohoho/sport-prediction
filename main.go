package main

import (
	"context"
	"log"

	"github.com/chromedp/chromedp"
)

func main() {
	// WebSocket URL for the running Chromium instance
	allocatorContext, cancel := chromedp.NewRemoteAllocator(
		context.Background(),
		"ws://127.0.0.1:9222/devtools/browser",
	)
	defer cancel()

	// Create a new context using the remote allocator
	ctx, cancel := chromedp.NewContext(allocatorContext)
	defer cancel()

	// Navigate to a website and capture the page title
	var title string
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.vlr.gg"),
		chromedp.Title(&title),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Page title: %s", title)
}
