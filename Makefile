.PHONY: default test test-cover dev

defalt: dev

# for dev
dev:
	fresh

# for test
test:
	go test -race -cover -v ./...
