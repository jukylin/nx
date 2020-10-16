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

			txgroupRepo := repo.NewDbTxgroupRepo(logger)
			txrecordRepo := repo.NewDbTxrecordRepo(logger)

			ctx := context.Background()
			txID, _ := strconv.ParseUint(string(body), 10, 64)
			txgroup := txgroupRepo.FindByTxID(ctx, txID)
			if txgroup.State != value_object.TranEnd {
				logger.Errorf("事物未完成 %s", string(body))
			}

			c, _ := txrecordRepo.CountByTxID(ctx, txID)
			if c != 3 {
				logger.Errorf("补偿记录不对 %d", c)
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

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/index", index3)
		mux.HandleFunc("/compensate", compensate3)

		err := http.ListenAndServe(":8083", mux)
		if err != nil {
			logger.Fatalf("ListenAndServe: ", err)
		}
	}()


	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/index", index4)
		mux.HandleFunc("/compensate", compensate4)

		err := http.ListenAndServe(":8084", mux)
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

	ctx := sagas.ContextWithTxID(r.Context(), tran.Context().TxID())

	txrecord := entity.Txrecord{}
	txrecord.Host = "http://127.0.0.1:8081"
	txrecord.Path = "/compensate1"
	txrecord.Params = `{"hello":"saga1"}`
	txrecord.TransportType = value_object.TranSportHTTP

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

	tran.EndTransaction(ctx)

	fmt.Fprintf(w, "%d", tran.Context().TxID())
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
	txrecord.TransportType = value_object.TranSportHTTP

	err = saga.StartSaga(ctx, txrecord)
	if err != nil {
		logger.Errorc(ctx, err.Error())
	}

	req, err := http.NewRequest("Get", "http://127.0.0.1:8083/index", nil)
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

	saga.EndSaga(ctx)

	fmt.Fprintf(w, "Hello saga2")
}

func compensate2(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello compensate2!")
}


func index3(w http.ResponseWriter, r *http.Request) {
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
	txrecord.Host = "http://127.0.0.1:8083"
	txrecord.Path = "/compensate"
	txrecord.Params = `{"hello":"saga3"}`
	txrecord.Txid = tc.TxID()
	txrecord.TransportType = value_object.TranSportHTTP

	err = saga.StartSaga(ctx, txrecord)
	if err != nil {
		logger.Errorc(ctx, err.Error())
	}

	req, err := http.NewRequest("Get", "http://127.0.0.1:8084/index", nil)
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

	saga.EndSaga(ctx)

	fmt.Fprintf(w, "Hello saga3")
}

func compensate3(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello compensate2!")
}


func index4(w http.ResponseWriter, r *http.Request) {
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
	txrecord.Host = "http://127.0.0.1:8084"
	txrecord.Path = "/compensate"
	txrecord.Params = `{"hello":"saga4"}`
	txrecord.Txid = tc.TxID()
	txrecord.TransportType = value_object.TranSportHTTP

	err = saga.StartSaga(ctx, txrecord)
	if err != nil {
		logger.Errorc(ctx, err.Error())
	}

	// TODO do something
	time.Sleep(100 * time.Millisecond)

	saga.EndSaga(ctx)

	fmt.Fprintf(w, "Hello saga4")
}

func compensate4(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello compensate4!")
}