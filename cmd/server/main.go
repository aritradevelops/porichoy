package main

import (
	"log"
	"os"

	"github.com/aritradeveops/porichoy/internal/api"
)

func main() {
	err := api.Run()
	if err != nil {
		log.Printf("failed to start the service : %v", err)
		os.Exit(1)
	}
}
