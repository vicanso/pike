.PHONY: test

export GO111MODULE = on

# for dev
dev:
	fresh

test:
	mkdir -p ~/.pike
	cp ./test_config.yml ~/.pike/config.yml
	cp ./director.yml ~/.pike/director.yml
	GO_MODE=test go test -race -cover ./...

test-all:
	mkdir -p ~/.pike
	cp ./test_config.yml ~/.pike/config.yml
	cp ./director.yml ~/.pike/director.yml
	GO_MODE=test go test -race -cover -tags brotli ./...

test-cover:
	GO_MODE=test go test -race -coverprofile=test.out ./... && go tool cover --html=test.out

test-cover-all:
	GO_MODE=test go test -race -tags brotli -coverprofile=test.out ./... && go tool cover --html=test.out

build:
	go build -tags 'brotli netgo' -ldflags "-X main.BuildedAt=`date -u +%Y%m%d.%H%M%S` -X main.CommitID=`git rev-parse --short HEAD`" -o pike

bench:
	go test -bench=. ./...
