.PHONY: test

export GO111MODULE = on

# for dev
dev:
	fresh

test:
	go test -race -cover ./...

test-all:
	go test -race -cover -tags brotli ./...

test-cover:
	go test -race -coverprofile=test.out ./... && go tool cover --html=test.out

test-cover-all:
	go test -race -tags brotli -coverprofile=test.out ./... && go tool cover --html=test.out

bench:
	go test -bench=. ./...
