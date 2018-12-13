package main

import (
	"log"
	"qilin-api/cmd"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cmd.Execute()
}
