package main

import (
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
		log.Fatalf("Failed to fetch all pages: %s", err)
	}

	for _, page := range pages {
		videos, err := getAllVideoUrls(page, *noWatermark)
		if err != nil {
			log.Fatalf("Failed to download content: %s", err)
		}

		for _, video := range videos {
			video.videoUrl = fmt.Sprintf("https://%s%s", *instance, video.videoUrl)
			err := downloadVideo(video, username, downloadDir)
			if err != nil {
				log.Printf("Failed to download video: %s", err)
			}
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
		time.Sleep(300 * time.Millisecond)
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
		log.Println()
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
			src, _ := selection.Find("a.button.is-success:contains('No watermark')").Attr("href")
			video.videoUrl = src
		} else {
			src, _ := selection.Find("a.button.is-info:contains('Watermark')").Attr("href")
			video.videoUrl = src
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
