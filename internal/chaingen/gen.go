package chaingen

import (
	"crypto/ecdsa"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
)

type GenContext struct {
	ChainConfig *params.ChainConfig
	GenFuncs    []func(int, *core.BlockGen)
	DB          ethdb.Database

	Keys      map[common.Address]*ecdsa.PrivateKey
	Contracts map[string]*ContractSpec
	Genesis   *types.Block

	block    *core.BlockGen // cache the current block for my methods' use
	deployed map[common.Address]string
}

func NewGenContext(chainConfig *params.ChainConfig, db ethdb.Database) *GenContext {
	return &GenContext{
		ChainConfig: chainConfig,
		DB:          db,
		Keys:        make(map[common.Address]*ecdsa.PrivateKey),
		Contracts:   make(map[string]*ContractSpec),

		deployed: make(map[common.Address]string),
	}
}

func (gen *GenContext) AddFunction(fn func(int, *core.BlockGen)) {
	gen.GenFuncs = append(gen.GenFuncs, fn)
}

func (gen *GenContext) AddOwnedAccount(key *ecdsa.PrivateKey) common.Address {
	addr := crypto.PubkeyToAddress(key.PublicKey)
	gen.Keys[addr] = key
	return addr
}

func (gen *GenContext) AddContract(name string, spec *ContractSpec) {
	gen.Contracts[name] = spec
}

func (gen *GenContext) generate(i int, block *core.BlockGen) {
	gen.block = block
	for _, fn := range gen.GenFuncs {
		fn(i, block)
	}
}

// MakeChain creates a chain of n blocks starting at and including the genesis block.
// the returned hash chain is ordered head->parent.
func (gen *GenContext) MakeChain(n int) ([]*types.Block, []types.Receipts, *core.BlockChain) {
	blocks, receipts := core.GenerateChain(
		gen.ChainConfig, gen.Genesis, ethash.NewFaker(), gen.DB, n, gen.generate,
	)
	chain, err := core.NewBlockChain(gen.DB, nil, nil, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	if err != nil {
		panic(err)
	}
	return append([]*types.Block{gen.Genesis}, blocks...), receipts, chain
}

func (gen *GenContext) CreateSendTx(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error) {
	return gen.createTx(from, &to, amount, params.TxGas, nil)
}

func (gen *GenContext) CreateContractTx(from common.Address, contractName string) (*types.Transaction, error) {
	contract := gen.Contracts[contractName]
	if contract == nil {
		return nil, errors.New("No contract with name " + contractName)
	}
	return gen.createTx(from, nil, big.NewInt(0), 1000000, contract.DeploymentCode)
}

func (gen *GenContext) CreateCallTx(from common.Address, to common.Address, methodName string, args ...interface{}) (*types.Transaction, error) {
	contractName, ok := gen.deployed[to]
	if !ok {
		return nil, errors.New("No contract deployed at address " + to.String())
	}
	contract := gen.Contracts[contractName]
	if contract == nil {
		return nil, errors.New("No contract with name " + contractName)
	}

	packed, err := contract.ABI.Pack(methodName, args...)
	if err != nil {
		panic(err)
	}
	return gen.createTx(from, &to, big.NewInt(0), 100000, packed)
}

func (gen *GenContext) DeployContract(from common.Address, contractName string) (common.Address, error) {
	tx, err := gen.CreateContractTx(from, contractName)
	if err != nil {
		return common.Address{}, err
	}
	addr := crypto.CreateAddress(from, gen.block.TxNonce(from))
	gen.deployed[addr] = contractName
	gen.block.AddTx(tx)
	return addr, nil
}

func (gen *GenContext) createTx(from common.Address, to *common.Address, amount *big.Int, gasLimit uint64, data []byte) (*types.Transaction, error) {
	signer := types.MakeSigner(gen.ChainConfig, gen.block.Number())
	nonce := gen.block.TxNonce(from)
	priv, ok := gen.Keys[from]
	if !ok {
		return nil, errors.New("No private key for sender address" + from.String())
	}

	var tx *types.Transaction
	if gen.ChainConfig.IsLondon(gen.block.Number()) {
		tx = types.NewTx(&types.DynamicFeeTx{
			ChainID:   gen.ChainConfig.ChainID,
			Nonce:     nonce,
			To:        to,
			Gas:       gasLimit,
			GasTipCap: big.NewInt(50),
			GasFeeCap: big.NewInt(1000000000),
			Value:     amount,
			Data:      data,
		})
	} else {
		tx = types.NewTx(&types.LegacyTx{
			Nonce: nonce,
			To:    to,
			Value: amount,
			Gas:   gasLimit,
			Data:  data,
		})
	}
	return types.SignTx(tx, signer, priv)
}

func (gen *GenContext) CreateLondonTx(block *core.BlockGen, addr *common.Address, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	londonTrx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   gen.ChainConfig.ChainID,
		Nonce:     block.TxNonce(*addr),
		GasTipCap: big.NewInt(50),
		GasFeeCap: big.NewInt(1000000000),
		Gas:       21000,
		To:        addr,
		Value:     big.NewInt(1000),
		Data:      []byte{},
	})

	transactionSigner := types.MakeSigner(gen.ChainConfig, block.Number())
	return types.SignTx(londonTrx, transactionSigner, key)
}
