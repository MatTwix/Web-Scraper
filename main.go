package main

import (
	"io"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func checkDeadLink(url string, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := http.Head(url)
	if err != nil {
		return // Treat error as dead link
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 OK
	if resp.StatusCode >= 400 && resp.StatusCode <= 500 {
		ch <- ("Dead link found: " + url)
	}
}

func main() {
	url := "https://pkg.go.dev/golang.org/x/net/html"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	ch := make(chan string)

	// Горутина для вывода сообщений из канала
	go func() {
		for msg := range ch {
			println(msg)
		}
	}()

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.DataAtom == atom.A {
			for _, a := range n.Attr {
				if a.Key == "href" {
					wg.Add(1)
					go checkDeadLink(a.Val, ch, wg)
				}
			}
		}
	}

	// Закрываем канал после завершения всех проверок
	wg.Wait()
	close(ch)
}
