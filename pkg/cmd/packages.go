package cmd

import (
	"qilin-api/services/packages"
)

func init() {
	command.AddCommand(packages.Init())
}


