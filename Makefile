.PHONY: test

export GO111MODULE = on

# for dev
dev:
	fresh

test:
	GO_MODE=test CONFIG=etcd://127.0.0.1:2379 go test -race -cover ./...

test-cover:
	GO_MODE=test CONFIG=etcd://127.0.0.1:2379 go test -race -coverprofile=test.out ./... && go tool cover --html=test.out

build:
	packr2
	go build -tags 'brotli netgo' -ldflags "-X main.BuildedAt=`date -u +%Y%m%d.%H%M%S` -X main.CommitID=`git rev-parse --short HEAD`" -o pike

bench:
	GO_MODE=test CONFIG=etcd://127.0.0.1:2379 go test -bench=. ./...
