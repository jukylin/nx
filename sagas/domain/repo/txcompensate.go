package repo

import (
	"context"

	"fmt"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/sagas/domain/dao"
	"github.com/jukylin/nx/sagas/domain/entity"
	"gorm.io/gorm"
)

type TxcompensateRepo interface {
	FindByTxID(context.Context, uint64) entity.Txcompensate

	Create(ctx context.Context, txgroup *entity.Txcompensate) error

	InsertUpdateFromRecord(context.Context, *gorm.DB, uint64) error

	GetCompensateListByTxID(context.Context, uint64) []entity.Txcompensate

	// 修改状态 补偿成功
	CompensateSuccess(context.Context, int) error

	// 是否还存在未完成补偿事物
	StillHaveUnfinshedCompensationInTransactionGroup(ctx context.Context, txId uint64) (bool, error)
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
	if err != nil && err != gorm.ErrRecordNotFound {
		dtr.logger.Errorc(ctx, err.Error())
	}

	return txcompensate
}

func (dtr *DbTxcompensateRepo) Create(ctx context.Context, txcompensate *entity.Txcompensate) error {
	var err error
	_, err = dtr.txcompensateDao.Create(ctx, txcompensate)
	if err != nil {
		return err
	}

	return nil
}

func (dtr *DbTxcompensateRepo) InsertUpdateFromRecord(ctx context.Context, tx *gorm.DB, txID uint64) error {
	_, err := dtr.txcompensateDao.InsertUpdateFromRecord(ctx, tx, txID)
	return err
}

func (dtr *DbTxcompensateRepo) GetCompensateListByTxID(ctx context.Context, txID uint64) []entity.Txcompensate {
	var txcompensates []entity.Txcompensate
	var err error

	txcompensates, err = dtr.txcompensateDao.List(ctx, "*", 20, "txid = ? and is_deleted = ? and success = 0", txID, 0)
	if err != nil && err != gorm.ErrRecordNotFound {
		dtr.logger.Errorc(ctx, err.Error())
	}

	return txcompensates
}

func (dtr *DbTxcompensateRepo) CompensateSuccess(ctx context.Context, id int) error {
	var err error
	var rowsAffected int64

	rowsAffected, err = dtr.txcompensateDao.Update(ctx, map[string]interface{}{"success": 1}, "id = ?", id)
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return fmt.Errorf("CompensateSuccess id %d 修改数据失败", id)
	}

	return nil
}

func (dtr *DbTxcompensateRepo) StillHaveUnfinshedCompensationInTransactionGroup(ctx context.Context, txID uint64) (bool, error) {
	var err error
	var c int64
	c, err = dtr.txcompensateDao.Count(ctx, "txid = ? and success = 0", txID)
	if err != nil {
		return true, err
	}

	return c > 0, nil
}
