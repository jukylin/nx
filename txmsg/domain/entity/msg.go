package entity

import (
	"time"
)

// 事物消息实体
type Msg struct {
	ID int64 `json:"id"`

	Url string `json:"url"`

	CreateTime int64 `json:"create_time"`

	NextExpireTime int64 `json:"next_expire_time"`

	// 处理了多少次
	HaveDealedTimes int `json:"have_dealed_times"`
}

type Msgs []Msg

func (ms Msgs) GetIds() []int64 {
	idsLen := len(ms)
	ids := make([]int64, idsLen)
	if idsLen == 0 {
		return ids
	}

	for k, mi := range ms {
		ids[k] = mi.ID
	}

	return ids
}

func NewMsg(id int64, url string) Msg {
	msg := Msg{}
	msg.ID = id
	msg.Url = url
	msg.CreateTime = time.Now().Unix()
	return msg
}

func (m Msg) IsExpire() bool {
	return m.NextExpireTime > time.Now().Unix()
}
