package saga

import (
	"context"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/saga/domain/entity"
	"github.com/jukylin/nx/saga/domain/repo"
)

type esimTransaction struct {
	context TransactionContext

	logger log.Logger

	txgroup entity.Txgroup

	txgroupRepo repo.TxgroupRepo

	state State
}

func (et *esimTransaction) AbortTransaction(ctx context.Context) error {
	et.logger.Infoc(ctx, "事物已终止：%d", et.context.TxID())
	et.state.Set(1)
	et.txgroup.State = 1
	return et.txgroupRepo.SetStateById(ctx, 1, int64(et.txgroup.ID))
}

func (et *esimTransaction) EndTransaction(ctx context.Context, state int) error {
	et.state.Set(state)
	et.txgroup.State = state
	et.logger.Infoc(ctx, "事物已结束：%d，%s", et.context.TxID(), et.state.String())
	return et.txgroupRepo.SetStateById(ctx, state, int64(et.txgroup.ID))
}

func (et *esimTransaction) Context() TransactionContext {
	return et.context
}