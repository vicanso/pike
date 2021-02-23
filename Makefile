.PHONY: default test test-cover dev hooks

# for dev
dev:
	air -c .air.toml	

# for test
test:
	go test -race -cover ./...

test-cover:
	go test -race -coverprofile=test.out ./... && go tool cover --html=test.out

bench:
	go test -benchmem -bench=. ./...

lint:
	golangci-lint run --timeout=2m

tidy:
	go mod tidy

cp-asset:
	rm -rf asset/web && cp -rf web asset/web

build:
	go build -ldflags "-X main.BuildedAt=`date -u +%Y%m%d.%H%M%S` -X main.CommitID=`git rev-parse --short HEAD`" -o pike

build-linux:
	GOOS=linux GOARCH=amd64 make build && mv pike pike-linux

build-darwin:
	GOOS=darwin GOARCH=amd64 make build && mv pike pike-darwin

build-win:
	GOOS=windows GOARCH=amd64 make build && mv pike pike-win.exe
	
hooks:
	cp hooks/* .git/hooks/