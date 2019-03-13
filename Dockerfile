FROM golang:1.12-alpine
ENV GO111MODULE=on
ENV NAT_ENV="production"
EXPOSE 8080
WORKDIR /go/src/github.com/icco/graphql
COPY . .

RUN go build -o /go/bin/server ./server

CMD ["/go/bin/server"]
