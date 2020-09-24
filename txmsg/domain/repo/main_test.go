package repo

import (
	"database/sql"
	"os"

	"testing"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/mysql"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
)

var db *sql.DB
var logger log.Logger
var conf config.Config
var mysqlClient *mysql.Client

func TestMain(m *testing.M) {
	logger = log.NewLogger(
		log.WithDebug(true),
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
		`create database txmsg;`,
		`CREATE TABLE IF NOT EXISTS txmsg.msg_info(
			 id int not NULL auto_increment,
			 content VARCHAR(300) not NULL DEFAULT '',
			 topic varchar(30) not null default '',
			 tag varchar(50) not null default '',
			 status int not null default 0,
			 create_time timestamp not null ,
			 delay int not null default 0,
			 PRIMARY KEY (id)
		)engine=innodb;`}

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
		clientOptions.WithDbConfig(
			[]mysql.DbConfig{
				mysql.DbConfig{
					Db:  "txmsg",
					Dsn: "root:123456@tcp(localhost:3306)/txmsg?charset=utf8&parseTime=True&loc=Local",
				},
				mysql.DbConfig{
					Db:  "txmsg_slave",
					Dsn: "root:123456@tcp(localhost:3306)/txmsg?charset=utf8&parseTime=True&loc=Local",
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

	db.Close()
	//mysqlClient.Close()
	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		logger.Fatalf("Could not purge resource: %s", err)
	}
	os.Exit(code)
}
