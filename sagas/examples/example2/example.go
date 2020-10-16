package main

import (
	"net/http"
	"fmt"
	"time"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/sagas/domain/repo"
	"github.com/jukylin/nx/sagas"
	"github.com/opentracing/opentracing-go"
	"github.com/jukylin/nx/sagas/domain/entity"
	"io/ioutil"
	"github.com/jukylin/esim/mysql"
	"github.com/jukylin/esim/config"
	"context"
	"strconv"
	"github.com/jukylin/nx/sagas/domain/value-object"
)

var logger log.Logger

func main()  {
	go func() {
		InitHttpServer()
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:8081/index")
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Errorf(err.Error())
		} else {
			logger.Infof("txID %s", string(body))
			time.Sleep(2 * time.Second)
			txgroupRepo := repo.NewDbTxgroupRepo(logger)

			ctx := context.Background()
			txID, _ := strconv.ParseUint(string(body), 10, 64)
			txgroup := txgroupRepo.FindByTxID(ctx, txID)
			if txgroup.State != value_object.TranCompensateFinish {
				logger.Errorf("事物没有补偿完成 %s %d", string(body), txgroup.State)
			}
		}
	}
}

func InitHttpServer() {
	logger = log.NewLogger(
		log.WithDebug(true),
	)

	conf := config.NewMemConfig()
	conf.Set("debug", true)
	clientOptions := mysql.ClientOptions{}
	mysql.NewClient(
		clientOptions.WithLogger(logger),
		clientOptions.WithConf(conf),
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

	es := sagas.NewEsimSagas(
		sagas.WithEssLogger(logger),
		sagas.WithEssTxgroupRepo(txgroupRepo),
		sagas.WithEssTxrecordRepo(txrecordRepo),
	)

	tran, err := es.StartTransaction(r.Context())
	if err != nil {
		logger.Errorc(r.Context(), err.Error())
	}
println(tran.Context().TxID())
	ctx := sagas.ContextWithTxID(r.Context(), tran.Context().TxID())

	req, err := http.NewRequest("Get", "http://127.0.0.1:8082/index", nil)
	if err != nil {
		logger.Errorc(ctx, err.Error())
	}
	carrier := opentracing.HTTPHeadersCarrier(req.Header)
	es.Inject(ctx, opentracing.HTTPHeaders, carrier)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorc(ctx, err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorc(ctx, err.Error())
	}
	if string(body) == "error" {
		logger.Errorc(ctx, "业务出错")
		fmt.Fprintf(w, "%d", tran.Context().TxID())
	} else {
		tran.EndTransaction(ctx)
		logger.Errorc(ctx, "事物完成")
		fmt.Fprintf(w, "%d", tran.Context().TxID())
	}
}

func compensate1(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello compensate1!")
}

func index2(w http.ResponseWriter, r *http.Request) {
	txgroupRepo := repo.NewDbTxgroupRepo(logger)
	txrecordRepo := repo.NewDbTxrecordRepo(logger)

	es := sagas.NewEsimSagas(
		sagas.WithEssLogger(logger),
		sagas.WithEssTxgroupRepo(txgroupRepo),
		sagas.WithEssTxrecordRepo(txrecordRepo),
	)

	extractCarrier := opentracing.HTTPHeadersCarrier(r.Header)
	tc, err := es.Extract(r.Context(), opentracing.HTTPHeaders, extractCarrier)
	if err != nil {
		logger.Errorc(r.Context(), err.Error())
	}

	ctx := sagas.ContextWithTxID(r.Context(), tc.TxID())

	saga, err := es.CreateSaga(ctx, tc.TxID())
	if err != nil {
		logger.Errorc(ctx, err.Error())
	}

	txrecord := entity.Txrecord{}
	txrecord.Host = "http://127.0.0.1:8082"
	txrecord.Path = "/compensate"
	txrecord.Params = `{"hello":"saga2"}`
	txrecord.Txid = tc.TxID()

	err = saga.StartSaga(ctx, txrecord)
	if err != nil {
		logger.Errorc(ctx, err.Error())
	}

	fmt.Fprintf(w, "error")
	return

	saga.EndSaga(ctx)

	fmt.Fprintf(w, "Hello saga2")
}

func compensate2(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello compensate2!")
}
