package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func checkDeadLink(url string) bool {
	resp, err := http.Head(url)
	if err != nil {
		return false // Treat error as dead link
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 OK
	return resp.StatusCode == http.StatusOK
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

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.DataAtom == atom.A {
			for _, a := range n.Attr {
				if a.Key == "href" && !checkDeadLink(a.Val) {
					fmt.Printf("Dead link found: %s\n", a.Val)
				}
			}
		}
	}
}
