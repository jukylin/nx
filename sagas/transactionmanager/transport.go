package transactionmanager

import (
	"context"
	"github.com/jukylin/nx/sagas/domain/value-object"
	"fmt"
	"bytes"
	"github.com/jukylin/nx/sagas/domain/entity"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/http"
	http2 "net/http"
	"io/ioutil"
	"github.com/jukylin/esim/grpc"
	"github.com/jukylin/nx/sagas"
	"strconv"
	"github.com/mercari/grpc-http-proxy/proxy"
	"net/url"
	"github.com/mercari/grpc-http-proxy/metadata"
)

type Transport interface {
	Invoke(ctx context.Context, txrecord entity.Txrecord) error
}

type TransportFactory struct {
	logger log.Logger

	httpClient *http.Client

	grpcClient *grpc.Client

	grpcProxy *proxy.Proxy
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

func WithTfGrpcProxy(grpcProxy *proxy.Proxy) TfOption {
	return func(tf *TransportFactory) {
		tf.grpcProxy = grpcProxy
	}
}

func (tf *TransportFactory) GetTransport(transportType int) (Transport, error) {
	if transportType == value_object.TranSportHTTP {
		return &HTTPTransport{
			tf.logger,
			tf.httpClient,
		}, nil
	} else if transportType == value_object.TranSportGRPC {
		return &GRPCTransport{
			tf.logger,
			tf.grpcProxy,
		}, nil
	}

	return nil, fmt.Errorf(ErrUnSupportTranSportType, transportType)
}

type HTTPTransport struct {
	logger log.Logger

	httpClient *http.Client
}

// 使用http协议调用补偿接口，响应状态码为200，即补偿成功
func (ht *HTTPTransport) Invoke(ctx context.Context, txrecord entity.Txrecord) error {
	if txrecord.Host == "" {
		return fmt.Errorf(ErrHostIsEmpty)
	}

	if txrecord.Params == "" {
		return fmt.Errorf(ErrParamsIsEmpty)
	}

	url := fmt.Sprintf("%s:%s", txrecord.Host, txrecord.Path)
	ht.logger.Infoc(ctx, "httpInvoker actionId: %d, txID: %d, url %s", txrecord.ID, txrecord.Txid, url)

	req, err := http2.NewRequestWithContext(ctx, http2.MethodPost, url, bytes.NewBuffer([]byte(txrecord.Params)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(sagas.TranContextHeaderName, strconv.FormatUint(txrecord.Txid, 10))

	resp, err := ht.httpClient.Do(ctx, req)
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

	ht.logger.Infoc(ctx, "%s body: %s", url, body)

	return nil
}

type GRPCTransport struct {
	logger log.Logger

	grpcProxy *proxy.Proxy
}

func (gt *GRPCTransport) Invoke(ctx context.Context, txrecord entity.Txrecord) error {
	var err error
	var u *url.URL
	var resp []byte

	u, err = url.ParseRequestURI(txrecord.RegAddress)
	if err != nil {
		return err
	}

	err = gt.grpcProxy.Connect(ctx, u)
	if err != nil {
		return err
	}

	md := make(metadata.Metadata, 1)
	md[sagas.TranContextHeaderName] = []string{strconv.FormatUint(txrecord.Txid, 10)}
	resp, err = gt.grpcProxy.Call(ctx, txrecord.ServiceName, txrecord.MethodName, []byte(txrecord.Params), &md)
	if err != nil {
		return err
	}

	gt.logger.Debugc(ctx, "grpc resp %s", resp)

	return nil
}

