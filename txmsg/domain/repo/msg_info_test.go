package repo

import (
	"context"
	"reflect"
	"study-go/txmsg/domain/entity"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_msgInfoRepo_FindById(t *testing.T) {
	ctx := context.Background()

	msgInfoRepo := NewDBMsgInfoRepo(
		WithMsgInfoLogger(logger),
	)
	msgInfo := entity.MsgInfo{}
	msgInfo.Content = "test"
	msgInfo.Topic = "topic"
	msgInfo.CreateTime = time.Now()
	msgInfoId := msgInfoRepo.Create(ctx, &msgInfo)
	result := msgInfoRepo.FindById(ctx, int64(msgInfoId))
	assert.Equal(t, result.ID, int64(msgInfoId))
}

func Test_msgInfoRepo_UpdateStatusById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int64
	}

	ctx := context.Background()
	tests := []struct {
		name string
		args args
		want int64
	}{
		{"更新状态", args{ctx, 1}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mir := NewDBMsgInfoRepo(
				WithMsgInfoLogger(logger),
			)
			if got := mir.UpdateStatusById(tt.args.ctx, tt.args.id); got != tt.want {
				t.Errorf("msgInfoRepo.UpdateStatusById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_msgInfoRepo_DelSendedMsg(t *testing.T) {
	type args struct {
		ctx context.Context
		num int
	}

	ctx := context.Background()
	tests := []struct {
		name string
		args args
		want int64
	}{
		{"删除前三天数据", args{ctx, 3}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mir := NewDBMsgInfoRepo(
				WithMsgInfoLogger(logger),
			)

			for i := 0; i < tt.args.num; i++ {
				msgInfo := entity.MsgInfo{}
				msgInfo.Content = "test"
				msgInfo.Topic = "topic"
				msgInfo.Status = 1
				msgInfo.CreateTime = time.Now().Add(-4 * time.Minute * 60 * 24)
				mir.Create(tt.args.ctx, &msgInfo)
			}

			if got := mir.DelSendedMsg(tt.args.ctx, tt.args.num); got != tt.want {
				t.Errorf("msgInfoRepo.DelSendedMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_msgInfoRepo_GetWaitingMsg(t *testing.T) {
	type args struct {
		ctx context.Context
		num int
	}

	ctx := context.Background()
	tests := []struct {
		name   string
		args   args
		count   int
	}{
		{"获取等待处理的数据", args{ctx, 2}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mir := NewDBMsgInfoRepo(
				WithMsgInfoLogger(logger),
			)

			for i := 0; i < tt.args.num; i++ {
				msgInfo := entity.MsgInfo{}
				msgInfo.Content = "test"
				msgInfo.Topic = "topic"
				msgInfo.Status = 0
				msgInfo.CreateTime = time.Now()
				mir.Create(tt.args.ctx, &msgInfo)
			}

			if got := mir.GetWaitingMsg(tt.args.ctx, tt.args.num); !reflect.DeepEqual(len(got), tt.count) {
				t.Errorf("msgInfoRepo.GetWaitingMsg() = %v, want %v", got, tt.count)
			}
		})
	}
}
