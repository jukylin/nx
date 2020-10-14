package repo

import (
	"context"
	"testing"

	"github.com/jukylin/nx/sagas/domain/entity"
	"gorm.io/gorm"
)

func TestDbTxcompensateRepo_InsertUpdateFromRecord(t *testing.T) {
	type args struct {
		ctx  context.Context
		db *gorm.DB
		txID uint64
	}

	ctx := context.Background()

	tr := NewDbTxrecordRepo(logger)
	tr.Create(ctx, &entity.Txrecord{
		Txid:100,
	})

	db := mysqlClient.GetCtxDb(ctx, "sagas").Begin()

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"写入数据", args{ctx, db, 100}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dtr := NewDbTxcompensateRepo(logger)
			if err := dtr.InsertUpdateFromRecord(tt.args.ctx, tt.args.db, tt.args.txID); (err != nil) != tt.wantErr {
				t.Errorf("DbTxcompensateRepo.InsertUpdateFromRecord() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
