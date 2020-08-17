package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"

	"gopkg.in/mgo.v2"
)

var (
	dbHost = os.Getenv("DBHOST")
	db     *mgo.Session
)

// poll contains the options for a poll object
type poll struct {
	Options []string
}

// connect to the database
func dialdb() error {
	var err error
	log.Printf("dialing mongodb: %s", dbHost)
	db, err = mgo.Dial(dbHost)
	return err
}

// disconenct from the database
func closedb() {
	db.Close()
	log.Println("closed database connection")
}

// loadOptions
func loadOptions() ([]string, error) {
	var options []string
	var p poll

	// query the polls collection in ballots without filter *Find(nil)*
	// and return an iterator capable of going over the returned polls.
	iter := db.DB("ballots").C("polls").Find(nil).Iter()
	// loop over the results and load the options into the options slice
	for iter.Next(&p) {
		options = append(options, p.Options...)
	}
	iter.Close()
	return options, iter.Err()
}

// publsihvotes takes in a votes channel which is a recieve
func publishVotes(votes <-chan tweet) <-chan struct{} {
	stopchan := make(chan struct{}, 1)
	pub, err := nsq.NewProducer("localhost:4150", nsq.NewConfig())
	if err != nil {
		log.Println(err)
	}
	go func() {
		for vote := range votes {
			log.Println(vote)
			b, err := json.Marshal(vote)
			if err != nil {
				log.Println("Marshall error: ", err)
			}
			pub.Publish("votes", b) // publish votes
		}
		log.Println("Publisher: Stopping")
		pub.Stop()
		log.Println("Publisher: Stopped")
		stopchan <- struct{}{}
	}()
	return stopchan
}

func main() {
	var stoplock sync.Mutex // protects stop
	stop := false
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		stoplock.Lock()
		stop = true
		stoplock.Unlock()
		log.Println("Stopping...")
		stopChan <- struct{}{}
		closeConn()
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	if err := dialdb(); err != nil {
		log.Fatalln("failed to dial MongoDB:", err)
	}
	defer closedb()

	// start things
	votes := make(chan tweet) // channel for votes
	publisherStoppedChan := publishVotes(votes)
	twitterStoppedChan := startTwitterStream(stopChan, votes)
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			closeConn()
			stoplock.Lock()
			if stop {
				stoplock.Unlock()
				return
			}
			stoplock.Unlock()
		}
	}()
	<-twitterStoppedChan
	close(votes)
	<-publisherStoppedChan
}
