CONTRACTS_DIR := ./test/contract/contracts
CONTRACTS_OUTPUT_DIR := ./fixture/build

GINKGO := go run github.com/onsi/ginkgo/v2/ginkgo

.PHONY: contracts
contracts: $(CONTRACTS_OUTPUT_DIR)/Test.bin $(CONTRACTS_OUTPUT_DIR)/Test.abi

$(CONTRACTS_OUTPUT_DIR)/%.bin $(CONTRACTS_OUTPUT_DIR)/%.abi: $(CONTRACTS_DIR)/%.sol
	solc --abi --bin -o $(CONTRACTS_OUTPUT_DIR) --overwrite $<

clean:
	rm $(CONTRACTS_OUTPUT_DIR)/*.bin $(CONTRACTS_OUTPUT_DIR)/*.abi

.PHONY: integration-test
integration-test:
	$(GINKGO) -r ./integration -v

.PHONY: test
test:
	# go generate
	$(GINKGO) -r ./validator_test -v

# build:
# 	go fmt ./...
# 	GO111MODULE=on go build
