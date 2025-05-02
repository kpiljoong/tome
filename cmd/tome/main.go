package main

import (
	"log"

	"github.com/kpiljoong/tome/cmd/tome/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}

func Execute() {
}
