FROM golang:alpine as builder

RUN apk update \
  && apk add git make g++ bash cmake \
  && git clone --depth=1 https://github.com/vicanso/pike.git /go/src/github.com/vicanso/pike \
  && go get -u github.com/golang/dep/cmd/dep \
  && cd /go/src/github.com/vicanso/pike \
  && dep ensure \
  && cd /go/src/github.com/vicanso/pike/vendor/github.com/google/brotli/ \
  && ./configure-cmake \
  && make && make install \
  && cd /go/src/github.com/vicanso/pike \
  && GOOS=linux GOARCH=amd64 go build -tags netgo -ldflags "-X main.buildAt=`date -u +%Y%m%d.%H%M%S`" -o pike

FROM alpine

RUN apk add --no-cache ca-certificates

COPY --from=builder /usr/local/lib/libbrotlicommon.so.1 /usr/lib/
COPY --from=builder /usr/local/lib/libbrotlienc.so.1 /usr/lib/
COPY --from=builder /usr/local/lib/libbrotlidec.so.1 /usr/lib/
COPY --from=builder /go/src/github.com/vicanso/pike/pike /


ADD ./config.yml /etc/pike/config.yml

CMD ["/pike", "-c", "/etc/pike/config.yml"]

HEALTHCHECK --interval=10s --timeout=3s \
  CMD ./pike -c /etc/pike/config.yml check || exit 1
