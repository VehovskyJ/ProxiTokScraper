package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/hashicorp/go-getter"
	"log"
	"net/http"
	"net/url"
	_ "net/url"
	"os"
	"time"
)

const SLEEP_TIME = 500

type media struct {
	videoUrl string
	date     string
}

func main() {
	instance := flag.String("instance", "", "ProxiTok instance domain")
	noWatermark := flag.Bool("no-watermark", false, "Disable downloading with watermark")

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		log.Fatal("Only one username can be specified")
	}
	username := args[0]

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %s", err)
	}

	downloadDir := fmt.Sprintf("%s/tiktok/%s", wd, username)
	err = os.MkdirAll(downloadDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create download directory: %s", err)
	}

	pages, err := getAllPages(*instance, username)
	if err != nil {
		log.Printf("Failed to fetch all pages: %s", err)
		log.Printf("Some contect may be missing or incomplete")
	}

	for _, page := range pages {
		videos, err := getAllVideoUrls(page, *noWatermark)
		if err != nil {
			log.Printf("Failed to fetch page contents: %s", err)
		}

		for _, video := range videos {
			video.videoUrl = fmt.Sprintf("https://%s%s", *instance, video.videoUrl)
			err := downloadVideo(video, username, downloadDir)
			if err != nil {
				log.Printf("Failed to download video: %s", err)
			}
			time.Sleep(SLEEP_TIME * time.Millisecond)
		}
	}
}

func getAllPages(domain string, username string) ([]string, error) {
	var urls []string
	cursor := "0"

	for {
		proxitokUrl := fmt.Sprintf("https://%s/%s/?cursor=%s", domain, username, cursor)

		urls = append(urls, proxitokUrl)
		log.Println(proxitokUrl)

		res, err := http.Get(proxitokUrl)
		if err != nil {
			return urls, err
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			return urls, errors.New(fmt.Sprintf("status code error %s", res.Status))
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return urls, err
		}

		// Finds the next button and extracts the url of the cursor
		nextButton := doc.Find(".buttons > a.button.is-success").First()
		cursorUrl := nextButton.AttrOr("href", "")

		// If the cursor is empty break the loop
		if cursorUrl == "" {
			break
		}

		u, err := url.Parse(cursorUrl)
		if err != nil {
			return urls, err
		}

		// Extracts the new cursor form the cursor url
		newCursor := u.Query().Get("cursor")

		/*
		 * If the new cursor is equal to zero or to the old cursor break the loop
		 * Comparing the new cursor to the old one is important since some instances
		 * of proxitok are buggy and return the old cursor which causes infinite loop
		 */
		if newCursor == "0" || newCursor == cursor {
			break
		}

		cursor = newCursor
		time.Sleep(SLEEP_TIME * time.Millisecond)
	}

	return urls, nil
}

func getAllVideoUrls(page string, noWatermark bool) ([]media, error) {
	res, err := http.Get(page)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("status code error %s", res.Status))
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	var videos []media
	doc.Find("div.media-content").Each(func(i int, selection *goquery.Selection) {
		var video media

		dateString, _ := selection.Find("small[title]").Attr("title")
		date, err := time.Parse("Jan 02, 2006 15:04:05 UTC", dateString)
		if err != nil {
			return
		}
		video.date = date.Format("20060102_150405")

		if noWatermark {
			video.videoUrl, _ = selection.Find("a.button.is-success:contains('No watermark')").Attr("href")
		} else {
			video.videoUrl, _ = selection.Find("a.button.is-info:contains('Watermark')").Attr("href")
		}

		videos = append(videos, video)
	})

	return videos, nil
}

func downloadVideo(video media, username string, outputDirectory string) error {
	log.Printf("Downlaoding video %s_%s.mp4", username, video.date)
	destination := fmt.Sprintf("%s/%s_%s.mp4", outputDirectory, username, video.date)

	if _, err := os.Stat(destination); os.IsNotExist(err) {
		client := &getter.Client{
			Src:  video.videoUrl,
			Dst:  destination,
			Mode: getter.ClientModeFile,
		}

		return client.Get()
	}

	log.Println("File already exists")
	return nil
}
