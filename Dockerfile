FROM node:12-alpine as webbuilder

COPY ./ /pike 

RUN cd /pike/web \
  && yarn \
  && yarn build \
  && rm -rf node_module

FROM golang:1.14-alpine as builder

COPY --from=webbuilder /pike /pike


RUN apk update \
  && apk add git make \
  && go get -u github.com/gobuffalo/packr/v2/packr2 \
  && cd /pike \
  && make build

FROM alpine

RUN addgroup -g 1000 pike \
  && adduser -u 1000 -G pike -s /bin/sh -D pike \
  && apk add --no-cache ca-certificates

COPY --from=builder /pike/pike /usr/local/bin/pike
COPY --from=builder /pike/entrypoint.sh /home/pike/entrypoint.sh

USER pike

WORKDIR /home/pike

CMD ["pike"]

ENTRYPOINT ["entrypoint.sh"]
