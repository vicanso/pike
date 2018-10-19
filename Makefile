.PHONY: default test test-cover dev

defalt: dev

# for dev
dev: export GO_ENV=dev
dev:
	fresh

# for test
test:
	go test -race -cover ./...
