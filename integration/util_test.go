package integration_test

import (
	"sync"
)

type atomicBlockSet struct {
	blocks map[uint64]struct{}
	sync.Mutex
}

func newBlockSet() *atomicBlockSet {
	return &atomicBlockSet{blocks: make(map[uint64]struct{})}
}

func (set *atomicBlockSet) contains(block uint64) bool {
	set.Lock()
	defer set.Unlock()
	for done := range set.blocks {
		if done == block {
			return true
		}
	}
	return false
}

func (set *atomicBlockSet) containsAll(blocks []uint64) bool {
	for _, block := range blocks {
		if !set.contains(block) {
			return false
		}
	}
	return true
}

func (set *atomicBlockSet) add(block uint64) {
	set.Lock()
	defer set.Unlock()
	set.blocks[block] = struct{}{}
}
