package repo

import (
	"os"
	"testing"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/mysql"
	docker_test "github.com/jukylin/nx/saga/docker-test"
	"gorm.io/gorm"
)

var logger log.Logger
var conf config.Config
var mysqlClient *mysql.Client

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

	mdt := docker_test.MysqlDockerTest{}
	mdt.InitMysql(logger)

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

	code := m.Run()

	mdt.Close(logger)
	//mysqlClient.Close()
	// You can't defer this because os.Exit doesn't care for defer
	os.Exit(code)
}
