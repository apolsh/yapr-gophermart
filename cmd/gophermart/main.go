package main

import (
	"log"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/app"
	"github.com/apolsh/yapr-gophermart/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Reading config error: %s", err)
	}

	app.Run(cfg)
}
