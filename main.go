package main

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var visitedURLs = make(map[string]bool)
var visitedMu sync.Mutex

var semaphore = make(chan struct{}, 10)

func checkDeadLink(url string, ch chan string, wg *sync.WaitGroup) {
	semaphore <- struct{}{}
	defer func() { <-semaphore }()
	defer wg.Done()

	client := &http.Client{}
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoScraper/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 || resp.StatusCode == 405 {
		resp.Body.Close()
		getReq, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return
		}
		getReq.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoScraper/1.0)")
		getReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		resp, err = client.Do(getReq)
		if err != nil {
			return
		}
		defer resp.Body.Close()
	}

	// Check if the response status code is 200 OK
	if resp.StatusCode >= 400 && resp.StatusCode <= 500 {
		ch <- ("Dead link found: " + url + " - Status: " + resp.Status)
	}
}

func absoluteURL(base, href string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	refURL, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	absURL := baseURL.ResolveReference(refURL)
	return absURL.String(), nil
}

func crawlLinks(link string, base string, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := http.Get(link)
	if err != nil {
		return // Treat error as dead link
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode <= 500 {
		ch <- ("Dead link found: " + link + " - Status: " + resp.Status)
		return
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return // Parsing error, treat as dead link
	}

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.DataAtom == atom.A {
			for _, a := range n.Attr {
				if a.Key == "href" {
					// ch <- ("link: " + a.Val)

					absURL, err := absoluteURL(base, a.Val)
					if err != nil {
						continue // Skip invalid URLs
					}

					baseURL, err := url.Parse(base)
					if err != nil {
						continue // Skip if base URL is invalid
					}

					linkURL, err := url.Parse(absURL)
					if err != nil {
						continue // Skip if link URL is invalid
					}

					visitedMu.Lock()
					if visitedURLs[absURL] {
						visitedMu.Unlock()
						continue
					}
					visitedURLs[absURL] = true
					visitedMu.Unlock()

					if linkURL.Host == baseURL.Host {
						wg.Add(1)
						go crawlLinks(absURL, base, ch, wg)
					} else {
						wg.Add(1)
						go checkDeadLink(absURL, ch, wg)
					}
				}
			}
		}
	}
}

func main() {
	var url string
	println("Enter the URL to check for dead links: ")
	fmt.Scanln(&url)
	if url == "" {
		fmt.Println("No URL provided. Exiting.")
		return
	}

	wg := &sync.WaitGroup{}
	ch := make(chan string)

	// Горутина для вывода сообщений из канала
	go func() {
		for msg := range ch {
			println(msg)
		}
	}()

	wg.Add(1)
	crawlLinks(url, url, ch, wg)

	// Закрываем канал после завершения всех проверок
	wg.Wait()
	close(ch)
}
