package saga

import (
	"context"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/saga/domain/repo"
	"github.com/jukylin/nx/saga/domain/entity"
)

type esimSaga struct {
	txID uint64

	logger log.Logger

	txrecordRepo repo.TxrecordRepo
}

func (es *esimSaga) StartSaga(ctx context.Context, txrecord entity.Txrecord) error {
	es.logger.Infoc(ctx, "start saga, txID : %d", es.txID)
	return es.txrecordRepo.Create(ctx, &txrecord)
}

func (es *esimSaga) AbortSaga(ctx context.Context) {

}

func (es *esimSaga) EndSaga(ctx context.Context) {

}

func (es *esimSaga) TxID() uint64 {
	return es.txID
}
