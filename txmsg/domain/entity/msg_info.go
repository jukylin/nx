package entity

import (
	"time"
)

// 事物消息实体
type MsgInfo struct {
	ID int64 `gorm:"column:id;primary_key"`

	// 事物消息
	Content string `gorm:"column:content"`

	// 主题
	Topic string `gorm:"column:topic"`

	// 标签
	Tag string `gorm:"column:tag"`

	// 状态 0 待处理 1 已处理
	Status int `gorm:"column:status"`

	// 延迟多长时间
	Delay int `gorm:"column:delay"`

	CreateTime time.Time `gorm:"column:create_time"`
}

type MsgInfos []MsgInfo

func (mis MsgInfos) GetIds() []int64 {
	idsLen := len(mis)
	ids := make([]int64, idsLen)
	if idsLen == 0 {
		return ids
	}

	for k, mi := range mis {
		ids[k] = mi.ID
	}

	return ids
}

func (mi MsgInfo) TableName() string {
	return "msg_info"
}

func (mi MsgInfo) IsEmpty() bool {
	return mi.ID == 0
}

