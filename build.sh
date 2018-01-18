# !/bin/sh
GOOS=linux go build

docker build -t vicanso/pike .

rm ./pike
