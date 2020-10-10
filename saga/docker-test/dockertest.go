package docker_test

import (
	"database/sql"

	"github.com/jukylin/esim/log"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
)

type MysqlDockerTest struct {
	pool *dockertest.Pool

	resource *dockertest.Resource

	db *sql.DB
}

func (mdt *MysqlDockerTest) InitMysql(logger log.Logger) {
	var err error
	mdt.pool, err = dockertest.NewPool("")
	if err != nil {
		logger.Fatalf("Could not connect to docker: %s", err)
	}

	opt := &dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "latest",
		Env:        []string{"MYSQL_ROOT_PASSWORD=123456"},
	}

	// pulls an image, creates a container based on it and runs it
	mdt.resource, err = mdt.pool.RunWithOptions(opt, func(hostConfig *dc.HostConfig) {
		hostConfig.PortBindings = map[dc.Port][]dc.PortBinding{
			"3306/tcp": {{HostIP: "", HostPort: "3306"}},
		}
	})
	if err != nil {
		logger.Fatalf("Could not start resource: %s", err.Error())
	}

	err = mdt.resource.Expire(150)
	if err != nil {
		logger.Fatalf(err.Error())
	}

	if err := mdt.pool.Retry(func() error {
		var err error
		mdt.db, err = sql.Open("mysql",
			"root:123456@tcp(localhost:3306)/mysql?charset=utf8&parseTime=True&loc=Local")
		if err != nil {
			return err
		}
		mdt.db.SetMaxOpenConns(100)

		return mdt.db.Ping()
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
		res, err := mdt.db.Exec(execSQL)
		if err != nil {
			logger.Errorf(err.Error())
		}
		_, err = res.RowsAffected()
		if err != nil {
			logger.Errorf(err.Error())
		}
	}
}

func (mdt *MysqlDockerTest) Close(logger log.Logger) {
	if err := mdt.pool.Purge(mdt.resource); err != nil {
		logger.Fatalf("Could not purge resource: %s", err)
	}
	mdt.db.Close()
}
