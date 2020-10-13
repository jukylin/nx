package transactionmanager

import (
	"context"
	"testing"

	"github.com/jukylin/nx/saga/domain/entity"
	"github.com/jukylin/nx/saga/domain/repo"
	"github.com/jukylin/nx/saga/domain/value-object"
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

	TxgroupRepo.Create(ctx, &entity.Txgroup{Txid:100})
	TxrecordRepo.Create(ctx, &entity.Txrecord{Txid:100})
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"修改状态开始=》补偿", args{ctx, entity.Txgroup{Txid:100}}, false},
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
