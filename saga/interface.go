package saga

import (
	"context"
	"github.com/jukylin/nx/saga/domain/entity"
)

const (
	TranStart = 0

	TranAbort = 1

	TranSucc = 2

	TranFail = 3

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

type State struct {
	val int
}

func NewState(val int) *State {
	s := &State{}
	s.val = val
	return s
}

func (s *State) Set(val int) {
	s.val = val
}

func (s *State) Value() int {
	return s.val
}

func (s *State) String() string {
	switch s.val {
	case TranStart:
		return "开始"
	case TranAbort:
		return "终止"
	case TranSucc:
		return "成功"
	case TranFail:
		return "失败"
	}

	return "未知状态"
}