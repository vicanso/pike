FROM golang:alpine as builder

RUN apk update \
  && apk add git \
  && git clone --depth=1 https://github.com/vicanso/pike.git /go/src/github.com/vicanso/pike \
  && go get -u github.com/golang/dep/cmd/dep \
  && cd /go/src/github.com/vicanso/pike\
  && dep ensure \
  && GOOS=linux go build -o pike

FROM alpine

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/github.com/vicanso/pike/pike /

ADD ./config.yml /etc/pike/config.yml

CMD ./pike -c /etc/pike/config.yml

HEALTHCHECK --interval=10s --timeout=3s \
  CMD ./pike -c /etc/pike/config.yml check || exit 1