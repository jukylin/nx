package sagas

import (
	"context"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/sagas/domain/entity"
	"github.com/jukylin/nx/sagas/domain/repo"
	"github.com/jukylin/nx/sagas/domain/value-object"
)

type esimTransaction struct {
	context TransactionContext

	logger log.Logger

	txgroup entity.Txgroup

	txgroupRepo repo.TxgroupRepo

	state value_object.State
}

func (et *esimTransaction) AbortTransaction(ctx context.Context) error {
	et.logger.Infoc(ctx, "事物已终止：%d", et.context.TxID())
	et.state.Set(value_object.TranAbort)
	et.txgroup.State = value_object.TranAbort
	return et.txgroupRepo.SetStateBytxID(ctx, value_object.TranAbort, et.txgroup.Txid)
}

func (et *esimTransaction) EndTransaction(ctx context.Context) error {
	et.state.Set(value_object.TranEnd)
	et.txgroup.State = value_object.TranEnd
	et.logger.Infoc(ctx, "事物已结束：%d", et.context.TxID())
	return et.txgroupRepo.SetStateBytxID(ctx, value_object.TranEnd, et.txgroup.Txid)
}

func (et *esimTransaction) Context() TransactionContext {
	return et.context
}
