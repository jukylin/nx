package transactionmanager

import (
	"context"
	"testing"

	"github.com/jukylin/nx/sagas/domain/entity"
	"github.com/jukylin/nx/sagas/domain/repo"
	value_object "github.com/jukylin/nx/sagas/domain/value-object"
	"github.com/stretchr/testify/assert"
)

func Test_backwardCompensate_BuildCompensate(t *testing.T) {
	type args struct {
		ctx     context.Context
		txgroup entity.Txgroup
	}

	ctx := context.Background()

	TxgroupRepo := repo.NewDbTxgroupRepo(logger)
	TxrecordRepo := repo.NewDbTxrecordRepo(logger)
	txcompensateRepo := repo.NewDbTxcompensateRepo(logger)

	TxgroupRepo.Create(ctx, &entity.Txgroup{Txid: 100})
	TxrecordRepo.Create(ctx, &entity.Txrecord{Txid: 100})
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"修改状态开始=》补偿", args{ctx, entity.Txgroup{Txid: 100}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := NewBackwardCompensate(
				WithBcMysqlClient(mysqlClient),
				WithBcTxgroupRepo(TxgroupRepo),
				WithBcTxcompensateRepo(txcompensateRepo),
			)
			if err := bc.BuildCompensate(tt.args.ctx, tt.args.txgroup); (err != nil) != tt.wantErr {
				t.Errorf("backwardCompensate.BuildCompensate() error = %v, wantErr %v", err, tt.wantErr)
			}

			txgroup := TxgroupRepo.FindByTxID(ctx, 100)
			txcompensate := txcompensateRepo.FindByTxID(ctx, 100)
			assert.Equal(t, value_object.TranCompensate, txgroup.State)
			assert.Equal(t, uint64(100), txcompensate.Txid)
		})
	}
}

func Test_backwardCompensate_CompensateHook(t *testing.T) {
	type args struct {
		ctx     context.Context
		txgroup entity.Txgroup
	}

	TxgroupRepo := repo.NewDbTxgroupRepo(logger)
	TxrecordRepo := repo.NewDbTxrecordRepo(logger)
	txcompensateRepo := repo.NewDbTxcompensateRepo(logger)

	ctx := context.Background()

	TxgroupRepo.Create(ctx, &entity.Txgroup{ID: 1, Txid: 1001, State: value_object.TranCompensate})

	txcompensateRepo.Create(ctx, &entity.Txcompensate{ID: 1, Success: 1, Txid: 1001})
	txcompensateRepo.Create(ctx, &entity.Txcompensate{ID: 2, Success: 1, Txid: 1001})

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"补偿完成", args{ctx, entity.Txgroup{ID: 1, Txid: 1001}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := NewBackwardCompensate(
				WithBcTxcompensateRepo(txcompensateRepo),
				WithBcTxgroupRepo(TxgroupRepo),
				WithBcTxrecordRepo(TxrecordRepo),
				WithBcMysqlClient(mysqlClient),
				WithBcLogger(logger),
			)
			if err := bc.CompensateHook(tt.args.ctx, tt.args.txgroup); (err != nil) != tt.wantErr {
				t.Errorf("backwardCompensate.CompensateHook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	txgroup := TxgroupRepo.FindByTxID(ctx, 1001)
	assert.Equal(t, value_object.TranCompensateFinish, txgroup.State)
}
