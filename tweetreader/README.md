## Requirements
-   go
-   Mongodb
-   nsq

## Run
-   start the nsq environment
>   `~$`    nsqlookupd
-   in a seperate terminal window
>   `~$`    nsqd --lookupd-tcp-address=localhost:4160
-   ensure mongo is installed and the mongo daemon is running
-   run the `mongo` command to open a database and add a test poll
-   run the following code
>   `> `   use ballots\
>   switched to db ballots\
>   `> `    db.polls.insert({\
>   `... `  "title":"Test poll",\
>   `... `  "options":["happy","sad","fail", win"]})
-   exit the mongo shell and run the following
>   nsq_tail --topic="votes" --lookupd-http-address=localhost:4161
-   in a seperate terminal
>   go build -o tweetreader\
>   ./tweetreader
##  Tweetreader is a program that:
-   Loads all polls from a datastore and collect all options from the options array in each document
-   Opens and maintains a connection to Twitter's streaming APIs looking for any mention of the options
-   Figures out which option is mentioned and push that option through to NSQ for each tweet that matches the filter
-   If connection to Twitter is dropped, after a short delay, reconnect and continue
-   Periodically re-query MongoDB for the latest polls and refresh the connection to Twitter to make sure we check for the right options
-   Gracefully stop itself when the user terminates the program by hitting ctrl + c

##  Authorisation with Twitter

To use the streaming API, authentication credentials from twitter is required.

-   Head over to https://apps.twitter.com and create a new app with a unique name
-   Visit the <b>API Keys</b> tab and locate the <b>Your access token</b> section and create a new access token.
-   Refresh the page and notice the two sets of keys and secrets:
    -   an API key and a secret 
    -   an access token and the corresponding secret
-   Set these values as environment variables
    -   TWITTER_KEY
    -   TWITTER_SECRET
    -   TWITTER_ACCESS_TOKEN
    -   TWITTER_ACCESS_SECRET
