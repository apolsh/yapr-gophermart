package app

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/apolsh/yapr-gophermart/config"
	httpController "github.com/apolsh/yapr-gophermart/internal/gophermart/controller/httpserver"
	"github.com/apolsh/yapr-gophermart/internal/gophermart/service"
	postgresStorage "github.com/apolsh/yapr-gophermart/internal/gophermart/storage/postgres"
	"github.com/apolsh/yapr-gophermart/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/shopspring/decimal"
)

var log = logger.LoggerOfComponent("main")

//go:embed tls/cert.pem
var tlsCert []byte

//go:embed tls/key.pem
var tlsKey []byte

func Run(cfg *config.Config) {

	logger.SetGlobalLevel(cfg.LogLevel)
	decimal.MarshalJSONWithoutQuotes = true

	var userStorage service.UserStorage
	var orderStorage service.OrderStorage

	if cfg.DatabaseType == config.PostgresStorageType {
		_, err := pgxpool.ParseConfig(cfg.DatabaseURI)
		if err != nil {
			log.Fatal(fmt.Errorf("error during parsing DatabaseURI: %w", err))
		}
		databaseURL := cfg.DatabaseURI
		if !strings.Contains(databaseURL, "sslmode") {
			databaseURL += "?sslmode=disable"
		}

		pool, err := pgxpool.Connect(context.Background(), cfg.DatabaseURI)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to connect to the database: %w", err))
		}
		err = postgresStorage.RunMigration(databaseURL)
		if err != nil {
			log.Fatal(fmt.Errorf("error while executing database migration scripts: %w", err))
		}
		defer pool.Close()
		orderStorage = postgresStorage.NewOrderStoragePG(pool)
		userStorage = postgresStorage.NewUserStoragePG(pool)
	}

	gophermartService, err := service.NewGophermartServiceImpl(cfg.TokenSecretKey, cfg.LoyaltyServiceMaxTries, cfg.AccrualSystemAddress, userStorage, orderStorage)
	if err != nil {
		log.Fatal(fmt.Errorf("error while init app: %w", err))
	}
	err = gophermartService.StartAccrualInfoSynchronizer(context.Background(), cfg.LoyaltyServiceRateLimit)
	if err != nil {
		log.Fatal(fmt.Errorf("error while init app: %w", err))
	}

	r := chi.NewRouter()
	httpController.RegisterRoutes(r, gophermartService)

	done := make(chan bool)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGTERM, syscall.SIGQUIT)

	httpServerConfigs := &http.Server{
		Addr:         cfg.RunAddress,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	httpServer := httpController.NewServer(httpServerConfigs)

	go func() {
		<-quit
		log.Info("server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := httpServer.Stop(ctx); err != nil {
			log.Fatal(fmt.Errorf("could not gracefully shutdown the http server: %v", err))
		}
		gophermartService.Close()
		close(done)
	}()

	if cfg.HTTPSEnabled {
		tlsConfig, err := getTLSConfig()
		if err != nil {
			log.Fatal(fmt.Errorf("could not get TLS configs %v", err))
		}

		err = httpServer.StartTLS(tlsConfig)
		if err != nil {
			log.Fatal(fmt.Errorf("could not start grpc server: %v", err))
		}

	} else {

		err = httpServer.Start()
		if err != nil {
			log.Fatal(fmt.Errorf("could not start grpc server: %v", err))
		}

	}

	<-done
	log.Info("Server stopped")

}

func getTLSConfig() (*tls.Config, error) {
	cer, err := tls.X509KeyPair(tlsCert, tlsKey)
	if err != nil {
		return nil, err
	}

	return &tls.Config{Certificates: []tls.Certificate{cer}}, nil
}
