package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/joeshaw/envdecode"
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
)

func dial(netw, addr string) (net.Conn, error) {
	if conn != nil {
		conn.Close()
		conn = nil
	}
	netc, err := net.DialTimeout(netw, addr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	conn = netc
	return netc, nil
}

// Connection is periodically closed and a new one initiated to reload options from the database
//  at regular intervals. The closeConn function handles this by closing the connection
// and also closes io.ReadCloser, which is used to read the body of responses

func closeConn() {
	if conn != nil {
		conn.Close()
	}
	if reader != nil {
		reader.Close()
	}
}

func setupTwitterAuth() {
	// A struct type to store the environment variables required to authenticate with twitter.
	var ts struct {
		ConsumerKey    string `env:"TWITTER_KEY,required"`
		ConsumerSecret string `env:"TWITTER_SECRET,required"`
		AccessToken    string `env:"TWITTER_ACCESS_TOKEN,required"`
		AccessSecret   string `env:"TWITTER_ACCESS_SECRET,required"`
	}
	if err := envdecode.Decode(&ts); err != nil {
		log.Fatalln(err)
	}

	creds = &oauth.Credentials{
		Token:  ts.AccessToken,
		Secret: ts.AccessSecret,
	}
	authClient = &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  ts.ConsumerKey,
			Secret: ts.ConsumerSecret,
		},
	}

}

func makeRequest(req *http.Request, params url.Values) (*http.Response, error) {
	// sync.Once is used to ensure initialization code gets run only once
	authSetUpOnce.Do(func() {
		setupTwitterAuth()
		httpClient = &http.Client{
			Transport: &http.Transport{
				Dial: dial,
			},
		}
	})
	formEnc := params.Encode()
	req.Header.Set("Content-Type", "application/x-www-form0 urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(formEnc)))
	req.Header.Set("Authorization", authClient.AuthorizationHeader(creds, "POST", req.URL, params))
	return httpClient.Do(req)
}
