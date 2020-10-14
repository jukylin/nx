package example1

import (
	"net/http"
	"context"
	"fmt"
	"time"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/nx/sagas/domain/repo"
	"github.com/jukylin/nx/sagas"
	"github.com/opentracing/opentracing-go"
	"github.com/jukylin/nx/sagas/domain/entity"
)

var logger log.Logger

func main()  {

}

func InitHttpServer() {
	logger = log.NewLogger(
		log.WithDebug(true),
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

	extractCarrier := opentracing.HTTPHeadersCarrier(r.Header)
	tc, err := es.Extract(r.Context(), opentracing.HTTPHeaders, extractCarrier)
	if err != nil {
		logger.Errorc(r.Context(), err.Error())
	}

	ctx := context.Background()
	ctx = sagas.ContextWithTxID(ctx, tc.TxID())
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
