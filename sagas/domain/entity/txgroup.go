package entity

import (
	"time"
)

type Txgroup struct {
	CreateTime time.Time `gorm:"column:create_time"`

	ID int `gorm:"column:id;primary_key"`

	IsDeleted int `gorm:"column:is_deleted;default:0"`

	Priority int `gorm:"column:priority"`

	// 0 开始 1 结束 2 补偿 3 补偿完成 4 终止
	State int `gorm:"column:state"`

	Txid uint64 `gorm:"column:txid"`

	UpdateTime time.Time `gorm:"column:update_time"`
}

// delete field
func (t Txgroup) DelKey() string {
	return "is_deleted"
}

func (t Txgroup) IsEmpty() bool {
	return t.ID == 0
}
