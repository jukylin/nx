package transactionmanager

import (
	"context"
	"testing"

	"github.com/jukylin/nx/sagas/domain/entity"
	"github.com/mercari/grpc-http-proxy/proxy"
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
		{"补偿成功", args{ctx, entity.Txrecord{
			ID:1,
			Txid:100,
			Host:"http://127.0.0.1:8082",
			Path: "/compensate",
			Params:`{"name":"test"}`,
		}}, false},
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

	ctx := context.Background()

	grpcProxy := proxy.NewProxy()

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"GRPC 补偿", args{ctx, entity.Txrecord{
			RegAddress:"127.0.0.1:50051",
			ServiceName:"helloworld.GreeterServer",
			MethodName:"SayHello",
			Txid:10001,
			Params:`{"name":"grpc proxy"}`,
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gt := GRPCTransport{
				logger,
				grpcProxy,
			}
			if err := gt.Invoke(tt.args.ctx, tt.args.txrecord); (err != nil) != tt.wantErr {
				t.Errorf("GRPCTransport.Invoke() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
