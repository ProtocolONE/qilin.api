package cmd

import (
	"qilin-api/pkg/api"
	"qilin-api/pkg/mongo"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run Qilin api server with given configuration",
	Run:   runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func runServer(cmd *cobra.Command, args []string) {
	session, err := mongo.NewSession(&config.Database)
	if err != nil {
		logger.Fatal("Failed to start mongo session: " + err.Error())
	}

	serverConfig := api.ServerConfig{
		Log:          logger,
		Jwt:          &config.Jwt,
		ServerConfig: &config.Server,
		Session:      session,
	}

	server, err := api.NewServer(&serverConfig)
	if err != nil {
		logger.Fatal("Failed to create server: " + err.Error())
	}

	logger.Infof("Starting up server")
	err = server.Start()
	if err != nil {
		logger.Fatal("Failed to start server: " + err.Error())
	}
}
