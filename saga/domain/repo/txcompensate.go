package repo

import (
	"context"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/saga/domain/dao"
	"github.com/jukylin/nx/saga/domain/entity"
	"gorm.io/gorm"
)

type TxcompensateRepo interface {
	FindByTxID(context.Context, uint64) entity.Txcompensate

	InsertUpdateFromRecord(context.Context, *gorm.DB, uint64) error

	ListByTxID(context.Context, uint64) []entity.Txcompensate
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

func (dtr *DbTxcompensateRepo) FindByTxID(ctx context.Context, txID uint64) entity.Txcompensate {
	var txcompensate entity.Txcompensate
	var err error

	txcompensate, err = dtr.txcompensateDao.Find(ctx, "*", "txid = ? and is_deleted = ?", txID, 0)
	if err != nil && err != gorm.ErrRecordNotFound{
		dtr.logger.Errorc(ctx, err.Error())
	}

	return txcompensate
}

func (dtr *DbTxcompensateRepo) InsertUpdateFromRecord(ctx context.Context, tx *gorm.DB, txID uint64) error {
	_, err := dtr.txcompensateDao.InsertUpdateFromRecord(ctx, tx, txID)
	return err
}

func (dtr *DbTxcompensateRepo) ListByTxID(ctx context.Context, txID uint64) []entity.Txcompensate {
	var txcompensates []entity.Txcompensate
	var err error

	txcompensates, err = dtr.txcompensateDao.List(ctx, "*", 20, "txid = ? and is_deleted = ? and success = 0", txID, 0)
	if err != nil && err != gorm.ErrRecordNotFound{
		dtr.logger.Errorc(ctx, err.Error())
	}

	return txcompensates
}

