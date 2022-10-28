build:
	@go build -o pluxy ./cmd/

test-all:
	@env $(shell cat .env | xargs) go test -cover ./...

fmt:
	@go fmt ./...

run: build
	./pluxy
