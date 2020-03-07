FROM node:12-alpine as webbuilder

COPY ./ /pike 

RUN cd /pike/web \
  && yarn \
  && yarn build \
  && rm -rf node_module

FROM golang:1.13-alpine as builder

COPY --from=webbuilder /pike /pike


RUN apk update \
  && apk add git make \
  && go get -u github.com/gobuffalo/packr/v2/packr2 \
  && cd /pike \
  && make build