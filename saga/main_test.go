package saga

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"context"
	"time"

	"github.com/jukylin/esim/config"
	ehttp "github.com/jukylin/esim/http"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/mysql"
	docker_test "github.com/jukylin/nx/saga/docker-test"
	"github.com/jukylin/nx/saga/domain/entity"
	"github.com/jukylin/nx/saga/domain/repo"
	"github.com/opentracing/opentracing-go"
	"github.com/ory/dockertest/v3"
	"gorm.io/gorm"
)

var logger log.Logger
var conf config.Config
var mysqlClient *mysql.Client
var httpClient *ehttp.Client
var pool *dockertest.Pool
var resource *dockertest.Resource

func TestMain(m *testing.M) {
	ez := log.NewEsimZap(
		log.WithEsimZapDebug(true),
	)

	glog := log.NewGormLogger(
		log.WithGLogEsimZap(ez),
	)

	logger = log.NewLogger(
		log.WithDebug(true),
		log.WithEsimZap(ez),
	)

	conf = config.NewMemConfig()
	conf.Set("debug", true)
	dt := docker_test.NewDockerTest(logger, 100)
	dt.RunMysql()

	clientOptions := mysql.ClientOptions{}
	mysqlClient = mysql.NewClient(
		clientOptions.WithLogger(logger),
		clientOptions.WithConf(conf),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: glog,
		}),
		clientOptions.WithDbConfig(
			[]mysql.DbConfig{
				mysql.DbConfig{
					Db:  "sagas",
					Dsn: "root:123456@tcp(localhost:3306)/sagas?charset=utf8&parseTime=True&loc=Local",
				},
				mysql.DbConfig{
					Db:  "sagas_slave",
					Dsn: "root:123456@tcp(localhost:3306)/sagas?charset=utf8&parseTime=True&loc=Local",
				},
			},
		),
		clientOptions.WithProxy(
			func() interface{} {
				monitorProxyOptions := mysql.MonitorProxyOptions{}
				return mysql.NewMonitorProxy(
					monitorProxyOptions.WithLogger(logger),
					monitorProxyOptions.WithConf(conf),
				)
			},
		),
	)

	httpClientOptions := ehttp.ClientOptions{}
	httpClient = ehttp.NewClient(
		httpClientOptions.WithLogger(logger),
		httpClientOptions.WithProxy(
			func() interface{} {
				monitorProxyOptions := ehttp.MonitorProxyOptions{}
				return ehttp.NewMonitorProxy(
					monitorProxyOptions.WithLogger(logger),
					monitorProxyOptions.WithConf(conf),
				)
			},
		),
	)

	InitHttpServer()

	code := m.Run()

	dt.Close()
	//mysqlClient.Close()
	// You can't defer this because os.Exit doesn't care for defer
	os.Exit(code)
}

func InitHttpServer() {
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/index", index1)
		mux.HandleFunc("/compensate", compensate1)

		err := http.ListenAndServe(":8081", mux)
		if err != nil {
			logger.Fatalf("ListenAndServe: ", err)
		}
	}()

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/index", index2)
		mux.HandleFunc("/compensate", compensate2)

		err := http.ListenAndServe(":8082", mux)
		if err != nil {
			logger.Fatalf("ListenAndServe: ", err)
		}
	}()
}

func index1(w http.ResponseWriter, r *http.Request) {
	txgroupRepo := repo.NewDbTxgroupRepo(logger)
	txrecordRepo := repo.NewDbTxrecordRepo(logger)

	es := NewEsimSagas(
		WithEssLogger(logger),
		WithEssTxgroupRepo(txgroupRepo),
		WithEssTxrecordRepo(txrecordRepo),
	)

	extractCarrier := opentracing.HTTPHeadersCarrier(r.Header)
	tc, err := es.Extract(r.Context(), opentracing.HTTPHeaders, extractCarrier)
	if err != nil {
		logger.Errorc(r.Context(), err.Error())
	}

	ctx := context.Background()
	ctx = ContextWithTxID(ctx, tc.TxID())
	saga, err := es.CreateSaga(r.Context(), tc.TxID())
	if err != nil {
		logger.Errorc(r.Context(), err.Error())
	}

	txrecord := entity.Txrecord{}
	txrecord.Host = "http://127.0.0.1:8081"
	txrecord.Path = "/compensate"
	txrecord.Params = `{"hello":"saga1"}`

	err = saga.StartSaga(r.Context(), txrecord)
	if err != nil {
		logger.Errorc(r.Context(), err.Error())
	}

	req, err := http.NewRequest("Get", "http://127.0.0.1:8082/index", nil)
	if err != nil {
		logger.Errorc(r.Context(), err.Error())
	}
	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	es.Inject(ctx, opentracing.HTTPHeaders, carrier)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorc(r.Context(), err.Error())
	}
	defer resp.Body.Close()

	saga.EndSaga(r.Context())

	fmt.Fprintf(w, "Hello saga1")
}

func compensate1(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello compensate1!")
}

func index2(w http.ResponseWriter, r *http.Request) {
	txgroupRepo := repo.NewDbTxgroupRepo(logger)
	txrecordRepo := repo.NewDbTxrecordRepo(logger)

	es := NewEsimSagas(
		WithEssLogger(logger),
		WithEssTxgroupRepo(txgroupRepo),
		WithEssTxrecordRepo(txrecordRepo),
	)

	extractCarrier := opentracing.HTTPHeadersCarrier(r.Header)
	tc, err := es.Extract(r.Context(), opentracing.HTTPHeaders, extractCarrier)
	if err != nil {
		logger.Errorc(r.Context(), err.Error())
	}

	saga, err := es.CreateSaga(r.Context(), tc.TxID())
	if err != nil {
		logger.Errorc(r.Context(), err.Error())
	}

	txrecord := entity.Txrecord{}
	txrecord.Host = "http://127.0.0.1:8082"
	txrecord.Path = "/compensate"
	txrecord.Params = `{"hello":"saga2"}`

	err = saga.StartSaga(r.Context(), txrecord)
	if err != nil {
		logger.Errorc(r.Context(), err.Error())
	}

	// TODO do something
	time.Sleep(100 * time.Millisecond)

	saga.EndSaga(r.Context())

	fmt.Fprintf(w, "Hello saga2")
}

func compensate2(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello compensate2!")
}
