package transactionmanager

import (
	"context"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/saga/domain/dao"
)

type BaseProcessor interface {
	ExeCompensate(ctx context.Context)

	BuildCompensate(ctx context.Context)

	CompensateRecord(ctx context.Context)

	CompensateHook(ctx context.Context)
}

type JdbcProcessor struct {
	logger log.Logger

	txrecordDao dao.TxrecordDao

	txgroupDao dao.TxgroupDao

	txcompensateDao dao.TxcompensateDao
}
