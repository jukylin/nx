package transactionmanager

import (
	"context"
	"github.com/jukylin/nx/saga/domain/value-object"
	"fmt"
	"bytes"
	"github.com/jukylin/nx/saga/domain/entity"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/http"
	"github.com/tidwall/gjson"
	http2 "net/http"
	"io/ioutil"
	"github.com/jukylin/esim/grpc"
	"github.com/jukylin/nx/saga"
	"strconv"
)

type Transport interface {
	Invoke(ctx context.Context, txrecord entity.Txrecord) error
}

type TransportFactory struct {
	logger log.Logger

	httpClient *http.Client

	grpcClient *grpc.Client
}

type TfOption func(*TransportFactory)

func NewTransportFactory(options ...TfOption) *TransportFactory  {
	tf := &TransportFactory{}
	for _, option := range options {
		option(tf)
	}

	return tf
}

func WithTfLogger(logger log.Logger) TfOption {
	return func(tf *TransportFactory) {
		tf.logger = logger
	}
}

func WithTfHTTPClient(httpClient *http.Client) TfOption {
	return func(tf *TransportFactory) {
		tf.httpClient = httpClient
	}
}

func WithTfGRPCClient(grpcClient *grpc.Client) TfOption {
	return func(tf *TransportFactory) {
		tf.grpcClient = grpcClient
	}
}

func (tf *TransportFactory) GetTransport(transportType int) (Transport, error) {
	if transportType == value_object.TranSportHTTP {
		return &HTTPTransport{
			tf.logger,
			tf.httpClient,
		}, nil
	}

	return nil, fmt.Errorf(ErrUnSupportTranSportType, transportType)

}

type HTTPTransport struct {
	logger log.Logger

	httpClient *http.Client
}

func (ht *HTTPTransport) Invoke(ctx context.Context, txrecord entity.Txrecord) error {
	if txrecord.Host == "" {
		return fmt.Errorf(ErrHostIsEmpty)
	}

	url := fmt.Sprintf("%s:%s", txrecord.Host, txrecord.Path)
	ht.logger.Infoc(ctx, "httpInvoker actionId: %d, txID: %d, url %s", txrecord.ID, txrecord.Txid, url)

	if txrecord.Params == "" {
		return fmt.Errorf(ErrParamsIsEmpty)
	}

	if !gjson.Valid(txrecord.Params) {
		return fmt.Errorf(ErrParamsMustJson)
	}

	req, err := http2.NewRequestWithContext(ctx, http2.MethodPost, url, bytes.NewBuffer([]byte(txrecord.Params)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(saga.TranContextHeaderName, strconv.FormatUint(txrecord.Txid, 10))

	resp , err := ht.httpClient.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http2.StatusOK {
		return fmt.Errorf(ErrHTTPStatus, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	ht.logger.Infoc(ctx, "%s body %s", url, body)

	return nil
}

type GRPCTransport struct {}

func (gt *GRPCTransport) Invoke(ctx context.Context, txrecord entity.Txrecord) error {
	return nil
}

