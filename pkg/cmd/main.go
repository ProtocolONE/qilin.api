package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	command = &cobra.Command{}
)

func Execute() {
	if err := command.Execute(); err != nil {
		log.Fatalf("Command execution failed with error %s", err.Error())
		os.Exit(1)
	}
}
