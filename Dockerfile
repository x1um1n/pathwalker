FROM golang:1.14.0 as build

RUN apt-get update -q && apt-get install  -q -y libexif12 libexif-dev

WORKDIR /go/src/pathwalker
RUN mkdir temp-images
COPY . .
RUN go build .
RUN go test ./...

EXPOSE 8080 9080

CMD ["/go/src/pathwalker/pathwalker"]
