FROM golang:1.10
EXPOSE 8080
WORKDIR /go/src/
COPY . /go/src/github.com/icco/writing
RUN ls -al /go/src/github.com/icco/writing

RUN go build -o ../bin/server ./github.com/icco/writing/server

CMD ["server"]
