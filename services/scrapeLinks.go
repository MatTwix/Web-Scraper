package services

import (
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func checkDeadLink(url string, aliveCh chan string, deadCh chan string, wg *sync.WaitGroup, semaphore chan struct{}) error {
	semaphore <- struct{}{}
	defer func() { <-semaphore }()
	defer wg.Done()

	client := &http.Client{}
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoScraper/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 || resp.StatusCode == 405 {
		resp.Body.Close()
		getReq, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		getReq.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoScraper/1.0)")
		getReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		resp, err = client.Do(getReq)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	}

	// Check if the response status code is 200 OK
	if resp.StatusCode >= 400 && resp.StatusCode <= 500 {
		deadCh <- (url + " - Status: " + resp.Status)
	} else {
		aliveCh <- url
	}

	return nil
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

func crawlLinks(link string, base string, aliveCh chan string, deadCh chan string, wg *sync.WaitGroup, visitedMU *sync.Mutex, visitedURLs map[string]bool, semaphore chan struct{}) {
	defer wg.Done()

	resp, err := http.Get(link)
	if err != nil {
		return // Treat error as dead link
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode <= 500 {
		deadCh <- (link + " - Status: " + resp.Status)
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

					visitedMU.Lock()
					if visitedURLs[absURL] {
						visitedMU.Unlock()
						continue
					}
					visitedURLs[absURL] = true
					visitedMU.Unlock()

					if linkURL.Host == baseURL.Host {
						wg.Add(1)
						go crawlLinks(absURL, base, aliveCh, deadCh, wg, visitedMU, visitedURLs, semaphore)
					} else {
						wg.Add(1)
						go checkDeadLink(absURL, aliveCh, deadCh, wg, semaphore)
					}
				}
			}
		}
	}
}

func ScrapeLinks(baseURL string) ([]string, []string, error) {
	var visitedMu sync.Mutex
	var semaphore = make(chan struct{}, 10)

	var visitedURLs = make(map[string]bool)

	var aliveLinks []string
	var deadLinks []string

	wg := &sync.WaitGroup{}
	aliveCh := make(chan string)
	deadCh := make(chan string)

	go func() {
		for {
			select {
			case link := <-aliveCh:
				aliveLinks = append(aliveLinks, link)
			case link := <-deadCh:
				deadLinks = append(deadLinks, link)
			}
		}
	}()

	wg.Add(1)
	crawlLinks(baseURL, baseURL, aliveCh, deadCh, wg, &visitedMu, visitedURLs, semaphore)
	wg.Wait()
	close(aliveCh)
	close(deadCh)
	return aliveLinks, deadLinks, nil
}
