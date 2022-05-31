BIN = $(GOPATH)/bin
BASE = $(GOPATH)/src/$(PACKAGE)
PKGS = go list ./... | grep -v "^vendor/"

# Tools

.PHONY: integrationtest
integrationtest: | $(GOOSE)
	go vet ./...
	go fmt ./...
	go run github.com/onsi/ginkgo/ginkgo -r test/ -v

.PHONY: test
test: | $(GOOSE)
	go vet ./...
	go fmt ./...
	go run github.com/onsi/ginkgo/ginkgo -r validator_test/ -v

build:
	go fmt ./...
	GO111MODULE=on go build
