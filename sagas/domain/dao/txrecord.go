package dao

import (
	"context"
	"errors"
	"time"

	"github.com/jukylin/esim/mysql"
	"github.com/jukylin/nx/sagas/domain/entity"
	"gorm.io/gorm"
)

type TxrecordDao struct {
	mysql *mysql.Client
}

func NewTxrecordDao() *TxrecordDao {
	dao := &TxrecordDao{
		mysql: mysql.NewClient(),
	}

	return dao
}

// master
func (td *TxrecordDao) GetDb(ctx context.Context) *gorm.DB {
	return td.mysql.GetCtxDb(ctx, "sagas").Table("txrecord")
}

// slave
func (td *TxrecordDao) GetSlaveDb(ctx context.Context) *gorm.DB {
	return td.mysql.GetCtxDb(ctx, "sagas_slave").Table("txrecord")
}

// primary key，error
func (td *TxrecordDao) Create(ctx context.Context,
	txrecord *entity.Txrecord) (int, error) {
	txrecord.CreateTime = time.Now()
	txrecord.UpdateTime = time.Now()

	db := td.GetDb(ctx).Create(txrecord)
	if db.Error != nil {
		return int(0), db.Error
	} else {
		return int(txrecord.ID), nil
	}
}

// ctx, "name = ?", "test"
func (td *TxrecordDao) Count(ctx context.Context,
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
func (td *TxrecordDao) Find(ctx context.Context, squery,
	wquery interface{}, args ...interface{}) (entity.Txrecord, error) {
	var txrecord entity.Txrecord
	db := td.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).First(&txrecord)
	if db.Error != nil {
		return txrecord, db.Error
	} else {
		return txrecord, nil
	}
}

// ctx, "id,name", "name = ?", "test"
// return a max of 10 pieces of data
func (td *TxrecordDao) List(ctx context.Context, squery,
	wquery interface{}, args ...interface{}) ([]entity.Txrecord, error) {
	txrecords := make([]entity.Txrecord, 0)
	db := td.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).Limit(10).Find(&txrecords)
	if db.Error != nil {
		return txrecords, db.Error
	} else {
		return txrecords, nil
	}
}

func (td *TxrecordDao) DelById(ctx context.Context,
	id int) (bool, error) {
	var delTxrecord entity.Txrecord

	if delTxrecord.DelKey() == "" {
		return false, errors.New("not found is_del / is_deleted / is_delete")
	}

	delMap := make(map[string]interface{}, 0)
	delMap[delTxrecord.DelKey()] = 1

	delTxrecord.ID = id
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
func (td *TxrecordDao) Update(ctx context.Context,
	update map[string]interface{}, query interface{}, args ...interface{}) (int64, error) {

	db := td.GetDb(ctx).Where(query, args).
		Updates(update)
	return db.RowsAffected, db.Error
}
