package sagas

import (
	"context"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/sagas/domain/entity"
	"github.com/jukylin/nx/sagas/domain/repo"
	"github.com/opentracing/opentracing-go"
)

type esimSaga struct {
	txID uint64

	logger log.Logger

	txrecordRepo repo.TxrecordRepo

	span opentracing.Span

	tracer opentracing.Tracer

}

func (es *esimSaga) StartSaga(ctx context.Context, txrecord entity.Txrecord) error {
	es.logger.Infoc(ctx, "start saga, txID : %d", es.txID)

	parSpan := opentracing.SpanFromContext(ctx)
	span := es.tracer.StartSpan("transaction", opentracing.ChildOf(parSpan.Context()))
	span.SetTag("tran_id", es.txID)
	es.span = span

	defer func() {
		span.SetTag("tran_id", es.txID)
		span.SetTag("record_id", txrecord.ID)
	}()

	return es.txrecordRepo.Create(ctx, &txrecord)
}

func (es *esimSaga) AbortSaga(ctx context.Context) {
	es.span.SetTag("state", "abort")
	es.span.Finish()
	es.logger.Infoc(ctx, "abort saga, txID : %d", es.txID)
}

func (es *esimSaga) EndSaga(ctx context.Context) {
	es.span.SetTag("state", "end")
	es.span.Finish()
	es.logger.Infoc(ctx, "end saga, txID : %d", es.txID)
}

func (es *esimSaga) TxID() uint64 {
	return es.txID
}
