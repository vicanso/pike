FROM golang:alpine as builder

ADD . /go/src/pike-test

RUN apk update \
  && apk add git \
  && go get -u github.com/golang/dep/cmd/dep \
  && cd /go/src/pike-test \
  && dep ensure \
  && GOOS=linux go build -o server

FROM alpine

COPY --from=builder /go/src/pike-test/server /
