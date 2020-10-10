package repo

import (
	"context"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/saga/domain/dao"
	"github.com/jukylin/nx/saga/domain/entity"
	value_object "github.com/jukylin/nx/saga/domain/value-object"
)

type TxgroupRepo interface {
	FindById(ctx context.Context, id int64) entity.Txgroup

	Create(ctx context.Context, txgroup *entity.Txgroup) error

	SetStateById(ctx context.Context, state int, id int64) error

	// 获取需要补偿的事物组
	GetCompensateList(ctx context.Context, limit int) ([]entity.Txgroup, error)
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

func (dtr *DbTxgroupRepo) FindById(ctx context.Context, id int64) entity.Txgroup {
	var txgroup entity.Txgroup
	var err error

	txgroup, err = dtr.txgroupDao.Find(ctx, "*", "id = ? and is_deleted = ?", id, 0)
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

func (dtr *DbTxgroupRepo) SetStateById(ctx context.Context, state int, id int64) error {
	var err error
	_, err = dtr.txgroupDao.Update(ctx, map[string]interface{}{"state": state}, "id = ?", id)
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
