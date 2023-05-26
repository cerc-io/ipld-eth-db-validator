CONTRACTS_DIR := ./test/contract/contracts
CONTRACTS_OUTPUT_DIR := ./internal/testdata/build

GINKGO := go run github.com/onsi/ginkgo/v2/ginkgo

contracts: $(CONTRACTS_OUTPUT_DIR)/Test.bin $(CONTRACTS_OUTPUT_DIR)/Test.abi
.PHONY: contracts

# Solidity 0.8.20 defaults to the Shanghai fork which includes the PUSH0 opcode
# which our testnet doesn't yet support, so use Paris.
$(CONTRACTS_OUTPUT_DIR)/%.bin $(CONTRACTS_OUTPUT_DIR)/%.abi: $(CONTRACTS_DIR)/%.sol
	solc $< --abi --bin -o $(CONTRACTS_OUTPUT_DIR) --overwrite --evm-version=paris

test: contracts
	$(GINKGO) -v -r ./validator_test
.PHONY: test

clean:
	rm $(CONTRACTS_OUTPUT_DIR)/*.bin $(CONTRACTS_OUTPUT_DIR)/*.abi
