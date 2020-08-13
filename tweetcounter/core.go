package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"gopkg.in/mgo.v2/bson"

	"github.com/nsqio/go-nsq"
	"gopkg.in/mgo.v2"
)

var (
	dbHost = os.Getenv("DBHOST")
	db     *mgo.Session
)

type tweet struct {
	CreatedAt string `bson:"created_at"`
	Text      string `bson:"text"`
	User      struct {
		Name       string `bson:"name"`
		ScreenName string `bson:"screen_name"`
	} `bson:"user"`
	// Place            interface{}              `bson:"place"`
	// Urls             []map[string]interface{} `bson:"urls"`
	// Entities         struct  {
	// 	Hashtags []interface{}
	// }
	// ExtendedEntities map[string]interface{}   `bson:"extended_entities"`
}

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

func decodeTweet(b []byte) tweet {
	var buf bytes.Buffer
	var t tweet
	enc := gob.NewDecoder(&buf)
	err := enc.Decode(&t)
	if err != nil {
		log.Fatal("encode error:", err)
	}
	return t
}

// In order to count the votes, the messages are consumed in the votes topic in NSQ
func consume() *nsq.Consumer {
	// var counts map[tweet]int // hold the vote counts
	// var countsLock sync.Mutex
	var t tweet

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
			counts = make(map[tweet]int)
		}

		err := json.Unmarshal(m.Body, &t)
		if err != nil {
			log.Println("Unmarshall error: ", err)
		}
		vote := t
		log.Println(vote)
		// vote := decodeTweet(m.Body)
		counts[vote]++
		return nil
	}))
	if err := q.ConnectToNSQLookupd("localhost:4161"); err != nil {
		fatal(err)
		return nil
	}
	return q
}

// TODO: Update database with tweets
// push reults to database
func doCount(countsLock *sync.Mutex, counts *map[tweet]int, pollData *mgo.Collection) {
	countsLock.Lock()
	defer countsLock.Unlock()
	if len(*counts) == 0 {
		log.Println("No new votes, skippin database update")
		return
	}
	log.Println("Updating database...")
	log.Println(*counts)
	ok := true
	log.Println("check")
	for option, count := range *counts {
		log.Println(option)
		sel := bson.M{"options": bson.M{"$in": []string{option.Text}}}
		up := bson.M{"$inc": bson.M{"results." + option.Text: count}}
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

func doPush(updateLock *sync.Mutex, counts *map[tweet]int, collection *mgo.Collection) {
	updateLock.Lock()
	defer updateLock.Unlock()
	ok := true
	// collection := db.DB("ballots").C("tweets")
	for option := range *counts {
		// b, err := bson.Marshal(option)
		b, err := toBson(option)
		if err != nil {
			log.Println("error converting to bson: ", err)
		}
		if err = collection.Insert(b); err != nil {
			log.Println("Insert Error: ", err)
			ok = false
		}
	}
	if ok {
		log.Println("Finished updating database...")
	}
}

// convert tweet to a bson document
func toBson(v interface{}) (doc *bson.D, err error) {
	b, err := bson.Marshal(v)
	if err != nil {
		return nil, err
	}
	err = bson.Unmarshal(b, &doc)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
