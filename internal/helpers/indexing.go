package helpers

import (
	"context"
	"fmt"
	"math/big"

	"github.com/cerc-io/plugeth-statediff"
	"github.com/cerc-io/plugeth-statediff/adapt"
	"github.com/cerc-io/plugeth-statediff/indexer"
	"github.com/cerc-io/plugeth-statediff/indexer/interfaces"
	"github.com/cerc-io/plugeth-statediff/indexer/node"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

func TestStateDiffIndexer(ctx context.Context, chainConfig *params.ChainConfig, genHash common.Hash) (interfaces.StateDiffIndexer, error) {
	testInfo := node.Info{
		GenesisBlock: genHash.String(),
		NetworkID:    "1",
		ID:           "1",
		ClientName:   "geth",
		ChainID:      chainConfig.ChainID.Uint64(),
	}
	_, indexer, err := indexer.NewStateDiffIndexer(ctx, chainConfig, testInfo, TestDBConfig, true)
	return indexer, err
}

type IndexChainParams struct {
	Blocks     []*types.Block
	Receipts   []types.Receipts
	StateCache state.Database

	StateDiffParams statediff.Params
	TotalDifficulty *big.Int
	// Whether to skip indexing state nodes (state_cids, storage_cids)
	SkipStateNodes bool
	// Whether to skip indexing IPLD blocks
	SkipIPLDs bool
}

func IndexChain(indexer interfaces.StateDiffIndexer, params IndexChainParams) error {
	builder := statediff.NewBuilder(adapt.GethStateView(params.StateCache))
	// iterate over the blocks, generating statediff payloads, and transforming the data into Postgres
	for i, block := range params.Blocks {
		var args statediff.Args
		var rcts types.Receipts
		if i == 0 {
			args = statediff.Args{
				OldStateRoot: common.Hash{},
				NewStateRoot: block.Root(),
				BlockNumber:  block.Number(),
				BlockHash:    block.Hash(),
			}
		} else {
			args = statediff.Args{
				OldStateRoot: params.Blocks[i-1].Root(),
				NewStateRoot: block.Root(),
				BlockNumber:  block.Number(),
				BlockHash:    block.Hash(),
			}
			rcts = params.Receipts[i-1]
		}

		diff, err := builder.BuildStateDiffObject(args, params.StateDiffParams)
		if err != nil {
			return fmt.Errorf("failed to build diff (block %d): %w", block.Number(), err)
		}
		tx, err := indexer.PushBlock(block, rcts, params.TotalDifficulty)
		if err != nil {
			return fmt.Errorf("failed to index block (block %d): %w", block.Number(), err)
		}
		defer tx.RollbackOnFailure(err)

		if !params.SkipStateNodes {
			for _, node := range diff.Nodes {
				if err = indexer.PushStateNode(tx, node, block.Hash().String()); err != nil {
					if err != nil {
						return fmt.Errorf("failed to index state node: %w", err)
					}
				}
			}
		}
		if !params.SkipIPLDs {
			for _, ipld := range diff.IPLDs {
				if err := indexer.PushIPLD(tx, ipld); err != nil {
					if err != nil {
						return fmt.Errorf("failed to index IPLD: %w", err)
					}
				}
			}
		}
		if err = tx.Submit(); err != nil {
			return fmt.Errorf("failed to commit diff: %w", err)
		}
	}
	return nil
}
