package chaingen

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type ContractSpec struct {
	DeploymentCode []byte
	ABI            abi.ABI
}

func ParseContract(abiStr, binStr string) (*ContractSpec, error) {
	println(abiStr)
	parsedABI, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		return nil, err
	}
	data := common.Hex2Bytes(binStr)
	return &ContractSpec{data, parsedABI}, nil
}
