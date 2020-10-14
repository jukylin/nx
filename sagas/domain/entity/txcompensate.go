package entity

import (
	"time"
)

type Txcompensate struct {
	CreateTime time.Time `gorm:"column:create_time"`

	ID int `gorm:"column:id;primary_key"`

	IsDeleted int `gorm:"column:is_deleted;default:0"`

	Step int `gorm:"column:step"`

	Success int `gorm:"column:success;default:0"`

	Txid uint64 `gorm:"column:txid"`

	UpdateTime time.Time `gorm:"column:update_time"`
}

// delete field
func (t Txcompensate) DelKey() string {
	return "is_deleted"
}

func (t Txcompensate) IsEmpty() bool {
	return t.ID == 0
}
