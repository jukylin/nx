package repo

import (
	"context"
	"reflect"
	"testing"

	"github.com/jukylin/nx/sagas/domain/entity"
	"github.com/magiconair/properties/assert"
	"time"
)

func TestDbTxgroupRepo_GetCompensateList(t *testing.T) {
	type args struct {
		ctx   context.Context
		limit int
	}
	ctx := context.Background()
	tests := []struct {
		name    string
		args    args
		want    []entity.Txgroup
		wantErr bool
	}{
		{"查询数据", args{ctx, 200}, []entity.Txgroup{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dtr := NewDbTxgroupRepo(logger)
			got, err := dtr.GetCompensateList(tt.args.ctx, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("DbTxgroupRepo.GetCompensateList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DbTxgroupRepo.GetCompensateList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDbTxgroupRepo_GetUnfishedTransactionGroup(t *testing.T) {
	type args struct {
		ctx       context.Context
		intervals int
	}

	ctx := context.Background()
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"查询为结束的事物", args{ctx, 3600},2 ,false},
	}

	dtr := NewDbTxgroupRepo(logger)
	dtr.Create(ctx, &entity.Txgroup{
		Txid:100,
		CreateTime:time.Now().Add(- 3600 * 2 * time.Second),
	})
	dtr.Create(ctx, &entity.Txgroup{
		Txid:200,
		CreateTime:time.Now().Add(- 3600 * 2 * time.Second),
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dtr.GetUnfishedTransactionGroup(tt.args.ctx, tt.args.intervals)
			if (err != nil) != tt.wantErr {
				t.Errorf("DbTxgroupRepo.GetUnfishedTransactionGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, len(got))
		})
	}
}
