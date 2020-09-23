package dao

import (
	"context"
	"github.com/jukylin/esim/mysql"
	"gorm.io/gorm"
	"github.com/jukylin/nx/txmsg/domain/entity"
)

type MsgInfoDao struct {
	mysql *mysql.Client
}

func NewMsgInfoDao() *MsgInfoDao {
	dao := &MsgInfoDao{
		mysql: mysql.NewClient(),
	}

	return dao
}

// 主库
func (mid *MsgInfoDao) GetDb(ctx context.Context) *gorm.DB {
	return mid.mysql.GetCtxDb(ctx, "txmsg").Table("msg_info")
}

// 从库
func (mid *MsgInfoDao) GetSlaveDb(ctx context.Context) *gorm.DB {
	return mid.mysql.GetCtxDb(ctx, "txmsg_slave").Table("msg_info")
}

// primary key，error
func (mid *MsgInfoDao) Create(ctx context.Context,
	msgInfo *entity.MsgInfo) (int, error) {
	db := mid.GetDb(ctx).Create(msgInfo)
	if db.Error != nil {
		return int(0), db.Error
	} else {
		return int(msgInfo.ID), nil
	}
}

func (mid *MsgInfoDao) Find(ctx context.Context, squery, wquery interface{},
		args ...interface{}) (entity.MsgInfo, error) {
	var msgInfo entity.MsgInfo
	db := mid.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).Scan(&msgInfo)
	if db.Error != nil {
		return msgInfo, db.Error
	} else {
		return msgInfo, nil
	}
}

func (mid *MsgInfoDao) List(ctx context.Context, limit int, squery, wquery interface{},
	args ...interface{}) (entity.MsgInfos, error) {
	var msgInfos entity.MsgInfos
	db := mid.GetSlaveDb(ctx).Limit(limit).Select(squery).
		Where(wquery, args...).Find(&msgInfos)
	if db.Error != nil {
		return msgInfos, db.Error
	} else {
		return msgInfos, nil
	}
}

func (mid *MsgInfoDao) UpdateWithTx(ctx context.Context, tx *gorm.DB, update map[string]interface{},
		args ...interface{}) (int64, error) {
	db := tx.Table("msg_info").Where("id IN ?", args...).Updates(update)
	return db.RowsAffected, db.Error
}


func (mid *MsgInfoDao) Update(ctx context.Context, update map[string]interface{},
	args ...interface{}) (int64, error) {
	db := mid.GetDb(ctx).Where("id = ?", args...).Updates(update)
	return db.RowsAffected, db.Error
}

func (mid *MsgInfoDao) Delete(ctx context.Context, limit int, wquery interface{},
	args ...interface{}) (int64, error) {
	db := mid.GetDb(ctx).Where(wquery, args...).Limit(limit).Delete(entity.MsgInfo{})
	return db.RowsAffected, db.Error
}