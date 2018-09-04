FROM golang:1.11
ENV GO111MODULE=on
EXPOSE 8080
WORKDIR /go/src/github.com/icco/graphql
COPY . .

RUN go build -o /go/bin/server ./server

CMD ["/go/bin/server"]
