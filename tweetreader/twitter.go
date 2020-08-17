package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/go-oauth/oauth"
)

// First we create a connection to Twitter's streaming APIs
// The dial function ensures that conn is first closed and then opens a new conenction
// and keeps the conn variable updated with the current connection.
// IF the connection dies, we redial without worrying about zombie connections.
var (
	conn          net.Conn
	reader        io.ReadCloser
	authClient    *oauth.Client
	creds         *oauth.Credentials
	authSetUpOnce sync.Once
	httpClient    *http.Client
	baseURL       = "https://stream.twitter.com/1.1/statuses/filter.json"
	options []string
)

// tweet structure
type tweet struct {
	CreatedAt string `json:"created_at"`
	Text      string `json:"text"`
	User      struct {
		Name       string `json:"name"`
		ScreenName string `json:"screen_name"`
	} `json:"user"`
}

// Connection is periodically closed and a new one initiated to reload options from the database
//  at regular intervals. The closeConn function handles this by closing the connection
// and also closes io.ReadCloser, which is used to read the body of responses


func dial(ctx context.Context, netw, addr string) (net.Conn, error) {
	if conn != nil {
		conn.Close()
		conn = nil
	}
	netc, err := net.DialTimeout(netw, addr, 10*time.Second)
	if err != nil {
		return nil, err
	}
	conn = netc
	return netc, nil
}

func closeConn() {
	if conn != nil {
		conn.Close()
	}
	if reader != nil {
		reader.Close()
	}
}

func setupTwitterAuth() {
	var ts = make(map[string]string)

	ts["ConsumerKey"] = os.Getenv("TWITTER_KEY")
	ts["ConsumerSecret"] = os.Getenv("TWITTER_SECRET")
	ts["AccessToken"] = os.Getenv("TWITTER_ACCESS_TOKEN")
	ts["AccessSecret"] = os.Getenv("TWITTER_ACCESS_SECRET")

	creds = &oauth.Credentials{
		Token:  ts["AccessToken"],
		Secret: ts["AccessSecret"],
	}
	authClient = &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  ts["ConsumerKey"],
			Secret: ts["ConsumerSecret"],
		},
	}

}

// readFromTwitter takes a send only channel called votes; this is how this function
// will inform the rest of our program that it has noticed a vote on Twitter
// votes chan<- string
func readFromTwitter(votes chan<- tweet) {
	// build request object and query
	req, query, err := buildQuery()
	if err != nil {
		log.Println(err)
	}

	// Pass the query and request object to makeRequest
	resp, err := makeRequest(req, query)
	if err != nil {
		log.Println("making request failed:", err)
		return
	}

	// make a new json.Decoder from the body of the request
	reader := resp.Body
	decoder := json.NewDecoder(reader)

	// keep reading inside an infinite for loop by calling the Decode method
	for {
		// Decode tweet into t
		var t tweet
		if err := decoder.Decode(&t); err != nil {
			break
		}
		// Iterate over all possible options, if the tweet has mentioned it,
		// send it on the votes channel.
		for _, option := range options {
			if strings.Contains(
				strings.ToLower(t.Text),
				strings.ToLower(option),
			) {
				log.Println("vote:", option)
				votes <- t
			}
		}
	}
}

// startTwitterStream takes in a recieve only channel (stopchan) to recieve signals on when the goroutine should stop.
// A send only channel (votes)
func startTwitterStream(stopchan <-chan struct{}, votes chan<- tweet) <-chan struct{} {
	stoppedchan := make(chan struct{}, 1)
	go func() {
		defer func() {
			stoppedchan <- struct{}{}
		}()
		for {
			select {
			case <-stopchan:
				log.Println("Stopping Twitter...")
				return
			default:
				log.Println("Querying Twitter...")
				readFromTwitter(votes)
				log.Println(" (waiting)")
				time.Sleep(10 * time.Second) // wait before reconnecting
			}
		}
	}()
	return stoppedchan
}

// buildQuery creates a request to the url endpoint with a query string
func buildQuery() (req *http.Request, query url.Values, err error) {
	// load options from all the polls data
	options, err = loadOptions()
	log.Println("vote:", options)
	if err != nil {
		log.Println("Failed to load options:")
		return nil, nil, err
	}

	// create a url object
	u, err := url.Parse(baseURL)
	if err != nil {
		log.Println("Failed to parse url:")
		return nil, nil, err
	}
	
	// builld query string
	query = make(url.Values)
	query.Set("track", strings.Join(options, ","))

	// build the request object
	req, err = http.NewRequest("POST", u.String(), strings.NewReader(query.Encode()))
	if err != nil {
		log.Println("Failed to create request object:")
		return nil, nil, err
	}
	return req, query, nil
}

func makeRequest(req *http.Request, params url.Values) (*http.Response, error) {
	// sync.Once is used to ensure initialization code gets run only once
	authSetUpOnce.Do(func() {
		setupTwitterAuth()
		httpClient = &http.Client{
			Transport: &http.Transport{
				DialContext: dial,
			},
		}
	})
	formEnc := params.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(formEnc)))
	authClient.SetAuthorizationHeader(req.Header, creds, "POST", req.URL, params)
	return httpClient.Do(req)
}
