.PHONY: test

export GO111MODULE = on

# for dev
dev:
	GO_MODE=dev fresh

test:
	GO_MODE=test go test -race -cover ./... 

test-cover:
	GO_MODE=test go test -race -coverprofile=test.out ./... && go tool cover --html=test.out

build-web:
	cd web && yarn build

format:
	cd web && yarn format

build:
	packr2
	go build -ldflags "-X main.BuildedAt=`date -u +%Y%m%d.%H%M%S` -X main.CommitID=`git rev-parse --short HEAD`" -o pike

bench:
	GO_MODE=test go test -bench=. ./...

lint:
	golangci-lint run

clean:
	packr2 clean
