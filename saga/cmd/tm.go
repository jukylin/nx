package main

import (
	"os"

	"github.com/jukylin/esim/log"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

var cfgFile string
var v = viper.New()
var logger = log.NewLogger(
	log.WithEsimZap(
		log.NewEsimZap(
			log.WithLogLevel(zapcore.InfoLevel),
		),
	),
)

var rootCmd = &cobra.Command{
	Use:   "tm",
	Short: "分布式事物管理器(Transaction Manager).",
	Long:  `事物管理器，负责管理全局事物，分配事物唯一标识，监控事物的执行进度，并负责事物的提交，回滚，失败恢复等.`,
	RunE: func(cmd *cobra.Command, args []string) error {

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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.esim.yaml)")

	rootCmd.PersistentFlags().BoolP("inject", "", true, "Automatic inject instance to infra")

	rootCmd.PersistentFlags().StringP("infra_dir", "", "internal/infra/", "Infra dir")

	rootCmd.PersistentFlags().StringP("infra_file", "", "infra.go", "Infra file name")

	rootCmd.PersistentFlags().BoolP("star", "", false, "With star")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	err := v.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		logger.Errorf(err.Error())
	}

	err = v.BindPFlags(rootCmd.Flags())
	if err != nil {
		logger.Errorf(err.Error())
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			logger.Errorf(err.Error())
			os.Exit(1)
		}

		// Search config in home directory with name ".esim" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".saga")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.Errorf("Using config file: %s", viper.ConfigFileUsed())
	}
}
