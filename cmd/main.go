package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/micastar/file-to-storage-and-share/config"
	"github.com/micastar/file-to-storage-and-share/pkg/db"
	"github.com/micastar/file-to-storage-and-share/pkg/utils"
	"github.com/micastar/file-to-storage-and-share/pkg/web"
)

func main() {

	var clr = make(chan string, 1)

	// redis-cli config set notify-keyspace-events KEA
	// Subscribe a expire event from redis
	var rdb = db.RedisClient(config.REDIS_DB_MAIN)
	pubsub := rdb.PSubscribe(context.Background(), "__keyevent@0__:expired")
	go func() {
		for msg := range pubsub.Channel() {
			log.Println("Expired key:", msg.Payload)
			clr <- msg.Payload
		}
	}()

	// Accept a Event
	// Start Clean Event
	go func() {
		for {
			select {
			case inExpired := <-clr:
				log.Println("Start clean Files")
				utils.CleanupExpiredFiles(inExpired)
			}
		}
	}()

	// Start Web Server
	web.Server()

	// Gracefull Shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-done:
		log.Println("!!!!!!!!!!!Shutdown all!!!!!!!!!!!")
		close(clr)
		web.Shutdown()
	}
	log.Println("Graceful Exit Successfully!")
}
