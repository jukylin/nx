package saga

import (
	"context"
	"testing"
	"net/http"

	"github.com/jukylin/nx/saga/domain/repo"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/nx/saga/domain/entity"
	"github.com/opentracing/opentracing-go"
)

func TestEsimSagas_StartTransaction(t *testing.T) {
	txgroupRepo := repo.NewDbTxgroupRepo(logger)

	type args struct {
		ctx context.Context
	}

	ctx := context.Background()
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"创建一个事物", args{ctx}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			es := NewEsimSagas(
				WithEssLogger(logger),
				WithEssTxgroupRepo(txgroupRepo),
			)
			ts, err := es.StartTransaction(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("EsimSagas.StartTransaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tg := txgroupRepo.FindById(ctx, int64(ts.(*esimTransaction).txgroup.ID))
			assert.Equal(t, 1, tg.ID)
		})
	}
}

func TestEsimSaga_StartSaga(t *testing.T) {
	txgroupRepo := repo.NewDbTxgroupRepo(logger)
	txrecordRepo := repo.NewDbTxrecordRepo(logger)

	es := NewEsimSagas(
		WithEssLogger(logger),
		WithEssTxgroupRepo(txgroupRepo),
		WithEssTxrecordRepo(txrecordRepo),
	)

	ctx := context.Background()
	ts, err := es.StartTransaction(ctx)
	if err != nil {
		logger.Errorc(ctx, err.Error())
		return
	}

	ctx = ContextWithTxID(ctx, ts.Context().TxID())

	sage, err := es.CreateSaga(ctx, ts.Context().TxID())
	if err != nil {
		logger.Errorc(ctx, err.Error())
		ts.EndTransaction(ctx, TranFail)
		return
	}

	txrecord := entity.Txrecord{}
	txrecord.TransportType = 1
	txrecord.ServiceName = "test"
	txrecord.Host = "http:127.0.0.1:8080"
	txrecord.Path = "/compensate"
	txrecord.Params = `{"name":"test"}`
	// txrecord.
	txrecord.Txid = ts.Context().TxID()
	err = sage.StartSaga(ctx, txrecord)
	if err != nil {
		logger.Errorc(ctx, err.Error())
		ts.EndTransaction(ctx, TranFail)
		return
	}
	// TODO do something

	sage.EndSaga(ctx)

	ts.EndTransaction(ctx, TranSucc)
}

func TestEsimSaga_Inject_Extract(t *testing.T) {
	txgroupRepo := repo.NewDbTxgroupRepo(logger)
	txrecordRepo := repo.NewDbTxrecordRepo(logger)

	es := NewEsimSagas(
		WithEssLogger(logger),
		WithEssTxgroupRepo(txgroupRepo),
		WithEssTxrecordRepo(txrecordRepo),
	)

	ctx := context.Background()
	ts, err := es.StartTransaction(ctx)
	if err != nil {
		logger.Errorc(ctx, err.Error())
		return
	}

	type args struct {
		ctx context.Context
		format interface{}
		abstractCarrier interface{}
	}

	injectSuccCarrier := opentracing.HTTPHeadersCarrier(http.Header{})

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"无效的载体", args{ctx, nil, nil}, true},
		{"获取txID失败", args{ctx, nil, opentracing.HTTPHeadersCarrier(http.Header{})}, true},
		{"注入成功", args{ContextWithTxID(ctx, ts.Context().TxID()), nil, injectSuccCarrier}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = es.Inject(tt.args.ctx, tt.args.format, tt.args.abstractCarrier)
			if (err != nil) != tt.wantErr {
				t.Errorf("EsimSagas.Inject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

	var tc TransactionContext
	extractCarrier := opentracing.HTTPHeadersCarrier(http.Header{})
	extractCarrier.Set(TranContextHeaderName, "")

	testsExtract := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"无效的载体", args{ctx, nil, nil}, true},
		{"提取失败", args{ctx, nil, extractCarrier}, true},
		{"提取成功", args{ctx, nil, injectSuccCarrier}, false},
	}
	for _, tt := range testsExtract {
		t.Run(tt.name, func(t *testing.T) {
			tc, err = es.Extract(tt.args.ctx, tt.args.format, tt.args.abstractCarrier)
			if (err != nil) != tt.wantErr {
				t.Errorf("EsimSagas.Extract() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

	assert.Equal(t, ts.Context().TxID(), tc.TxID())
}

func TestEsimSagas(t *testing.T) {
	txgroupRepo := repo.NewDbTxgroupRepo(logger)
	txrecordRepo := repo.NewDbTxrecordRepo(logger)

	es := NewEsimSagas(
		WithEssLogger(logger),
		WithEssTxgroupRepo(txgroupRepo),
		WithEssTxrecordRepo(txrecordRepo),
	)

	ctx := context.Background()
	ts, err := es.StartTransaction(ctx)
	assert.Nil(t, err)
	ctx = ContextWithTxID(ctx, ts.Context().TxID())

	req, err := http.NewRequest("Get", "http://127.0.0.1:8081/index", nil)
	assert.Nil(t, err)

	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	es.Inject(ctx, opentracing.HTTPHeaders, carrier)
	client := http.Client{}
	resp, err := client.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	ts.EndTransaction(ctx, TranSucc)
}
