package main

import (
	"fmt"

	"github.com/MatTwix/Web-Scraper/services"
)

func main() {
	var url string
	println("Enter the URL to check for dead links: ")
	fmt.Scanln(&url)
	if url == "" {
		fmt.Println("No URL provided. Exiting.")
		return
	}

	aliveLinks, deadLinks, _ := services.ScrapeLinks(url)
	for _, link := range aliveLinks {
		fmt.Println("Alive Link:", link)
	}
	for _, link := range deadLinks {
		fmt.Println("Dead Link:", link)
	}
}
