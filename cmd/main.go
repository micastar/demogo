package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/micastar/discord-feed/bin"
	"github.com/micastar/discord-feed/global"
	"github.com/micastar/discord-feed/util"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
}

func createAndSaveItems(fs *bin.FeedStore, reader *bin.RSSReader) {
	for _, item := range reader.Feed.Items {
		err := fs.SaveItem(item)
		if err != nil {
			log.Printf("Error saving item: %v\n", err)
		}
	}
}

func initialRssReader() {
	reader := bin.NewRSSReader(
		os.Getenv("URL"),
	)
	err := reader.FetchFeed()
	if err != nil {
		log.Printf("Error fetching Level Up Coding RSS feed: %v\n", err)
	}

	fs, err := bin.NewFeedStore(os.Getenv("DB_URL"))
	if err != nil {
		fmt.Printf("Error creating FeedStore: %v\n", err)
		return
	}

	util.ReverseItem(reader.Feed.Items)

	createAndSaveItems(fs, reader)
}

func main() {

	log.Println("Everything Starting!!!")

	log.Println("Initial Store Interval Time")
	// Pull data from remote (rss) with interval time
	storeInterval := global.DefaultStoreInterval
	storeTicker := time.NewTicker(storeInterval)

	go func() {
		for {
			select {
			case <-storeTicker.C:
				log.Println("Start Get Data From Remote")
				initialRssReader()
			}
		}
	}()

	var conId string

	log.Println("Initial Discrod Data")

	// Initial discord structure to send data to discrod via webhook
	discod := bin.Config{}
	discod.WebhookURL = os.Getenv("WebhookURL")

	fs, err := bin.NewFeedStore(os.Getenv("DB_URL"))
	if err != nil {
		log.Printf("Error creating FeedStore: %v\n", err)
	}

	// Get the id from db
	conId, err = fs.ReadDataGId()
	if err != nil {
		log.Fatalf("Error Get Data G_id: %v\n", err)
	}

	// Send data with interval time
	setData := func(i []*bin.Post) {
		for _, v := range i {
			log.Println("Initial Discord channel to Send data")

			err := discod.InitialConfig(*v)
			if err != nil {
				log.Printf("InitialConfig: %s", err)
			}
		}
	}

	log.Println("Initial Send Interval Time")

	sendInterval := global.DefaultSendInterval
	ticker := time.NewTicker(sendInterval)

	var list []*bin.Post

	log.Println("First Read Data with ReadDataWithLimit")

	go func() {

		list, _ = fs.ReadDataWithLimit()

		for {
			log.Println("Waiting Ticker Ready!!!")

			select {
			case <-ticker.C:
				if len(list) > 0 {

					log.Println("Second Send Data")
					setData(list)
					list = nil
				} else {
					log.Println("No data to send, pausing the timer...")
					ticker.Stop()
					for len(list) == 0 {
						log.Println("Pausing...")
						time.Sleep(global.DefaultPausingInterval)

						log.Println("Check Data with ReadLatestData Again")
						list, _ = fs.ReadLatestData(&conId)

						if len(list) > 0 {
							ticker.Reset(sendInterval)
							log.Println("New data detected, resuming the timer.")
							break
						}
					}
				}
			}
		}
	}()

	exit := func() {
		fs.Close()
		storeTicker.Stop()
		ticker.Stop()
	}

	// Shutdown program
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT)

	select {
	case <-done:
		log.Println("!!! RECEIVED THE SIGTERM EXIT SIGNAL, EXITING... !!!")
		exit()
	}

	log.Println("Graceful Exit Successfully!")
}
