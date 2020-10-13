package repo

import (
	"context"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/saga/domain/dao"
	"github.com/jukylin/nx/saga/domain/entity"
	value_object "github.com/jukylin/nx/saga/domain/value-object"
	"time"
	"gorm.io/gorm"
)

type TxgroupRepo interface {
	FindByTxID(ctx context.Context, txID uint64) entity.Txgroup

	Create(ctx context.Context, txgroup *entity.Txgroup) error

	SetStateBytxID(ctx context.Context, tx *gorm.DB, state int, txID uint64) error

	// 获取需要补偿的事物组
	GetCompensateList(ctx context.Context, limit int) ([]entity.Txgroup, error)

	// 获取未完成的事物
	GetUnfishedTransactionGroup(ctx context.Context, intervals int) ([]entity.Txgroup, error)
}

type DbTxgroupRepo struct {
	logger log.Logger

	txgroupDao *dao.TxgroupDao
}

func NewDbTxgroupRepo(logger log.Logger) TxgroupRepo {
	dtr := &DbTxgroupRepo{
		logger: logger,
	}

	if dtr.txgroupDao == nil {
		dtr.txgroupDao = dao.NewTxgroupDao()
	}

	return dtr
}

func (dtr *DbTxgroupRepo) FindByTxID(ctx context.Context, txID uint64) entity.Txgroup {
	var txgroup entity.Txgroup
	var err error

	txgroup, err = dtr.txgroupDao.Find(ctx, "*", "txid = ? and is_deleted = ?", txID, 0)
	if err != nil {
		dtr.logger.Errorc(ctx, err.Error())
	}

	return txgroup
}

func (dtr *DbTxgroupRepo) Create(ctx context.Context, txgroup *entity.Txgroup) error {
	var err error
	_, err = dtr.txgroupDao.Create(ctx, txgroup)
	if err != nil {
		return err
	}

	return nil
}

func (dtr *DbTxgroupRepo) SetStateBytxID(ctx context.Context, tx *gorm.DB, state int, txID uint64) error {
	var err error
	_, err = dtr.txgroupDao.UpdateByTran(ctx, tx, entity.Txgroup{State:state}, "txid = ?", txID)
	if err != nil {
		return err
	}

	return nil
}

func (dtr *DbTxgroupRepo) GetCompensateList(ctx context.Context, limit int) ([]entity.Txgroup, error) {
	var err error
	var txgroup []entity.Txgroup
	txgroup, err = dtr.txgroupDao.List(ctx, "txid, state, priority", limit, "state = ? and is_deleted = 0", value_object.TranCompensate)
	if err != nil {
		return txgroup, err
	}

	return txgroup, nil
}

func (dtr *DbTxgroupRepo) GetUnfishedTransactionGroup(ctx context.Context, intervals int) ([]entity.Txgroup, error) {
	var err error
	var txgroup []entity.Txgroup
	txgroup, err = dtr.txgroupDao.List(ctx, "txid, state, priority", 1000,
		"state = ? and is_deleted = 0 and create_time < ?", value_object.TranStart,
			time.Now().Add(- time.Duration(intervals) * time.Second))
	if err != nil {
		return txgroup, err
	}

	return txgroup, nil
}
