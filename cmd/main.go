package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/micastar/discord-feed/pkg/bin"
	"github.com/micastar/discord-feed/pkg/global"
	"github.com/micastar/discord-feed/pkg/util"
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

	var list []*bin.Post

	var conId *string

	fs, err := bin.NewFeedStore(os.Getenv("DB_URL"))
	if err != nil {
		log.Printf("Error creating FeedStore: %v\n", err)
	}

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

	getIdInterval := global.GetIdInterval
	idTicker := time.NewTicker(getIdInterval)

	log.Println("Initial Send Interval Time")

	sendInterval := global.DefaultSendInterval
	sendTicker := time.NewTicker(sendInterval)

	for {
		// Stop it until idTicker's value to 0
		<-idTicker.C

		conId, err = fs.ReadDataGId()

		// Get the id from db
		if err != nil {

			log.Printf("Error Cannot Get Data G_id: %v\n", err)

			idTicker.Reset(getIdInterval)
			sendTicker.Reset(sendInterval)

			log.Println("Reset Ticker.")
			log.Println("Waking!!!")
		}
		if conId != nil {
			idTicker.Stop()

			log.Println("!!!!!!!!Ecerything Starting!!!!!!!!!!")
			log.Println("First Read Data with ReadDataWithLimit")

			list, _ = fs.ReadDataWithLimit()

			break
		}
	}

	log.Println("Initial Discrod Data")

	// Initial discord structure to send data to discrod via webhook
	discod := bin.Config{}
	discod.WebhookURL = os.Getenv("WebhookURL")

	// Send data with interval time
	setData := func(i []*bin.Post) {

		log.Println("Initial Discord channel to Send data")

		err := discod.InitialConfig(i)
		if err != nil {
			log.Printf("InitialConfig: %s", err)
		}
	}

	go func() {

		for {
			log.Println("Waiting Ticker Ready!!!")

			select {
			case <-sendTicker.C:

				if len(list) > 0 {

					log.Println("len(list) > 0  Send Data")
					setData(list)
					list = nil

				} else {
					log.Println("len(list) < 0  No Data")

					log.Println("No data to send, pausing the timer...")
					sendTicker.Stop()
					for len(list) == 0 {
						log.Println("Pausing...")
						time.Sleep(global.DefaultPausingInterval)

						log.Println("Check Data with ReadLatestData Again")
						log.Printf("conId: %v\tlist: %v\n", *conId, len(list))

						list, err = fs.ReadLatestData(conId)

						log.Printf("conId: %v\tlist: %v\n", *conId, len(list))
						if err != nil {
							log.Println(err)
							continue
						}

						if len(list) > 0 {
							sendTicker.Reset(sendInterval)
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
		sendTicker.Stop()
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
