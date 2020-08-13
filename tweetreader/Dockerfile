FROM scratch

ADD tweetreader ./
ENV TWITTER_KEY="Qre7HEUwgWZMFR9dWk1BuyXDE"
ENV TWITTER_SECRET="pCnYvlpVsWgdcJOwEey6PmivN0rEBnEKbrIGc7ICap0k7OGkj8"
ENV TWITTER_ACCESS_TOKEN="1191664442473402368-vWMHmhQhuYMbFSOkAkRFGkz1ZIMazZ"
ENV TWITTER_ACCESS_SECRET="JZPq00fY8clSUutgBiIzK5bIn2B6wQnzDAqdT14xdNJto"
ENV DBHOST=127.0.0.1
ENTRYPOINT [ "/tweetreader" ]
# WORKDIR /go/src/github.com/olawolu/twitter-polls/tweetreader
# COPY . .

# RUN go get -d -v ./...
# RUN go install -v ./...

# RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
# CMD ["./main"]