FROM alpine

ADD ./pike /

ADD ./config.yml /etc/pike/config.yml

CMD ./pike -c /etc/pike/config.yml
