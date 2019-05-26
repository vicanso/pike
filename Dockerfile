FROM node:10-alpine as webbuilder

RUN apk update \
  && apk add git \
  && git clone --depth=1 https://github.com/vicanso/pike.git /pike \
  && cd /pike/web \
  && npm i \
  && npm run build \
  && rm -rf node_module

FROM golang:1.12-alpine as builder

COPY --from=webbuilder /pike /pike

ENV GOOS linux
ENV GOARCH amd64

RUN apk update \
  && apk add git make g++ bash cmake \
  && go get -u github.com/gobuffalo/packr/v2/packr2 \
  && git clone --depth=1 https://github.com/google/brotli /brotli \
  cd /brotli \
  && ./configure-cmake \
  && make && make install \
  && cd /pike \
  && make test-all \
  && make build
