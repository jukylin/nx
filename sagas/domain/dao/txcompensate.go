package dao

import (
	"context"
	"errors"
	"time"

	"github.com/jukylin/esim/mysql"
	"github.com/jukylin/nx/sagas/domain/entity"
	"gorm.io/gorm"
)

type TxcompensateDao struct {
	mysql *mysql.Client
}

func NewTxcompensateDao() *TxcompensateDao {
	dao := &TxcompensateDao{
		mysql: mysql.NewClient(),
	}

	return dao
}

// master
func (td *TxcompensateDao) GetDb(ctx context.Context) *gorm.DB {
	return td.mysql.GetCtxDb(ctx, "sagas").Table("txcompensate")
}

// slave
func (td *TxcompensateDao) GetSlaveDb(ctx context.Context) *gorm.DB {
	return td.mysql.GetCtxDb(ctx, "sagas_slave").Table("txcompensate")
}

// primary key，error
func (td *TxcompensateDao) Create(ctx context.Context,
	txcompensate *entity.Txcompensate) (int, error) {
	txcompensate.CreateTime = time.Now()
	txcompensate.UpdateTime = time.Now()

	db := td.GetDb(ctx).Create(txcompensate)
	if db.Error != nil {
		return int(0), db.Error
	} else {
		return int(txcompensate.ID), nil
	}
}

// ctx, "name = ?", "test"
func (td *TxcompensateDao) Count(ctx context.Context,
	query interface{}, args ...interface{}) (int64, error) {
	var count int64
	db := td.GetSlaveDb(ctx).Where(query, args...).Count(&count)
	if db.Error != nil {
		return count, db.Error
	} else {
		return count, nil
	}
}

// ctx, "id,name", "name = ?", "test"
func (td *TxcompensateDao) Find(ctx context.Context, squery,
	wquery interface{}, args ...interface{}) (entity.Txcompensate, error) {
	var txcompensate entity.Txcompensate
	db := td.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).First(&txcompensate)
	if db.Error != nil {
		return txcompensate, db.Error
	} else {
		return txcompensate, nil
	}
}

// ctx, "id,name", "name = ?", "test"
// return a max of 10 pieces of data
func (td *TxcompensateDao) List(ctx context.Context, squery string, limit int,
	wquery interface{}, args ...interface{}) ([]entity.Txcompensate, error) {
	txcompensates := make([]entity.Txcompensate, 0)
	db := td.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).Order("create_time asc").Limit(limit).Find(&txcompensates)
	if db.Error != nil {
		return txcompensates, db.Error
	} else {
		return txcompensates, nil
	}
}

func (td *TxcompensateDao) DelById(ctx context.Context,
	id int) (bool, error) {
	var delTxcompensate entity.Txcompensate

	if delTxcompensate.DelKey() == "" {
		return false, errors.New("not found is_del / is_deleted / is_delete")
	}

	delMap := make(map[string]interface{}, 0)
	delMap[delTxcompensate.DelKey()] = 1

	delTxcompensate.ID = id
	db := td.GetDb(ctx).Where("id = ?", id).
		Updates(delMap)
	if db.Error != nil {
		return false, db.Error
	} else {
		return true, nil
	}
}

// ctx, map[string]interface{}{"name": "hello"}, "name = ?", "test"
// return RowsAffected, error
func (td *TxcompensateDao) Update(ctx context.Context,
	update map[string]interface{}, query interface{}, args ...interface{}) (int64, error) {

	db := td.GetDb(ctx).Where(query, args).
		Updates(update)
	return db.RowsAffected, db.Error
}

func (td *TxcompensateDao) InsertUpdateFromRecord(ctx context.Context, tx *gorm.DB, txID uint64) (int64, error) {
	resDb := tx.Exec("insert into txcompensate(txid, id, success, create_time, update_time, step) " +
		"SELECT txid, id, 0, NOW(), NOW(), step FROM txrecord where txid = ?", txID)
	return resDb.RowsAffected, resDb.Error
}
