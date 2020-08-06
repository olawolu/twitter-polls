#!/bin/bash
git init
git add .
git commit -m "Initial commit"

# Add a remote for and fetch the old repo
git remote add -f tweetreader https://github.com/olawolu/tweetreader.git

# Merge the files from tweetreader/master into twitter-poll/master
git merge origin/master --allow-unrelated-histories

# Move tweetreader files and folders into a subdirectory
mkdir tweetreader
dir -exclude tweetreader | %{git mv $_.Name tweetreader}

# commit the move
git commit -m "Move tweetreader files into subdir"

# Repeat for tweetcounter, rest-api, polls-webclient
git remote add -f tweetcounter https://github.com/olawolu/tweetcounter.git
git merge origin/master --allow-unrelated-histories
mkdir tweetcounter
dir -exclude tweetreader,tweetcounter | %{git mv $_.Name tweetcounter}
git commit -m "Move tweetcounter files into subdir"

git remote add -f rest-api https://github.com/olawolu/rest-api.git
git merge origin/master --allow-unrelated-histories
mkdir rest-api
dir -exclude tweetreader,tweetcounter,rest-api | %{git mv $_.Name rest-api}
git commit -m "Move rest-api files into subdir"

git remote add -f polls-web-client https://github.com/olawolu/polls-web-client.git
git merge origin/master --allow-unrelated-histories
mkdir polls-web-client
dir -exclude tweetreader,tweetcounter,rest-api,polls-web-client | %{git mv $_.Name polls-web-client}
git commit -m "Move polls-web-client files into subdir"