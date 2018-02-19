FROM alpine

RUN apk add --no-cache ca-certificates

ADD ./pike /

ADD ./config.yml /etc/pike/config.yml

CMD ./pike -c /etc/pike/config.yml

HEALTHCHECK --interval=30s --timeout=3s \
  CMD ./pike -c /etc/pike/config.yml check || exit 1