package integration

import (
	"encoding/json"
	"fmt"
	"net/http"

	ethServerIntegration "github.com/cerc-io/ipld-eth-server/v4/test"
)

type PutResult struct {
	BlockNumber int64 `json:"blockNumber"`
}

const srvUrl = "http://localhost:3000"

func DeployTestContract() (*ethServerIntegration.ContractDeployed, error) {
	ethServerIntegration.DeployContract()
	res, err := http.Get(fmt.Sprintf("%s/v1/deployTestContract", srvUrl))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var contract ethServerIntegration.ContractDeployed
	decoder := json.NewDecoder(res.Body)

	return &contract, decoder.Decode(&contract)
}

func PutTestValue(addr string, index, value int) (*PutResult, error) {
	res, err := http.Get(fmt.Sprintf("%s/v1/putTestValue?addr=%s&index=%d&value=%d", srvUrl, addr, index, value))
	if err != nil {
		return nil, err
	}

	var blockNumber PutResult
	decoder := json.NewDecoder(res.Body)

	return &blockNumber, decoder.Decode(&blockNumber)
}

func DestroyTestContract(addr string) (*ethServerIntegration.ContractDestroyed, error) {
	res, err := http.Get(fmt.Sprintf("%s/v1/destroyTestContract?addr=%s", srvUrl, addr))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data ethServerIntegration.ContractDestroyed
	decoder := json.NewDecoder(res.Body)

	return &data, decoder.Decode(&data)
}
