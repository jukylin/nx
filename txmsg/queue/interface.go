package queue

import (
	"time"
	"context"
	"github.com/jukylin/nx/txmsg/domain/entity"
)

// 本地队列
type LocalQueue interface {
	// 生产
	Produce(entity.Msg) bool

	// 消费
	Consumer(func(interface{})) bool

	// 丢弃，用于队列满后如何处理
	DroppedItem(interface{})

	Close() error
}

type Message struct {
	Content string `json:"content"`

	Topic string `json:"topic"`

	Tag string `json:"tag"`

	Id int `json:"id"`

	CreateTime time.Time `json:"create_time"`
}

func NewMessage() Message {
	msg := Message{}
	msg.CreateTime = time.Now()
	return msg
}

// 远程队列
type RemoteQueue interface {
	// 生产
	Send(ctx context.Context, msg Message) error
}
