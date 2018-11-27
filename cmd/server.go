package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"qilin-api/pkg/api"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/mongo"
)

var serverCmd = cobra.Command{
	Use:   "server",
	Short: "Run Qilin api server with given configuration",
	Run:   run,
}

func ServerCommand() *cobra.Command {
	serverCmd.PersistentFlags().StringP("config", "c", "", "the config file to use")
	return &serverCmd
}

func run(cmd *cobra.Command, args []string) {
	config, err := conf.LoadConfig(cmd)
	if err != nil {
		log.Fatal("Failed to load config: " + err.Error())
	}

	logger, err := conf.ConfigureLogging(&config.LogConfig)
	if err != nil {
		log.Fatal("Failed to configure logging: " + err.Error())
	}

	session, err := mongo.NewSession(&config.Database)
	if err != nil {
		log.Fatal("Failed to start mongo session: " + err.Error())
	}

	serverConfig := api.ServerConfig{
		Log:          logger,
		Jwt:          &config.Jwt,
		ServerConfig: &config.Server,
		Session:      session,
	}

	server, err := api.NewServer(&serverConfig)
	if err != nil {
		log.Fatal("Failed to create server: " + err.Error())
	}

	logger.Infof("Starting up server")
	err = server.Start()
	if err != nil {
		log.Fatal("Failed to start server: " + err.Error())
	}
}
