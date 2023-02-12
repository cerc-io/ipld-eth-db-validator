package chaingen

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/statediff/test_helpers"

	"github.com/cerc-io/ipld-eth-db-validator/v5/fixture"
	"github.com/cerc-io/ipld-eth-db-validator/v5/test/helpers"
)

var (
	bank, acct1, acct2 common.Address
	contractAddr       common.Address
	contractDataRoot   string
	defaultContract    *ContractSpec
)

func init() {
	var err error
	defaultContract, err = ParseContract(fixture.TestContractABI, fixture.TestContractCode)
	if err != nil {
		panic(err)
	}
}

// A GenContext which exactly replicates the chain generator used in existing tests
func DefaultGenContext(chainConfig *params.ChainConfig, db ethdb.Database) *GenContext {
	gen := NewGenContext(chainConfig, db)
	bank = gen.AddOwnedAccount(helpers.TestBankKey)
	acct1 = gen.AddOwnedAccount(helpers.Account1Key)
	acct2 = gen.AddOwnedAccount(helpers.Account2Key)
	gen.AddContract("Test", defaultContract)

	gen.AddFunction(func(i int, block *core.BlockGen) {
		if err := defaultChainGen(gen, i, block); err != nil {
			panic(err)
		}
	})
	gen.Genesis = test_helpers.GenesisBlockForTesting(
		db, bank, helpers.TestBankFunds, big.NewInt(params.InitialBaseFee), params.MaxGasLimit,
	)
	return gen
}

func defaultChainGen(gen *GenContext, i int, block *core.BlockGen) error {
	switch i {
	case 0:
		// In block 1, the test bank sends account #1 some ether.
		tx, err := gen.CreateSendTx(bank, acct1, big.NewInt(10000))
		if err != nil {
			panic(err)
		}
		block.AddTx(tx)
	case 1:
		// In block 2, the test bank sends some more ether to account #1.
		// acct1 passes it on to account #2.
		// acct1 creates a test contract.
		tx1, err := gen.CreateSendTx(bank, acct1, big.NewInt(1000))
		if err != nil {
			panic(err)
		}
		block.AddTx(tx1)
		tx2, err := gen.CreateSendTx(acct1, acct2, big.NewInt(1000))
		if err != nil {
			panic(err)
		}
		block.AddTx(tx2)
		contractAddr, err = gen.DeployContract(acct1, "Test")
		if err != nil {
			panic(err)
		}
	case 2:
		block.SetCoinbase(acct2)
		tx, err := gen.CreateCallTx(bank, contractAddr, "Put", big.NewInt(3))
		if err != nil {
			panic(err)
		}
		block.AddTx(tx)
	case 3:
		block.SetCoinbase(acct2)
		tx, err := gen.CreateCallTx(bank, contractAddr, "Put", big.NewInt(9))
		if err != nil {
			panic(err)
		}
		block.AddTx(tx)
	case 4:
		block.SetCoinbase(acct1)
		tx, err := gen.CreateCallTx(bank, contractAddr, "Put", big.NewInt(0))
		if err != nil {
			panic(err)
		}
		block.AddTx(tx)
	}
	return nil
}
