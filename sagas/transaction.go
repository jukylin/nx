package sagas

import (
	"context"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/sagas/domain/entity"
	"github.com/jukylin/nx/sagas/domain/repo"
	value_object "github.com/jukylin/nx/sagas/domain/value-object"
	"github.com/opentracing/opentracing-go"
)

type esimTransaction struct {
	context TransactionContext

	logger log.Logger

	txgroup entity.Txgroup

	txgroupRepo repo.TxgroupRepo

	state value_object.State

	span opentracing.Span
}

func (et *esimTransaction) AbortTransaction(ctx context.Context) error {
	et.logger.Infoc(ctx, "事物已终止：%d", et.context.TxID())
	et.state.Set(value_object.TranAbort)
	et.txgroup.State = value_object.TranAbort

	defer func() {
		et.span.SetTag("tran_state", et.state.String())
		et.span.Finish()
	}()

	return et.txgroupRepo.SetStateBytxID(ctx, value_object.TranAbort, et.txgroup.Txid)
}

func (et *esimTransaction) EndTransaction(ctx context.Context) error {
	et.state.Set(value_object.TranEnd)
	et.txgroup.State = value_object.TranEnd
	et.logger.Infoc(ctx, "事物已结束：%d", et.context.TxID())

	defer func() {
		et.span.SetTag("tran_state", et.state.String())
		et.span.Finish()
	}()

	return et.txgroupRepo.SetStateBytxID(ctx, value_object.TranEnd, et.txgroup.Txid)
}

func (et *esimTransaction) Context() TransactionContext {
	return et.context
}
