package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"

	ciserver "github.com/FilipSolich/shark-ci/ci-server"
	"github.com/FilipSolich/shark-ci/ci-server/api"
	"github.com/FilipSolich/shark-ci/ci-server/config"
	"github.com/FilipSolich/shark-ci/ci-server/handler"
	"github.com/FilipSolich/shark-ci/ci-server/middleware"
	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/session"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/ci-server/template"
	"github.com/FilipSolich/shark-ci/shared/message_queue"
)

func cleaner(s store.Storer, d time.Duration) {
	ticker := time.NewTicker(d)
	go func() {
		for {
			<-ticker.C
			err := s.Clean(context.TODO())
			if err != nil {
				slog.Error("store: databse cleanup failed", "err", err)
			}
		}
	}()
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	config, err := config.NewConfigFromEnv()
	if err != nil {
		logger.Error("creating config failed", "err", err)
		os.Exit(1)
	}

	session.InitSessionStore(config.CIServer.SecretKey)

	template.LoadTemplates()

	logger.Info("connecting to PostgreSQL")
	pgStore, err := store.NewPostgresStore(config.DB.URI)
	if err != nil {
		logger.Error("store: connecting to PostgreSQL failed", "err", err)
		os.Exit(1)
	}
	defer pgStore.Close(context.TODO())

	err = pgStore.Ping(context.TODO())
	if err != nil {
		logger.Error("store: pinging to PostgreSQL failed", "err", err)
		os.Exit(1)
	}
	logger.Info("PostgreSQL connected")

	logger.Info("connecting to RabbitMQ")
	rabbitMQ, err := message_queue.NewRabbitMQ(config.MQ.URI)
	if err != nil {
		logger.Error("mq: connecting to RabbitMQ failed", "err", err)
		os.Exit(1)
	}
	defer rabbitMQ.Close(context.TODO())
	logger.Info("RabbitMQ connected")

	cleaner(pgStore, 24*time.Hour)

	services := service.InitServices(pgStore, config)

	CSRF := csrf.Protect([]byte(config.CIServer.SecretKey))

	loginHandler := handler.NewLoginHandler(logger, pgStore, services)
	logoutHandler := handler.NewLogoutHandler()
	eventHandler := handler.NewEventHandler(logger, pgStore, rabbitMQ, services, config.CIServer)
	oauth2Handler := handler.NewOAuth2Handler(logger, pgStore, services)
	repoHandler := handler.NewRepoHandler(logger, pgStore, services)

	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Handle("/", middleware.AuthMiddleware(pgStore)(http.HandlerFunc(handler.IndexHandler)))
	r.HandleFunc("/login", loginHandler.HandleLoginPage)
	r.HandleFunc("/logout", logoutHandler.HandleLogout)
	r.HandleFunc(ciserver.EventPath+"/{service}", eventHandler.HandleEvent).Methods(http.MethodPost)

	// OAuth2 subrouter.
	OAuth2 := r.PathPrefix("/oauth2").Subrouter()
	OAuth2.HandleFunc("/callback", oauth2Handler.HandleCallback)

	// Repositories subrouter.
	repos := r.PathPrefix("/repositories").Subrouter()
	repos.Use(CSRF)
	repos.Use(middleware.AuthMiddleware(pgStore))
	repos.HandleFunc("", repoHandler.HandleRepos)
	repos.HandleFunc("/register", repoHandler.HandleRegisterRepo).Methods(http.MethodPost)
	repos.HandleFunc("/unregister", repoHandler.HandleUnregisterRepo).Methods(http.MethodPost)
	repos.HandleFunc("/activate", repoHandler.HandleActivateRepo).Methods(http.MethodPost)
	repos.HandleFunc("/deactivate", repoHandler.HandleDeactivateRepo).Methods(http.MethodPost)

	reposAPIHandler := api.NewRepoAPI(logger, pgStore, services)

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware(pgStore))
	api.Use(middleware.ContentTypeMiddleware)

	reposAPI := api.PathPrefix("/repos").Subrouter()
	reposAPI.HandleFunc("", reposAPIHandler.GetRepos).Methods(http.MethodGet)
	reposAPI.HandleFunc("/refresh", reposAPIHandler.RefreshRepos).Methods(http.MethodPost)
	reposAPI.HandleFunc("/{repoID}/webhook", reposAPIHandler.CreateWebhook).Methods(http.MethodPost)
	reposAPI.HandleFunc("/{repoID}/webhook", reposAPIHandler.DeleteWebhook).Methods(http.MethodDelete)

	server := &http.Server{
		Addr:         ":" + config.CIServer.Port,
		Handler:      r,
		ReadTimeout:  0,
		WriteTimeout: 0,
		IdleTimeout:  0,
	}
	logger.Info("server running on " + server.Addr)

	err = server.ListenAndServe()
	if err != nil {
		logger.Error("server crashed", "err", err)
		os.Exit(1)
	}
}
