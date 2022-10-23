package postgres

import (
	"embed"
	"errors"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"
	// migrate tools
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const _defaultAttempts = 5
const _defaultTimeout = 10 * time.Second

//go:embed migrations/*.sql
var fs embed.FS

func RunMigration(databaseURL string) {

	//if !strings.Contains(databaseURL, "sslmode") {
	//	databaseURL += "?sslmode=disable"
	//}

	var (
		attempts = _defaultAttempts
		err      error
		m        *migrate.Migrate
	)

	d, err := iofs.New(fs, "migrations")
	if err != nil {
		log.Error().Err(err).Msgf("Migrate: postgres connect error: %s", err)
		os.Exit(1)
	}

	for attempts > 0 {
		//m, err = migrate.New("file://migrations", databaseURL)
		m, err = migrate.NewWithSourceInstance("iofs", d, databaseURL)
		if err == nil {
			break
		}

		log.Info().Msgf("Migrate: postgres is trying to connect, attempts left: %d", attempts)
		time.Sleep(_defaultTimeout)
		attempts--
	}

	if err != nil {
		log.Error().Err(err).Msgf("Migrate: postgres connect error: %s", err)
		os.Exit(1)
	}

	err = m.Up()
	defer m.Close()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error().Err(err).Msgf("Migrate: up error: %s", err)
		os.Exit(1)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Info().Msg("Migrate: no change")
		return
	}

	log.Info().Msg("Migrate: up success")
}
