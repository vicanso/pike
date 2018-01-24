# !/bin/sh
cd assets && npm i && npm run build && cd ..

go-bindata -pkg server -o server/bindata.go assets/dist

GOOS=darwin go build -o pike-darwin

GOOS=windows go build -o pike-win.exe

GOOS=linux go build -o pike

docker build -t vicanso/pike .
