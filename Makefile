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

.PHONY: integrationtest_local
integrationtest_local: | $(GINKGO) $(GOOSE)
	go vet ./...
	go fmt ./...
	./scripts/run_integration_test.sh
