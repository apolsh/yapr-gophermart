package postgres

import (
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/apolsh/yapr-gophermart/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	// migrate tools
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const attempts = 5
const timeout = 10 * time.Second

var migrationLogger = logger.LoggerOfComponent("migrate_logger")

//go:embed migrations/*.sql
var fs embed.FS

func RunMigration(databaseURL string) error {

	var (
		attempts = attempts
		err      error
		m        *migrate.Migrate
	)

	d, err := iofs.New(fs, "migrations")
	if err != nil {
		err := fmt.Errorf("migrate: postgres connect error: %w", err)
		migrationLogger.Error(err)
		return err
	}

	for attempts > 0 {
		m, err = migrate.NewWithSourceInstance("iofs", d, databaseURL)
		if err == nil {
			break
		}

		migrationLogger.Info("migrate: postgres is trying to connect, attempts left: %d", attempts)
		time.Sleep(timeout)
		attempts--
	}

	if err != nil {
		err = fmt.Errorf("migrate: postgres connect error: %s", err)
		migrationLogger.Error(err)
		return err
	}

	if m == nil {
		return errors.New("migrate instance is nil")
	}

	err = m.Up()
	defer func(m *migrate.Migrate) {
		err, _ := m.Close()
		if err != nil {
			migrationLogger.Error(err)
		}
	}(m)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		err = fmt.Errorf("migrate: up error: %s", err)
		migrationLogger.Error(err)
		return err
	}

	if errors.Is(err, migrate.ErrNoChange) {
		migrationLogger.Info("Migrate: no change")
		return nil
	}

	migrationLogger.Info("Migrate: up success")
	return nil
}
