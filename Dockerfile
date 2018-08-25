FROM golang:1.11
EXPOSE 8080
WORKDIR /go/src/
COPY . /go/src/github.com/icco/graphql
RUN ls -al /go/src/github.com/icco/graphql

RUN go build -o ../bin/server ./github.com/icco/graphql/server

CMD ["server"]
