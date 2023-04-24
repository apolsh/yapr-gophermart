package main

import (
	"log"

	"github.com/apolsh/yapr-gophermart/config"
	"github.com/apolsh/yapr-gophermart/internal/gophermart/app"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Reading config error: %s", err)
	}

	app.Run(cfg)
}
