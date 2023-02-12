package integration

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
)

type ContractDeployed struct {
	Address         string `json:"address"`
	TransactionHash string `json:"txHash"`
	BlockNumber     uint64 `json:"blockNumber"`
	BlockHash       string `json:"blockHash"`
}

type ContractDestroyed struct {
	BlockNumber uint64 `json:"blockNumber"`
}

type PutResult struct {
	BlockNumber uint64 `json:"blockNumber"`
}

type Tx struct {
	From            string   `json:"from"`
	To              string   `json:"to"`
	Value           *big.Int `json:"value"`
	TransactionHash string   `json:"txHash"`
	BlockNumber     uint64   `json:"blockNumber"`
	BlockHash       string   `json:"blockHash"`
}

const ContractServerUrl = "http://localhost:3000"

// Factory to generate endpoint functions
func MakeGetAndDecodeFunc[R any](format string) func(...interface{}) (*R, error) {
	return func(params ...interface{}) (*R, error) {
		params = append([]interface{}{ContractServerUrl}, params...)
		url := fmt.Sprintf(format, params...)
		res, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("%s: %s", url, res.Status)
		}

		var data R
		decoder := json.NewDecoder(res.Body)
		return &data, decoder.Decode(&data)
	}
}

var (
	SendEth             = MakeGetAndDecodeFunc[Tx]("%s/v1/sendEth?to=%s&value=%s")
	DeployContract      = MakeGetAndDecodeFunc[ContractDeployed]("%s/v1/deployContract")
	DestroyContract     = MakeGetAndDecodeFunc[ContractDestroyed]("%s/v1/destroyContract?addr=%s")
	DeployTestContract  = MakeGetAndDecodeFunc[ContractDeployed]("%s/v1/deployTestContract")
	DestroyTestContract = MakeGetAndDecodeFunc[ContractDestroyed]("%s/v1/destroyTestContract?addr=%s")
	PutTestValue        = MakeGetAndDecodeFunc[PutResult]("%s/v1/putTestValue?addr=%s&value=%d")
)
