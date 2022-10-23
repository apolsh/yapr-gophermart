package app

import (
	defaultLogger "log"
	"net/http"

	"github.com/apolsh/yapr-gophermart/config"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

func Run(cfg *config.Config) {

	//var userStorage storage.UserStorage = nil
	//var orderStorage storage.OrderStorage = nil
	decimal.MarshalJSONWithoutQuotes = true

	//if cfg.DatabaseType == "postgresql" {
	//	_, err := pgxpool.ParseConfig(cfg.DatabaseURI)
	//	if err != nil {
	//		log.Error().Err(err).Msg(err.Error())
	//		os.Exit(1)
	//	}
	//
	//	pool, err := pgxpool.Connect(context.Background(), cfg.DatabaseURI)
	//	if err != nil {
	//		err := errors.New("failed to connect to the database")
	//		log.Error().Err(err).Msg(err.Error())
	//		os.Exit(1)
	//	}
	//	postgres.RunMigration(cfg.DatabaseURI)
	//	defer pool.Close()
	//	orderStorage = postgres.NewOrderStoragePG(pool)
	//	userStorage = postgres.NewUserStoragePG(pool)
	//}

	//gophermartService, err := service.NewGophermartServiceImpl(*cfg, userStorage, orderStorage)
	//if err != nil {
	//	log.Error().Err(err).Msg(err.Error())
	//	os.Exit(1)
	//}

	r := chi.NewRouter()
	//httpserver.RegisterRoutes(r, nil)
	defaultLogger.Fatal(http.ListenAndServe(cfg.RunAddress, r))

	//httpServer := httpserver.NewServer(r, cfg)
	//
	//interrupt := make(chan os.Signal, 1)
	//signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	//
	//select {
	//case s := <-interrupt:
	//	log.Info().Str("message", "process interrupted by signal: "+s.String())
	//case err := <-httpServer.GetStopSignalChannel():
	//	err = fmt.Errorf("process interrupted by httpServer: %w", err)
	//	log.Error().Err(err).Msg(err.Error())
	//}
	//
	//err = httpServer.Shutdown()
	//if err != nil {
	//	log.Err(fmt.Errorf("error during server shutdown: %w", err))
	//}

}
