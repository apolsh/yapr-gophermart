package postgres

import (
	"errors"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/rs/zerolog/log"
	// migrate tools
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const _defaultAttempts = 5
const _defaultTimeout = 10 * time.Second

func RunMigration(databaseURL string) {
	databaseURL += "?sslmode=disable"

	var (
		attempts = _defaultAttempts
		err      error
		m        *migrate.Migrate
	)
	log.Info().Msg(databaseURL)

	for attempts > 0 {
		m, err = migrate.New("file://cmd/internal/gophermart/storage/postgres/migrations", databaseURL)
		if err == nil {
			break
		}

		log.Info().Msgf("Migrate: postgres is trying to connect, attempts left: %d", attempts)
		time.Sleep(_defaultTimeout)
		attempts--
	}

	if err != nil {
		log.Error().Err(err).Msgf("Migrate: postgres connect error: %s", err)
	}

	err = m.Up()
	defer m.Close()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error().Err(err).Msgf("Migrate: up error: %s", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Info().Msg("Migrate: no change")
		return
	}

	log.Info().Msg("Migrate: up success")
}
