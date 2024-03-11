package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	"github.com/shark-ci/shark-ci/internal/config"
	"github.com/shark-ci/shark-ci/internal/message_queue"
	pb "github.com/shark-ci/shark-ci/internal/proto"
	ciserverGrpc "github.com/shark-ci/shark-ci/internal/server/grpc"
	"github.com/shark-ci/shark-ci/internal/server/handler"
	"github.com/shark-ci/shark-ci/internal/server/middleware"
	"github.com/shark-ci/shark-ci/internal/server/service"
	"github.com/shark-ci/shark-ci/internal/server/session"
	"github.com/shark-ci/shark-ci/internal/server/store"
)

func fatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	err := config.LoadServerConfigFromEnv()
	if err != nil {
		fatal("Loading config from environment failed.", err)
	}

	session.InitSessionStore(config.ServerConf.SecretKey)

	slog.Info("Connecting to PostgreSQL.")
	pgStore, err := store.NewPostgresStore(context.TODO(), config.ServerConf.DB.URI)
	if err != nil {
		fatal("Connecting to PostgreSQL failed.", err)
	}
	defer pgStore.Close(context.TODO())

	err = pgStore.Ping(context.TODO())
	if err != nil {
		fatal("Pinging to PostgreSQL failed.", err)
	}
	slog.Info("PostgreSQL connected.")

	slog.Info("Connecting to RabbitMQ.")
	rabbitMQ, err := message_queue.NewRabbitMQ(config.ServerConf.MQ.URI)
	if err != nil {
		fatal("Connecting to RabbitMQ failed.", err)
	}
	defer rabbitMQ.Close(context.TODO())
	slog.Info("RabbitMQ connected.")

	store.Cleaner(pgStore, 24*time.Hour)

	services := service.InitServices(pgStore)

	slog.Info("Starting gRPC server.")
	lis, err := net.Listen("tcp", ":"+config.ServerConf.GRPCPort)
	if err != nil {
		fatal("Failed to listen.", err)
	}
	s := grpc.NewServer()
	grpcServer := ciserverGrpc.NewGRPCServer(pgStore, services)
	pb.RegisterPipelineReporterServer(s, grpcServer)
	go s.Serve(lis)
	slog.Info("gRPC server is running.", "port", config.ServerConf.GRPCPort)

	slog.Info("Starting HTTP server.")
	CSRF := csrf.Protect([]byte(config.ServerConf.SecretKey))

	indexHandler := handler.NewIndexHandler(pgStore)
	eventHandler := handler.NewEventHandler(pgStore, rabbitMQ, services)
	repoHandler := handler.NewRepoHandler(pgStore, services)
	authHandler := handler.NewAuthHandler(pgStore, services)

	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Handle("/", middleware.AuthMiddleware(pgStore)(indexHandler))
	r.HandleFunc("/login", authHandler.LoginPage)
	r.HandleFunc("/logout", authHandler.Logout)
	r.HandleFunc("/event_handler/{service}", eventHandler.HandleEvent).Methods(http.MethodPost)

	// OAuth2 subrouter.
	OAuth2 := r.PathPrefix("/oauth2").Subrouter()
	OAuth2.HandleFunc("/callback", authHandler.OAuth2Callback)

	// Repositories subrouter.
	repos := r.PathPrefix("/repositories").Subrouter()
	repos.Use(CSRF)
	repos.Use(middleware.AuthMiddleware(pgStore))
	repos.HandleFunc("", repoHandler.HandleRepos)
	repos.HandleFunc("/register/{service}/{repoID}", repoHandler.HandleRegisterRepo).Methods(http.MethodPost)
	repos.HandleFunc("/fetch-unregistered/{service}", repoHandler.FetchUnregistredRepos).Methods(http.MethodGet)

	server := &http.Server{
		Addr:         ":" + config.ServerConf.Port,
		Handler:      r,
		ReadTimeout:  0,
		WriteTimeout: 0,
		IdleTimeout:  0,
	}

	go func() {
		if err = server.ListenAndServe(); err != nil {
			fatal("HTTP server error.", err)
		}
	}()
	slog.Info("HTTP server is running.", "port", config.ServerConf.Port)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh
	slog.Info("Recived interrupt signal. Shuting down.")
}
