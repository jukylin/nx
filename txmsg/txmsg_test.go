package txmsg

import (
	"context"
	"os"
	"github.com/jukylin/nx/nxlock"
	mocks2 "github.com/jukylin/nx/nxlock/pkg/mocks"
	"github.com/jukylin/nx/txmsg/domain/entity"
	"github.com/jukylin/nx/txmsg/domain/repo/mocks"
	"github.com/jukylin/nx/txmsg/queue"
	"testing"
	"time"

	"github.com/jukylin/esim/log"
)

var logger log.Logger

func TestMain(m *testing.M) {
	logger = log.NewLogger(
		log.WithDebug(true),
	)

	code := m.Run()

	os.Exit(code)
}

func TestTxMsg_processorTimeWheel(t *testing.T) {
	type args struct {
		item interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{"类型错误", args{"123"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTxMsg(
				WithLogger(logger),
			)
			tm.processorTimeWheel(tt.args.item)
		})
	}
}

func TestTxMsg_cleanMsgTask(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	ctx := context.Background()
	tests := []struct {
		name    string
		args    args
		initNum int
	}{
		{"执行清理任务-没有过期数据", args{ctx}, 0},
		{"执行清理任务-有过期数据", args{ctx}, 2},
	}

	msgInfoRepo := &mocks.MsgInfoRepo{}
	msgInfoRepo.On("DelSendedMsg", ctx, 0).Return(int64(0))
	msgInfoRepo.On("DelSendedMsg", ctx, 2).Return(int64(2))

	nxlockSolution := &mocks2.NxlockSolution{}
	nxlockSolution.On("Close").Return(nil)

	nl := nxlock.NewNxlock(
		nxlock.WithLogger(logger),
		nxlock.WithSolution(nxlockSolution),
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTxMsg(
				WithLogger(logger),
				WithMsgInfoRepo(msgInfoRepo),
			)

			tm.nxlock = nl

			tm.deleteMsgOneTimeNum = tt.initNum
			tm.maxDealNumOneTime = 5

			tm.cleanMsgChan <- struct{}{}

			tm.workingTaskWg.Add(1)
			go tm.cleanMsgTask(tt.args.ctx)
			tm.workingTaskWg.Wait()
			tm.Close()
		})
	}
}

func TestTxMsg_scanMsgTask(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	ctx := context.Background()
	ctx, cnacel := context.WithDeadline(ctx, time.Now().Add(10*time.Millisecond))

	msgInfoRepo := &mocks.MsgInfoRepo{}
	msgInfoRepo.On("GetWaitingMsg", ctx, 0).Return(entity.MsgInfos{})
	msgInfoRepo.On("GetWaitingMsg", ctx, 2).Return(entity.MsgInfos{
		entity.MsgInfo{ID: 100},
		entity.MsgInfo{ID: 101},
	})
	msgInfoRepo.On("UpdateStatusById", ctx, int64(100)).Return(int64(1))
	msgInfoRepo.On("UpdateStatusById", ctx, int64(101)).Return(int64(1))

	nxlockSolution := &mocks2.NxlockSolution{}
	nxlockSolution.On("Close").Return(nil)

	fakeRemoteQueue := &queue.FakeRemoteQueue{}

	nl := nxlock.NewNxlock(
		nxlock.WithLogger(logger),
		nxlock.WithSolution(nxlockSolution),
	)

	tests := []struct {
		name string
		args args
		// 数据初始化数量
		initNum int
		done    bool
	}{
		{"没有未提交数据", args{ctx}, 0, false},
		{"有提交数据", args{ctx}, 2, false},
		{"关闭任务", args{ctx}, 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTxMsg(
				WithLogger(logger),
				WithMsgInfoRepo(msgInfoRepo),
				WithRemoteQueue(fakeRemoteQueue),
			)

			tm.nxlock = nl
			tm.limitNum = tt.initNum
			tm.maxDealNumOneTime = 5

			if tt.done == false {
				tm.scanMsgChan <- struct{}{}
				tm.workingTaskWg.Add(1)
			} else {
				cnacel()
			}
			go tm.scanMsgTask(tt.args.ctx)

			if tt.done == false {
				tm.workingTaskWg.Wait()
			} else {
				cnacel()
			}

			tm.Close()
		})
	}
}

func TestTxMsg_putTotimeWheel(t *testing.T) {
	type args struct {
		ctx context.Context
		msg entity.Msg
	}

	ctx := context.Background()

	msg1 := entity.Msg{}
	msg1.ID = 1
	msg1.HaveDealedTimes = 6

	msg2 := entity.Msg{}
	msg2.ID = 2
	msg2.HaveDealedTimes = 5

	msg3 := entity.Msg{}
	msg3.ID = 3
	msg3.HaveDealedTimes = 10

	localQueue := &queue.FakeLocalQueue{}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"超出处理上限", args{ctx, msg1}, true},
		{"处理成功", args{ctx, msg2}, false},
		{"超出边界", args{ctx, msg3}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTxMsg(
				WithLogger(logger),
			)

			tm.timeWheel = localQueue
			if err := tm.putTotimeWheel(tt.args.ctx, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("TxMsg.putTotimeWheel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTxMsg_processor(t *testing.T) {
	type args struct {
		item interface{}
	}

	ctx := context.Background()

	msg1 := entity.NewMsg(1, "")
	msg1.HaveDealedTimes = 6

	msg2 := entity.NewMsg(2, "")
	msg2.HaveDealedTimes = 0

	msg3 := entity.NewMsg(3, "")
	msg3.HaveDealedTimes = 0

	msgInfoRepo := &mocks.MsgInfoRepo{}
	msgInfoRepo.On("FindById", ctx, int64(2)).Return(entity.MsgInfo{})
	msgInfoRepo.On("FindById", ctx, int64(3)).Return(entity.MsgInfo{ID:1})
	msgInfoRepo.On("UpdateStatusById", ctx, int64(3)).Return(int64(1))

	localQueue := &queue.FakeLocalQueue{}

	remoteQueue := &queue.FakeRemoteQueue{}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"超出处理次数", args{msg1}, true},
		{"数据不存在,放入时间轮", args{msg2}, true},
		{"投递成功", args{msg3}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTxMsg(
				WithLogger(logger),
				WithTimeWheel(localQueue),
				WithMsgInfoRepo(msgInfoRepo),
				WithRemoteQueue(remoteQueue),
			)
			if err := tm.processor(tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("TxMsg.processor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
