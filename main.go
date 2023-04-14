package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	_ "net/url"
	"os"
	"time"
)

func main() {
	domain := os.Args[1]
	username := os.Args[2]

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %s", err)
	}

	downloadDir := fmt.Sprintf("%s/tiktok/%s", wd, username)
	err = os.MkdirAll(downloadDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create download directory: %s", err)
	}

	pages, err := getAllPages(domain, username)
	if err != nil {
		log.Fatalf("Failed to fetch all pages: %s", err)
	}

	for _, page := range pages {
		err := downloadFromPage(page)
		if err != nil {
			log.Fatalf("Failed to download content: %s", err)
		}
	}
}

func getAllPages(domain string, username string) ([]string, error) {
	var urls []string
	cursor := "0"

	for {
		proxitokUrl := fmt.Sprintf("https://%s/%s/?cursor=%s", domain, username, cursor)
		res, err := http.Get(proxitokUrl)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			log.Printf("Status code error: %s", res.Status)
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, err
		}

		nextButton := doc.Find(".buttons > a.button.is-success").First()
		nextCursor := nextButton.AttrOr("href", "")

		if nextCursor == "" {
			break
		}

		u, err := url.Parse(nextCursor)
		if err != nil {
			return nil, err
		}

		cursor = u.Query().Get("cursor")

		if cursor == "0" {
			break
		}

		urls = append(urls, proxitokUrl)
		log.Println(proxitokUrl)
		time.Sleep(time.Second)
	}

	return urls, nil
}

func getAllVideoUrls(page string) ([]string, error) {
	res, err := http.Get(page)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Println()
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var urls []string
	doc.Find(".has-text-centered video source").Each(func(i int, selection *goquery.Selection) {
		src, exists := selection.Attr("src")
		if exists {
			urls = append(urls, src)
		}
	})

	return urls, nil
}
