package main

import (
	"log"
)

func main() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
