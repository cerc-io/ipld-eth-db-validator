CONTRACTS_DIR := ./test/contract/contracts
CONTRACTS_OUTPUT_DIR := ./internal/testdata/build

CONTRACTS := GLDToken Test

contracts: $(foreach C,$(CONTRACTS), $(CONTRACTS_OUTPUT_DIR)/$C.bin $(CONTRACTS_OUTPUT_DIR)/$C.abi)
.PHONY: contracts

test: contracts
	go test -p 1 -v ./pkg/...
.PHONY: test

clean:
	rm $(CONTRACTS_OUTPUT_DIR)/*.bin $(CONTRACTS_OUTPUT_DIR)/*.abi
