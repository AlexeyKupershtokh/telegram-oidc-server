package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/AlexeyKupershtokh/telegram-oidc-server/config"
	"github.com/AlexeyKupershtokh/telegram-oidc-server/exampleop"
	"github.com/AlexeyKupershtokh/telegram-oidc-server/storage"
)

func getUserStore(cfg *config.Config) (storage.UserStore, error) {
	if cfg.UsersFile == "" {
		return storage.NewUserStore(fmt.Sprintf("http://localhost:%s/", cfg.Port)), nil
	}
	return storage.StoreFromFile(cfg.UsersFile)
}

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatal(err)
	}
	logger := slog.New(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}),
	)

	issuer := cfg.Issuer

	storage.RegisterClients(
		//storage.NativeClient("native", cfg.RedirectURI...),
		//storage.WebClient("web", "secret", cfg.RedirectURI...),
		storage.WebClient("telegram-oidc", "secret", cfg.RedirectURI...),
		//storage.WebClient("api", "secret", cfg.RedirectURI...),
	)

	// the OpenIDProvider interface needs a Storage interface handling various checks and state manipulations
	// this might be the layer for accessing your database
	// in this example it will be handled in-memory
	store, err := getUserStore(&cfg)
	if err != nil {
		logger.Error("cannot create UserStore", "error", err)
		os.Exit(1)
	}
	storage := storage.NewStorage(store)
	router := exampleop.SetupServer(issuer, storage, logger, false)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}
	logger.Info("server listening, press ctrl+c to stop", "addr", issuer)
	err = server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server terminated", "error", err)
		os.Exit(1)
	}
}
