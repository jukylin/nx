package saga

import (
	"fmt"
)

type esimTransactionContext struct {
	txId uint64
}

func (etc esimTransactionContext) String() string {
	return fmt.Sprintf("%016x", etc.txId)
}

func (etc esimTransactionContext) TxID() uint64 {
	return etc.txId
}
