package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	u "net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type config struct {
	pages              map[string]int
	baseURL            *u.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages           int
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		println("no website provided\r\nusage: ./go run main.go <URL>")
		os.Exit(1)
	}
	if len(args) > 3 {
		println("too many arguments provided, only one URL is expected")
		os.Exit(1)
	}
	baseURL, err := u.Parse(args[0])
	if err != nil {
		println("error parsing URL:", err.Error())
		os.Exit(1)
	}
	cfg := &config{
		pages:              make(map[string]int),
		baseURL:            baseURL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, 100), // limit concurrency to 10
		wg:                 &sync.WaitGroup{},
		maxPages:           1,
	}
	if len(args) >= 2 {
		maxPages, err := strconv.Atoi(args[2])
		if err != nil {
			println("error parsing max pages:", err.Error())
			os.Exit(1)
		}
		cfg.maxPages = maxPages
	}
	if len(args) == 3 {
		maxConcurrency, err := strconv.Atoi(args[1])
		if err != nil {
			println("error parsing max concurrency:", err.Error())
			os.Exit(1)
		}
		cfg.concurrencyControl = make(chan struct{}, maxConcurrency)
	}
	cfg.wg.Add(1)
	go cfg.crawlPage(args[0])

	cfg.wg.Wait()
	fmt.Println("\nCrawling completed. Found URLs:")
	cfg.printReport(args[0])

}
func getHTML(rawURL string) (string, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		return "", errors.New(fmt.Sprintf("expected Content-Type 'text/html', got '%s'", resp.Header.Get("Content-Type")))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
func (cfg *config) crawlPage(rawCurrentURL string) {
	cfg.concurrencyControl <- struct{}{} // acquire a slot in the concurrency control

	defer func() {
		<-cfg.concurrencyControl // release the slot
		cfg.wg.Done()
	}()

	currentURL, err := u.Parse(rawCurrentURL)
	if err != nil {
		println("error parsing current URL:", err.Error())
		return
	}
	if cfg.baseURL.Host != currentURL.Host {
		println("skipping external link:", currentURL.String())
		return
	}
	normalizedURL := NormalizeURL(rawCurrentURL)
	crawlIsFirst := cfg.addPageVisit(normalizedURL)
	if crawlIsFirst {
		html, err := getHTML(rawCurrentURL)
		if err != nil {
			println("error fetching page:", normalizedURL, " - ", err.Error())
			return
		}
		fmt.Println("crawling page:", normalizedURL)

		urls, err := getURLsFromHTML(html, rawCurrentURL)
		if err != io.EOF {
			println("error getting urls:", err)
			return
		}
		for _, url := range urls {
			if len(cfg.pages) >= cfg.maxPages {
				fmt.Println("max pages reached, stopping crawl")
				return
			}
			cfg.wg.Add(1)
			go cfg.crawlPage(url)
		}
	}

}

func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	if _, exists := cfg.pages[normalizedURL]; exists {
		cfg.pages[normalizedURL]++
		return false
	}
	cfg.pages[normalizedURL] = 1
	return true
}

func (cfg *config) printReport(baseUrl string) {
	fmt.Printf("=============================REPORT for %v\n=============================\r\n", baseUrl)
	// Convert map to slice for sorting
	type urlCount struct {
		url   string
		count int
	}

	var urls []urlCount
	for url, count := range cfg.pages {
		urls = append(urls, urlCount{url, count})
	}

	// Sort by count (desc), then alphabetically
	sort.Slice(urls, func(i, j int) bool {
		if urls[i].count == urls[j].count {
			return urls[i].url < urls[j].url
		}
		return urls[i].count > urls[j].count
	})

	// Print results
	for _, u := range urls {
		fmt.Printf("Found %d internal links to %s", u.count, u.url)
	}
}
