package repo

import (
	"context"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/sagas/domain/dao"
	"github.com/jukylin/nx/sagas/domain/entity"
)

type TxrecordRepo interface {
	FindById(context.Context, int64) entity.Txrecord

	Create(ctx context.Context, txgroup *entity.Txrecord) error
}

type DbTxrecordRepo struct {
	logger log.Logger

	txrecordDao *dao.TxrecordDao
}

func NewDbTxrecordRepo(logger log.Logger) TxrecordRepo {
	dtr := &DbTxrecordRepo{
		logger: logger,
	}

	if dtr.txrecordDao == nil {
		dtr.txrecordDao = dao.NewTxrecordDao()
	}

	return dtr
}

func (dtr *DbTxrecordRepo) FindById(ctx context.Context, id int64) entity.Txrecord {
	var txrecord entity.Txrecord
	var err error

	txrecord, err = dtr.txrecordDao.Find(ctx, "*", "id = ? and is_deleted = ?", id, 0)
	if err != nil {
		dtr.logger.Errorc(ctx, err.Error())
	}

	return txrecord
}

func (dtr *DbTxrecordRepo) Create(ctx context.Context, txrecord *entity.Txrecord) error {
	var err error
	_, err = dtr.txrecordDao.Create(ctx, txrecord)
	if err != nil {
		return err
	}

	return nil
}
