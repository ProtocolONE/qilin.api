package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"log"
	"os"
	"qilin-api/pkg/conf"
)

var (
	config  *conf.Config
	logger  *logrus.Entry
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "qilin",
	Short: "A brief description of your application",
	Long: `Qilin is an open source tool facilitating creation, 
distribution and activation of licenses for game content.

Qilin is a server application for manage api, database migrations 
and basic interactions with management data.
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile,
		"config", "c", "",
		"config file (default is $HOME/.qilin-config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error

	config, err = conf.LoadConfig(cfgFile)
	if err != nil {
		log.Fatal("Failed to load config: " + err.Error())
	}

	logger, err = conf.ConfigureLogging(&config.LogConfig)
	if err != nil {
		log.Fatal("Failed to configure logging: " + err.Error())
	}

	logger.Debugf("Config accepted")
}
