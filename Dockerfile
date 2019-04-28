FROM golang:1.12-alpine as builder

ENV GOOS linux
ENV GOARCH amd64

RUN apk update \
  && apk add git make g++ bash cmake \
  && git clone --depth=1 https://github.com/google/brotli /brotli \
  cd /brotli \
  && ./configure-cmake \
  && make && make install \
  && git clone --depth=1 https://github.com/vicanso/pike.git -b cod /pike \
  && cd /pike \
  && make test-all \
  && make build
