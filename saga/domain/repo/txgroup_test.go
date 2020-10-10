package repo

import (
	"context"
	"reflect"
	"testing"

	"github.com/jukylin/nx/saga/domain/entity"
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
