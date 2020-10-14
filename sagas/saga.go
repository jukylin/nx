package sagas

import (
	"context"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/sagas/domain/entity"
	"github.com/jukylin/nx/sagas/domain/repo"
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
	es.logger.Infoc(ctx, "end saga, txID : %d", es.txID)
}

func (es *esimSaga) TxID() uint64 {
	return es.txID
}
