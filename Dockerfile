FROM node:10-alpine as webbuilder

RUN apk update \
  && apk add git \
  && git clone --depth=1 https://github.com/vicanso/pike.git /pike \
  && cd /pike/web \
  && yarn \
  && yarn build \
  && rm -rf node_module

FROM golang:1.12-alpine as builder

COPY --from=webbuilder /pike /pike

ENV GOOS linux
ENV GOARCH amd64

RUN apk update \
  && apk add git make g++ bash cmake \
  && go get -u github.com/gobuffalo/packr/v2/packr2 \
  && git clone --depth=1 https://github.com/google/brotli /brotli \
  && cd /brotli \
  && ./configure-cmake \
  && make && make install \
  && cd /pike \
  && make build

FROM alpine

RUN addgroup -g 1000 pike \
  && adduser -u 1000 -G pike -s /bin/sh -D pike \
  && apk add --no-cache ca-certificates

COPY --from=builder /usr/local/lib/libbrotlicommon.so.1 /usr/lib/
COPY --from=builder /usr/local/lib/libbrotlienc.so.1 /usr/lib/
COPY --from=builder /usr/local/lib/libbrotlidec.so.1 /usr/lib/
COPY --from=builder /pike/pike /usr/local/bin/pike

USER pike

WORKDIR /home/pike

CMD ["pike"]
