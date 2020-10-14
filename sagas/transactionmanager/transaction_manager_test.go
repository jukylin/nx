package transactionmanager

import (
	"context"
	"testing"
	"time"

	"github.com/jukylin/nx/sagas/domain/entity"
	"github.com/jukylin/nx/sagas/domain/repo"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/nx/sagas/domain/value-object"
	"github.com/jukylin/nx/sagas/transactionmanager/mocks"
)

func TestTransactionManager_getCompensateToQueue(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	tg := repo.NewDbTxgroupRepo(logger)

	tg.Create(ctx, &entity.Txgroup{
		State:    value_object.TranCompensate,
		Txid:     123212,
		Priority: 2,
	})

	tg.Create(ctx, &entity.Txgroup{
		State:    value_object.TranCompensate,
		Txid:     123213,
		Priority: 2,
	})

	var num int
	tm := NewTransactionManager(
		WithTmLogger(logger),
		WithTmNxLock(nl),
		WithTmTxgroupRepo(tg),
		WithTmTxgroupConsumer(func(item interface{}) {
			num++
		}),
	)

	go tm.getCompensateToQueue(ctx)
	time.Sleep(2000 * time.Millisecond)
	cancel()
	assert.Equal(t, 2, num)
}

func TestTransactionManager_execCompensate(t *testing.T) {
	type args struct {
		item interface{}
	}

	ctx := context.Background()

	txgroup := entity.Txgroup{
		Txid:12312,
	}

	compensate := &mocks.Compensate{}
	compensate.On("ExeCompensate", ctx, txgroup)

	tests := []struct {
		name   string
		args   args
	}{
		{"执行补偿-类型错误", args{entity.Txrecord{}}},
		{"执行补偿-成功", args{txgroup}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTransactionManager(
				WithTmLogger(logger),
				WithTmNxLock(nl),
				WithTmCompensate(compensate),
			)
			tm.execCompensate(tt.args.item)
		})
	}
}
