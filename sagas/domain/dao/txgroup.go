package dao

import (
	"context"
	"errors"
	"time"

	"github.com/jukylin/esim/mysql"
	"github.com/jukylin/nx/sagas/domain/entity"
	"gorm.io/gorm"
)

type TxgroupDao struct {
	mysql *mysql.Client
}

func NewTxgroupDao() *TxgroupDao {
	dao := &TxgroupDao{
		mysql: mysql.NewClient(),
	}

	return dao
}

// master
func (td *TxgroupDao) GetDb(ctx context.Context) *gorm.DB {
	return td.mysql.GetCtxDb(ctx, "sagas").Table("txgroup")
}

// slave
func (td *TxgroupDao) GetSlaveDb(ctx context.Context) *gorm.DB {
	return td.mysql.GetCtxDb(ctx, "sagas_slave").Table("txgroup")
}

// primary key，error
func (td *TxgroupDao) Create(ctx context.Context,
	txgroup *entity.Txgroup) (int, error) {
	txgroup.UpdateTime = time.Now()
	txgroup.CreateTime = time.Now()

	db := td.GetDb(ctx).Create(txgroup)
	if db.Error != nil {
		return int(0), db.Error
	} else {
		return int(txgroup.ID), nil
	}
}

// ctx, "name = ?", "test"
func (td *TxgroupDao) Count(ctx context.Context,
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
func (td *TxgroupDao) Find(ctx context.Context, squery,
	wquery interface{}, args ...interface{}) (entity.Txgroup, error) {
	var txgroup entity.Txgroup
	db := td.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).First(&txgroup)
	if db.Error != nil {
		return txgroup, db.Error
	} else {
		return txgroup, nil
	}
}

// ctx, "id,name", "name = ?", "test"
// return a max of 10 pieces of data
func (td *TxgroupDao) List(ctx context.Context, squery string, limit int,
	wquery interface{}, args ...interface{}) ([]entity.Txgroup, error) {
	txgroups := make([]entity.Txgroup, 0)
	db := td.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).Limit(limit).Order("rand()").Find(&txgroups)
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		return txgroups, db.Error
	} else {
		return txgroups, nil
	}
}

func (td *TxgroupDao) DelById(ctx context.Context,
	id int) (bool, error) {
	var delTxgroup entity.Txgroup

	if delTxgroup.DelKey() == "" {
		return false, errors.New("not found is_del / is_deleted / is_delete")
	}

	delMap := make(map[string]interface{}, 0)
	delMap[delTxgroup.DelKey()] = 1

	delTxgroup.ID = id
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
func (td *TxgroupDao) Update(ctx context.Context,
	update map[string]interface{}, query interface{}, args ...interface{}) (int64, error) {
	dbRes := td.GetDb(ctx).Where(query, args).
		Updates(update)
	return dbRes.RowsAffected, dbRes.Error
}

func (td *TxgroupDao) UpdateByTran(ctx context.Context,
	tx *gorm.DB, txgroup entity.Txgroup, query interface{}, args ...interface{}) (int64, error) {
	dbRes := tx.Table("txgroup").Where(query, args).
		Updates(txgroup)
	return dbRes.RowsAffected, dbRes.Error
}
