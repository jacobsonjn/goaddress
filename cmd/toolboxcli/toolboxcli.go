package main

import (
	"log"

	"github.com/jacobsonjn/goaddress/internal/bootstrap"
)

func main() {
	// Initialize Bootstrap with default config
	bs := bootstrap.NewBootstrap()
	if err := bs.Init().Execute(); err != nil {
		log.Fatalf("Failed to execute CLI: %v", err)
	}
}
