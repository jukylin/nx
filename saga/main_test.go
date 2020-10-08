package saga

import (
	"database/sql"
	"os"
	"net/http"
	"fmt"
	"testing"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	ehttp "github.com/jukylin/esim/http"
	"github.com/jukylin/esim/mysql"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	"gorm.io/gorm"
	"github.com/jukylin/nx/saga/domain/repo"
	"github.com/opentracing/opentracing-go"
	"github.com/jukylin/nx/saga/domain/entity"
	"time"
	"context"
)

var db *sql.DB
var logger log.Logger
var conf config.Config
var mysqlClient *mysql.Client
var httpClient *ehttp.Client

func TestMain(m *testing.M) {
	ez := log.NewEsimZap(
		log.WithEsimZapDebug(true),
	)

	logger = log.NewLogger(
		log.WithDebug(true),
		log.WithEsimZap(ez),
	)

	glog := log.NewGormLogger(
		log.WithGLogEsimZap(ez),
	)

	conf = config.NewMemConfig()
	conf.Set("debug", true)

	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Fatalf("Could not connect to docker: %s", err)
	}

	opt := &dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "latest",
		Env:        []string{"MYSQL_ROOT_PASSWORD=123456"},
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(opt, func(hostConfig *dc.HostConfig) {
		hostConfig.PortBindings = map[dc.Port][]dc.PortBinding{
			"3306/tcp": {{HostIP: "", HostPort: "3306"}},
		}
	})
	if err != nil {
		logger.Fatalf("Could not start resource: %s", err.Error())
	}

	err = resource.Expire(150)
	if err != nil {
		logger.Fatalf(err.Error())
	}

	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open("mysql",
			"root:123456@tcp(localhost:3306)/mysql?charset=utf8&parseTime=True&loc=Local")
		if err != nil {
			return err
		}
		db.SetMaxOpenConns(100)

		return db.Ping()
	}); err != nil {
		logger.Fatalf("Could not connect to docker: %s", err)
	}

	sqls := []string{
		`create database sagas;`,
`CREATE TABLE sagas.txcompensate (
  id int COLLATE utf8mb4_general_ci not NULL auto_increment,
  txid bigint unsigned Not NULL default 0 COMMENT '事务编号',
  success int not NULL default 0,
  step int not NULL,
  create_time datetime not NULL DEFAULT CURRENT_TIMESTAMP,
  update_time datetime not NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除标识',
  PRIMARY KEY (id) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 comment="补偿结果" COLLATE=utf8mb4_general_ci;`,
`CREATE TABLE sagas.txgroup (
  id int COLLATE utf8mb4_general_ci not NULL auto_increment,
  txid bigint unsigned Not NULL default 0 COMMENT '事务编号',
  state int not NULL,
  priority int not NULL,
  create_time datetime not NULL,
  update_time datetime not NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除标识',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 comment="事物主表" COLLATE=utf8mb4_general_ci;`,
`CREATE TABLE sagas.txrecord (
  id int COLLATE utf8mb4_general_ci not NULL auto_increment,
  txid bigint unsigned Not NULL default 0 COMMENT '事务编号',
  manner_name varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  method_name varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  compensate_name varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  class_name varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  service_name varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  generic_param_types varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  param_types varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  params varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci not NULL,
  step smallint not NULL,
  lookup varchar(255) COLLATE utf8mb4_general_ci not NULL,
  reg_address varchar(500) COLLATE utf8mb4_general_ci not NULL,
  version varchar(255) COLLATE utf8mb4_general_ci not NULL,
  transport_type int COLLATE utf8mb4_general_ci not NULL default 0,
  host varchar(255) COLLATE utf8mb4_general_ci not NULL default '',
  path varchar(255) COLLATE utf8mb4_general_ci not NULL default '',
  create_time datetime not NULL DEFAULT CURRENT_TIMESTAMP,
  update_time datetime not NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  is_deleted TINYINT(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '删除标识',
  PRIMARY KEY (id) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 comment="事物步骤" COLLATE=utf8mb4_general_ci;`,
		}

	for _, execSQL := range sqls {
		res, err := db.Exec(execSQL)
		if err != nil {
			logger.Errorf(err.Error())
		}
		_, err = res.RowsAffected()
		if err != nil {
			logger.Errorf(err.Error())
		}
	}


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

	db.Close()
	//mysqlClient.Close()
	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		logger.Fatalf("Could not purge resource: %s", err)
	}
	os.Exit(code)
}

func InitHttpServer()  {
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
	es.Inject(ctx,  opentracing.HTTPHeaders, carrier)
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
	fmt.Fprintf(w, "Hello golang http!")
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
	fmt.Fprintf(w, "Hello golang http!")
}
