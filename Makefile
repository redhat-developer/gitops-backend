default: test

.PHONY: build
build:
	go build -v ./...

.PHONY: test
test:
	go test `go list ./... | grep -v test`

.PHONY: gomod_tidy
gomod_tidy:
	go mod tidy

.PHONY: gofmt
gofmt:
	go fmt -x ./...
