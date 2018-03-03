# !/bin/sh
cd assets && yarn && npm run build && cd ..

packr

GOOS=darwin go build -o pike-darwin

GOOS=windows go build -o pike-win.exe

GOOS=linux go build -o pike
