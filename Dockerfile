FROM golang:1.16-alpine as builder

COPY ./ /pike 

RUN apk update \
  && apk add git make \
  && cd /pike \
  && env \
  && make cp-asset \
  && CGO_ENABLED=0 make build

FROM alpine:3.13

COPY --from=builder /pike/pike /usr/local/bin/pike
COPY --from=builder /pike/entrypoint.sh /usr/local/bin/entrypoint.sh

RUN addgroup -g 1000 pike \
  && adduser -u 1000 -G pike -s /bin/sh -D pike \
  && chmod +x /usr/local/bin/entrypoint.sh \
  && apk add --no-cache ca-certificates


USER pike

WORKDIR /home/pike

CMD ["pike"]

ENTRYPOINT ["entrypoint.sh"]
