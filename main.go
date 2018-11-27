package main

import (
	"log"
	"qilin-api/cmd"
)

func main() {
	if err := cmd.ServerCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
