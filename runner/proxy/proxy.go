package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

func testProxy(proxyURL *url.URL, targetURL string) error {
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Request unsuccessfully with status code: %d\n", resp.StatusCode)
	}

	return nil
}

// Return a valid proxies list from the provided ones
func GetValidProxies(proxies []string, targetURL string) ([]string, error) {
	var validProxies []string

	for _, proxy := range proxies {
		var err error

		rawProxyHttpURL := fmt.Sprintf("http://%s", proxy)
		proxyHttpURL, err := url.Parse(rawProxyHttpURL)
		if err != nil {
			return nil, err
		}
		rawProxyHttpsURL := fmt.Sprintf("https://%s", proxy)
		proxyHttpsURL, err := url.Parse(rawProxyHttpsURL)
		if err != nil {
			return nil, err
		}

		var wg sync.WaitGroup
		wg.Add(2)

		var httpErrChan error
		var httpsErrChan error

		go func() {
			httpErrChan = testProxy(proxyHttpURL, targetURL)
			wg.Done()
		}()

		go func() {
			httpsErrChan = testProxy(proxyHttpsURL, targetURL)
			wg.Done()
		}()

		wg.Wait()

		if httpErrChan == nil {
			validProxies = append(validProxies, rawProxyHttpURL)
			fmt.Println(rawProxyHttpURL)
			continue
		}
		if httpsErrChan == nil {
			validProxies = append(validProxies, rawProxyHttpsURL)
			fmt.Println(rawProxyHttpURL)
			continue
		}
	}

	return validProxies, nil
}
