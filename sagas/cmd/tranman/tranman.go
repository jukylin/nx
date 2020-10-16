package main

import (
	"os"

	"github.com/jukylin/esim/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	"github.com/jukylin/nx/sagas/transactionmanager"
	"github.com/jukylin/nx/nxlock/nx-redis"
	"github.com/jukylin/nx/nxlock"
	"github.com/jukylin/esim/redis"
	"github.com/jukylin/nx/sagas/domain/repo"
	"github.com/jukylin/esim/mysql"
	"context"
	"os/signal"
	"syscall"
	"github.com/jukylin/esim/config"
	"github.com/opentracing/opentracing-go"
)

var logger log.Logger

var rootCmd = &cobra.Command{
	Use:   "tm",
	Short: "分布式事物管理器(Transaction Manager).",
	Long:  `事物管理器，负责管理全局事物，分配事物唯一标识，监控事物的执行进度，并负责事物的提交，回滚，失败恢复等.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		viperConfOptions := config.ViperConfOptions{}

		conf := config.NewViperConfig(
			viperConfOptions.WithConfigType("yaml"),
			viperConfOptions.WithConfFile([]string{"saga.yaml"}),
		)

		var logLevel zapcore.Level
		if conf.GetBool("debug") {
			logLevel = zapcore.DebugLevel
		} else {
			logLevel = zapcore.InfoLevel
		}

		logger = log.NewLogger(
			log.WithEsimZap(
				log.NewEsimZap(
					log.WithLogLevel(logLevel),
				),
			),
			log.WithDebug(conf.GetBool("debug")),
		)

		// 初始化分布式事物锁
		redisClientOptions := redis.ClientOptions{}
		redisClient := redis.NewClient(
			redisClientOptions.WithLogger(logger),
			redisClientOptions.WithConf(conf),
			redisClientOptions.WithProxy(func() interface{} {
				monitorProxyOptions := redis.MonitorProxyOptions{}
				return redis.NewMonitorProxy(
					monitorProxyOptions.WithLogger(logger),
					monitorProxyOptions.WithConf(conf),
					monitorProxyOptions.WithTracer(&opentracing.NoopTracer{}),
				)
			}),
		)

		nxRedis := nx_redis.NewClient(
			nx_redis.WithLogger(logger),
			nx_redis.WithClient(redisClient),
		)

		nl := nxlock.NewNxlock(
			nxlock.WithLogger(logger),
			nxlock.WithSolution(nxRedis),
		)
		// 分布式事物锁 初始化完成

		clientOptions := mysql.ClientOptions{}
		mysqlClient := mysql.NewClient(
			clientOptions.WithLogger(logger),
			clientOptions.WithConf(conf),
			clientOptions.WithProxy(
				func() interface{} {
					monitorProxyOptions := mysql.MonitorProxyOptions{}
					return mysql.NewMonitorProxy(
						monitorProxyOptions.WithLogger(logger),
						monitorProxyOptions.WithConf(conf),
						monitorProxyOptions.WithTracer(&opentracing.NoopTracer{}),
					)
				},
			),
		)

		txgroupRepo := repo.NewDbTxgroupRepo(logger)
		txrecordRepo := repo.NewDbTxrecordRepo(logger)
		txcompensateRepo := repo.NewDbTxcompensateRepo(logger)

		// 逆向补偿
		backwardCompensate := transactionmanager.NewBackwardCompensate(
			transactionmanager.WithBcLogger(logger),
			transactionmanager.WithBcMysqlClient(mysqlClient),
			transactionmanager.WithBcTxgroupRepo(txgroupRepo),
			transactionmanager.WithBcTxrecordRepo(txrecordRepo),
			transactionmanager.WithBcTxcompensateRepo(txcompensateRepo),
		)

		// TM 初始化
		tm :=  transactionmanager.NewTransactionManager(
			transactionmanager.WithTmLogger(logger),
			transactionmanager.WithTmNxLock(nl),
			transactionmanager.WithTmTxgroupRepo(txgroupRepo),
			transactionmanager.WithTmTxrecordRepo(txrecordRepo),
			transactionmanager.WithTmTxcompensateRepo(txcompensateRepo),
			transactionmanager.WithTmCompensate(backwardCompensate),
		)

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)

		logger.Infoc(ctx, "Startup Transaction Manager")
		tm.Start(ctx)

		c := make(chan os.Signal, 1)
		signal.Reset(syscall.SIGTERM, syscall.SIGINT)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

		s := <-c
		logger.Infof(`Received signal "%s"; beginning shutdown`, s.String())
		cancel()
		tm.Close()
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
}

//nolint:lll
func init() {
	// cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./.saga.yaml", "config file (default is $HOME/.saga.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//err := v.BindPFlags(rootCmd.PersistentFlags())
	//if err != nil {
	//	logger.Errorf(err.Error())
	//}
	//
	//err = v.BindPFlags(rootCmd.Flags())
	//if err != nil {
	//	logger.Errorf(err.Error())
	//}
}

// initConfig reads in config file and ENV variables if set.
//func initConfig() {
//	if cfgFile != "" {
//		// Use config file from the flag.
//		viper.SetConfigFile(cfgFile)
//	} else {
//		// Find home directory.
//		home, err := homedir.Dir()
//		if err != nil {
//			logger.Errorf(err.Error())
//			os.Exit(1)
//		}
//
//		// Search config in home directory with name ".esim" (without extension).
//		viper.AddConfigPath(home)
//		viper.SetConfigName(".saga")
//		viper.SetConfigType("yaml")
//	}
//
//	viper.AutomaticEnv() // read in environment variables that match
//
//	// If a config file is found, read it in.
//	if err := viper.ReadInConfig(); err != nil {
//		logger.Errorf("Using config file: %s", viper.ConfigFileUsed())
//	}
//}
