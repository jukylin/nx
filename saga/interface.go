package saga

import (
	"context"

	"github.com/jukylin/nx/saga/domain/entity"
)

const (
	TranContextHeaderName = "txid"
)

type contextKey struct{}

var activeTxIDKey = contextKey{}

func ContextWithTxID(ctx context.Context, txID uint64) context.Context {
	return context.WithValue(ctx, activeTxIDKey, txID)
}

func TxIDFromContext(ctx context.Context) uint64 {
	val := ctx.Value(activeTxIDKey)
	if txID, ok := val.(uint64); ok {
		return txID
	}
	return 0
}

type Sagas interface {
	// 开启事物.
	StartTransaction(ctx context.Context) (Transaction, error)

	CreateSaga(ctx context.Context, txID uint64) (Saga, error)

	Inject(ctx context.Context, format interface{}, abstractCarrier interface{}) error

	Extract(ctx context.Context, format interface{}, abstractCarrier interface{}) (TransactionContext, error)
}

type Transaction interface {
	// 终止事物，还没有实现.
	AbortTransaction(ctx context.Context) error

	// 事物结束.
	EndTransaction(ctx context.Context, state int) error

	// 上下文
	Context() TransactionContext
}

type Saga interface {
	StartSaga(ctx context.Context, txrecord entity.Txrecord) error

	// 还没有实现.
	AbortSaga(ctx context.Context)

	EndSaga(ctx context.Context)
}

type TransactionContext interface {
	TxID() uint64
}
