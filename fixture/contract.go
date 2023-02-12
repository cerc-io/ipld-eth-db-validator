package fixture

import _ "embed"

var (
	//go:embed build/Test.abi
	TestContractABI string
	//go:embed build/Test.bin
	TestContractCode string
)
