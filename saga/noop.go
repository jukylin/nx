package saga

import (
	"context"

	"github.com/jukylin/nx/saga/domain/entity"
)

type noopSaga struct{}

func (es *noopSaga) StartSaga(ctx context.Context, txrecord entity.Txrecord) error {
	return nil
}

func (es *noopSaga) AbortSaga(ctx context.Context) {

}

func (es *noopSaga) EndSaga(ctx context.Context) {

}

func (es *noopSaga) TxID() uint64 {
	return 0
}
