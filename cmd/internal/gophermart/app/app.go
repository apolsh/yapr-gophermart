package app

import (
	"context"
	"errors"
	"os"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/service"
	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/storage"
	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/storage/postgres"
	"github.com/apolsh/yapr-gophermart/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func Run(cfg *config.Config) {

	var userStorage storage.UserStorage = nil
	var orderStorage storage.OrderStorage = nil

	if cfg.DatabaseURI == "postgresql" {
		pool, err := pgxpool.New(context.Background(), cfg.DatabaseURI)
		if err != nil {
			log.Err(errors.New("failed to connect to the database"))
			os.Exit(1)
		}
		defer pool.Close()
		orderStorage = postgres.NewOrderStoragePG(pool)
		userStorage = postgres.NewUserStoragePG(pool)
	}

	service, err := service.NewGophermartServiceImpl(userStorage, orderStorage)
	if err != nil {
		log.Err(err)
		os.Exit(1)
	}

}
