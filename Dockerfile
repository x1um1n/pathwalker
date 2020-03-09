FROM golang:1.14.0 as build

WORKDIR /go/src/pathwalker

COPY . .
RUN go build .
RUN go test ./...

EXPOSE 8080 9080

CMD ["/go/src/pathwalker/pathwalker"]
