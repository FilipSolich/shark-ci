package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"

	ciserver "github.com/FilipSolich/shark-ci/ci-server"
	"github.com/FilipSolich/shark-ci/ci-server/config"
	"github.com/FilipSolich/shark-ci/ci-server/handler"
	"github.com/FilipSolich/shark-ci/ci-server/middleware"
	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/session"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/ci-server/template"
	"github.com/FilipSolich/shark-ci/shared/message_queue"
)

func clean(s store.Storer, d time.Duration, l *slog.Logger) {
	ticker := time.NewTicker(d)
	go func() {
		for {
			<-ticker.C
			err := s.Clean(context.TODO())
			if err != nil {
				l.Error("store: databse cleanup failed", "err", err)
			}
		}
	}()
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// TODO: Remove godotenv.
	err := godotenv.Load()
	if err != nil {
		logger.Error("loading environment failed", "err", err)
		os.Exit(1)
	}

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

	logger.Info("connecting to MongoDB")
	mongoStore, err := store.NewMongoStore(config.DB.URI)
	if err != nil {
		logger.Error("store: connecting to MongoDB failed", "err", err)
		os.Exit(1)
	}
	defer mongoStore.Close(context.TODO())

	err = mongoStore.Ping(context.TODO())
	if err != nil {
		logger.Error("store: pinging to MongoDB failed", "err", err)
		os.Exit(1)
	}
	logger.Info("MongoDB connected")

	logger.Info("connecting to RabbitMQ")
	rabbitMQ, err := message_queue.NewRabbitMQ(config.MQ.URI)
	if err != nil {
		logger.Error("mq: connecting to RabbitMQ failed", "err", err)
		os.Exit(1)
	}
	defer rabbitMQ.Close(context.TODO())
	logger.Info("RabbitMQ connected")

	clean(mongoStore, 24*time.Hour, logger)

	services := service.InitServices(mongoStore, config)

	CSRF := csrf.Protect([]byte(config.CIServer.SecretKey))

	loginHandler := handler.NewLoginHandler(logger, mongoStore, services)
	logoutHandler := handler.NewLogoutHandler()
	eventHandler := handler.NewEventHandler(logger, mongoStore, rabbitMQ, services)
	oauth2Handler := handler.NewOAuth2Handler(logger, mongoStore, services)
	repoHandler := handler.NewRepoHandler(logger, mongoStore, services)

	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Handle("/", middleware.AuthMiddleware(mongoStore)(http.HandlerFunc(handler.IndexHandler)))
	r.HandleFunc("/login", loginHandler.HandleLoginPage)
	r.HandleFunc("/logout", logoutHandler.HandleLogout)
	r.HandleFunc(ciserver.EventHandlerPath+"/{service}", eventHandler.HandleEvent).Methods(http.MethodPost)

	// OAuth2 subrouter.
	OAuth2 := r.PathPrefix("/oauth2").Subrouter()
	OAuth2.HandleFunc("/callback", oauth2Handler.HandleCallback)

	// Repositories subrouter.
	repos := r.PathPrefix("/repositories").Subrouter()
	repos.Use(CSRF)
	repos.Use(middleware.AuthMiddleware(mongoStore))
	repos.HandleFunc("", repoHandler.HandleRepos)
	repos.HandleFunc("/register", repoHandler.HandleRegisterRepo).Methods(http.MethodPost)
	repos.HandleFunc("/unregister", repoHandler.HandleUnregisterRepo).Methods(http.MethodPost)
	repos.HandleFunc("/activate", repoHandler.HandleActivateRepo).Methods(http.MethodPost)
	repos.HandleFunc("/deactivate", repoHandler.HandleDeactivateRepo).Methods(http.MethodPost)

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
