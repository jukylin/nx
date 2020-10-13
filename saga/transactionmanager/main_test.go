package transactionmanager

import (
	"os"
	"testing"
	"net/http"
	"fmt"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/mysql"
	docker_test "github.com/jukylin/nx/saga/docker-test"
	"gorm.io/gorm"
	"github.com/jukylin/esim/redis"
	"github.com/jukylin/nx/nxlock"
	"github.com/jukylin/nx/nxlock/nx-redis"
	ehttp "github.com/jukylin/esim/http"
	"github.com/davecgh/go-spew/spew"
)

var logger log.Logger
var conf config.Config
var mysqlClient *mysql.Client
var redisClient *redis.Client
var nl *nxlock.Nxlock
var httpClient *ehttp.Client

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

	// mdt := docker_test.MysqlDockerTest{}
	// mdt.InitMysql(logger)

	dt := docker_test.NewDockerTest(logger, 250)
	dt.RunReids()

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

	redisClientOptions := redis.ClientOptions{}
	redisClient = redis.NewClient(
		redisClientOptions.WithLogger(logger),
		redisClientOptions.WithConf(conf),
		redisClientOptions.WithProxy(func() interface{} {
			monitorProxyOptions := redis.MonitorProxyOptions{}
			return redis.NewMonitorProxy(
				monitorProxyOptions.WithLogger(logger),
				monitorProxyOptions.WithConf(conf),
			)
		}),
	)

	nxRedis := nx_redis.NewClient(
		nx_redis.WithLogger(logger),
		nx_redis.WithClient(redisClient),
	)

	nl = nxlock.NewNxlock(
		nxlock.WithLogger(logger),
		nxlock.WithSolution(nxRedis),
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

	StartHTTPServer()

	code := m.Run()

	mysqlClient.Close()
	redisClient.Close()

	dt.Close()
	//mysqlClient.Close()
	// You can't defer this because os.Exit doesn't care for defer
	os.Exit(code)
}

func StartHTTPServer()  {
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/compensate", compensate)

		err := http.ListenAndServe(":8082", mux)
		if err != nil {
			logger.Fatalf("ListenAndServe: ", err)
		}
	}()
}

func compensate(w http.ResponseWriter, r *http.Request) {
	spew.Dump(r.Header)
	fmt.Fprintf(w, "Hello compensate!")
}
