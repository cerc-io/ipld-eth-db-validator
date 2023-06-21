package validator

import (
	"fmt"
)

type ChainNotSyncedError struct {
	Head uint64
}

func (e *ChainNotSyncedError) Error() string {
	return fmt.Sprintf("chain not synced (current head: %d)", e.Head)
}
