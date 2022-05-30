BIN = $(GOPATH)/bin
BASE = $(GOPATH)/src/$(PACKAGE)
PKGS = go list ./... | grep -v "^vendor/"

# Tools
## Testing library
GINKGO = $(BIN)/ginkgo
$(BIN)/ginkgo:
	go install github.com/onsi/ginkgo/ginkgo

.PHONY: integrationtest
integrationtest: | $(GINKGO) $(GOOSE)
	go vet ./...
	go fmt ./...
	$(GINKGO) -r test/ -v

.PHONY: test
test: | $(GINKGO) $(GOOSE)
	go vet ./...
	go fmt ./...
	$(GINKGO) -r pkg/validator/ validator_test/

build:
	go fmt ./...
	GO111MODULE=on go build
