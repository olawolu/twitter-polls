package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/mgo.v2"
)

var fatalErr error
var counts map[string]int // hold the vote counts
var countsLock sync.Mutex

func main() {
	const updateDuration = 1 * time.Second

	defer func() {
		if fatalErr != nil {
			os.Exit(1)
		}
	}()
	// pollData := connectDB()
	log.Println("Connecting to database...")
	db, err := mgo.Dial("localhost")
	if err != nil {
		fatal(err)
		return
	}
	defer func() {
		log.Println("Closing database connection...")
		db.Close()
	}()
	pollData := db.DB("ballots").C("polls")
	q := consume()
	ticker := time.NewTicker(updateDuration)
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		select {
		case <-ticker.C:
			doCount(&countsLock, &counts, pollData)
		case <-termChan:
			ticker.Stop()
			q.Stop()
		case <-q.StopChan:
			return
		}
	}

}
