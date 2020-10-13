package transactionmanager

import (
	"context"
	"testing"

	"github.com/jukylin/nx/saga/domain/entity"
)

func TestHTTPTransport_Invoke(t *testing.T) {
	type args struct {
		ctx      context.Context
		txrecord entity.Txrecord
	}

	ctx := context.Background()
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Host为空", args{ctx, entity.Txrecord{
			ID:1,
			Txid:100,
		}}, true},
		{"Params为空", args{ctx, entity.Txrecord{
			ID:1,
			Txid:100,
			Host:"http://127.0.0.1",
		}}, true},
		{"参数非json", args{ctx, entity.Txrecord{
			ID:1,
			Txid:100,
			Host:"http://127.0.0.1",
			Params:"123",
		}}, true},
		{"参数非json", args{ctx, entity.Txrecord{
			ID:1,
			Txid:100,
			Host:"http://127.0.0.1",
			Params:`{"name":"test"}`,
		}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ht := &HTTPTransport{
				logger:     logger,
				httpClient: httpClient,
			}
			if err := ht.Invoke(tt.args.ctx, tt.args.txrecord); (err != nil) != tt.wantErr {
				t.Errorf("HTTPTransport.Invoke() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGRPCTransport_Invoke(t *testing.T) {
	type args struct {
		ctx      context.Context
		txrecord entity.Txrecord
	}
	tests := []struct {
		name    string
		gt      *GRPCTransport
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gt := &GRPCTransport{}
			if err := gt.Invoke(tt.args.ctx, tt.args.txrecord); (err != nil) != tt.wantErr {
				t.Errorf("GRPCTransport.Invoke() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
