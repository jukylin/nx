package transactionmanager

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	http2 "net/http"
	"net/url"
	"strconv"

	"github.com/jukylin/esim/grpc"
	"github.com/jukylin/esim/http"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/sagas"
	"github.com/jukylin/nx/sagas/domain/entity"
	value_object "github.com/jukylin/nx/sagas/domain/value-object"
	gpmetadata "github.com/mercari/grpc-http-proxy/metadata"
	"github.com/mercari/grpc-http-proxy/proxy"
	"google.golang.org/grpc/metadata"
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

func NewTransportFactory(options ...TfOption) *TransportFactory {
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
	err := txrecord.CheckHTTParam()
	if err != nil {
		return err
	}

	httpUrl := txrecord.BuildHTTPUrl()
	ht.logger.Infoc(ctx, "httpInvoker actionId: %d, txID: %d, url %s", txrecord.ID, txrecord.Txid, httpUrl)

	req, err := http2.NewRequestWithContext(ctx, http2.MethodPost, httpUrl, bytes.NewBuffer([]byte(txrecord.Params)))
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

	ht.logger.Infoc(ctx, "%s body: %s", httpUrl, body)

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

	u, err = url.Parse(txrecord.RegAddress)
	if err != nil {
		return err
	}

	gt.logger.Debugc(ctx, "target %s", u.String())

	err = gt.grpcProxy.Connect(ctx, u)
	if err != nil {
		return err
	}
	defer gt.grpcProxy.CloseConn()

	md := make(gpmetadata.Metadata)

	md["txid"] = []string{strconv.FormatUint(txrecord.Txid, 10)}
	ctx = metadata.NewOutgoingContext(ctx, (metadata.MD)(md))

	gt.logger.Infoc(ctx, "service name = %s, method name = %s, params = %s",
		txrecord.ServiceName, txrecord.MethodName, txrecord.Params)
	resp, err = gt.grpcProxy.Call(ctx, txrecord.ServiceName, txrecord.MethodName, []byte(txrecord.Params), &md)
	if err != nil {
		return err
	}

	gt.logger.Debugc(ctx, "grpc resp %s", resp)

	return nil
}
