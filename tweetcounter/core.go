package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"gopkg.in/mgo.v2/bson"

	"github.com/nsqio/go-nsq"
	"gopkg.in/mgo.v2"
)

func fatal(e error) {
	fmt.Println(e)
	flag.PrintDefaults()
	fatalErr = e
}

func connectDB() *mgo.Collection {
	log.Println("Connecting to database...")
	db, err := mgo.Dial("localhost")
	if err != nil {
		fatal(err)
		return nil
	}
	defer func() {
		log.Println("Closing database connection...")
		db.Close()
	}()
	return db.DB("ballots").C("polls")
}

// In order to count the votes, the messages are consumed in the votes topic in NSQ
func consume() *nsq.Consumer {
	// var counts map[string]int // hold the vote counts
	// var countsLock sync.Mutex

	log.Println("Connecting to nsq...")
	log.Println("Connecting to nsq...")

	// create a consumer
	q, err := nsq.NewConsumer("votes", "counter", nsq.NewConfig())
	if err != nil {
		fatal(err)
		return nil
	}
	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		countsLock.Lock()         //lock the countsLock mutex when a new vote comes in
		defer countsLock.Unlock() // defer til when the function exits
		// check whether the counts is nil and make a new map
		if counts == nil {
			counts = make(map[string]int)
		}
		vote := string(m.Body)
		counts[vote]++
		return nil
	}))
	if err := q.ConnectToNSQLookupd("localhost:4161"); err != nil {
		fatal(err)
		return nil
	}
	return q
}

// push reults to database
func doCount(countsLock *sync.Mutex, counts *map[string]int, pollData *mgo.Collection) {
	countsLock.Lock()
	defer countsLock.Unlock()
	if len(*counts) == 0 {
		log.Println("No new votes, skippin database update")
		return
	}
	log.Println("Updating databse...")
	log.Println(*counts)
	ok := true
	log.Println("check")
	for option, count := range *counts {
		sel := bson.M{"options": bson.M{"$in": []string{option}}}
		up := bson.M{"$inc": bson.M{"results." + option: count}}
		if _, err := pollData.UpdateAll(sel, up); err != nil {
			log.Println("failed to update:", err)
			ok = false
		}
	}
	if ok {
		log.Println("Finished updating database...")
		*counts = nil // reset counts
	}
}
