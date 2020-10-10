package repo

import (
	"context"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/saga/domain/dao"
	"github.com/jukylin/nx/saga/domain/entity"
)

type TxcompensateRepo interface {
	FindById(context.Context, int64) entity.Txcompensate
}

type DbTxcompensateRepo struct {
	logger log.Logger

	txcompensateDao *dao.TxcompensateDao
}

func NewDbTxcompensateRepo(logger log.Logger) TxcompensateRepo {
	dtr := &DbTxcompensateRepo{
		logger: logger,
	}

	if dtr.txcompensateDao == nil {
		dtr.txcompensateDao = dao.NewTxcompensateDao()
	}

	return dtr
}

func (dtr *DbTxcompensateRepo) FindById(ctx context.Context, id int64) entity.Txcompensate {
	var txcompensate entity.Txcompensate
	var err error

	txcompensate, err = dtr.txcompensateDao.Find(ctx, "*", "id = ? and is_deleted = ?", id, 0)
	if err != nil {
		dtr.logger.Errorc(ctx, err.Error())
	}

	return txcompensate
}
