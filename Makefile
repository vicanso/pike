.PHONY: default test test-cover dev

# for dev
dev:
	/usr/local/bin/gowatch 

# for test
test:
	go test -race -cover ./...

test-cover:
	go test -race -coverprofile=test.out ./... && go tool cover --html=test.out

bench:
	go test -bench=. ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

build:
	packr2
	go build -ldflags "-X main.BuildedAt=`date -u +%Y%m%d.%H%M%S` -X main.CommitID=`git rev-parse --short HEAD`" -o pike
	packr2 clean

build-linux:
	GOOS=linux GOARCH=amd64 make build && mv pike pike-linux

build-darwin:
	GOOS=darwin GOARCH=amd64 make build && mv pike pike-darwin

build-win:
	GOOS=windows GOARCH=amd64 make build && mv pike pike-win.exe
	