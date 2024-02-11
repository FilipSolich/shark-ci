package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	"github.com/shark-ci/shark-ci/internal/config"
	"github.com/shark-ci/shark-ci/internal/message_queue"
	pb "github.com/shark-ci/shark-ci/internal/proto"
	"github.com/shark-ci/shark-ci/internal/server/api"
	ciserverGrpc "github.com/shark-ci/shark-ci/internal/server/grpc"
	"github.com/shark-ci/shark-ci/internal/server/handlers"
	"github.com/shark-ci/shark-ci/internal/server/middleware"
	"github.com/shark-ci/shark-ci/internal/server/service"
	"github.com/shark-ci/shark-ci/internal/server/session"
	"github.com/shark-ci/shark-ci/internal/server/store"
	"github.com/shark-ci/shark-ci/internal/server/templates"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	err := config.LoadServerConfigFromEnv()
	if err != nil {
		slog.Error("Loading config from environment failed.", "err", err)
		os.Exit(1)
	}

	session.InitSessionStore(config.ServerConf.SecretKey)

	templates.ParseTemplates()

	slog.Info("connecting to PostgreSQL")
	pgStore, err := store.NewPostgresStore(config.ServerConf.DB.URI)
	if err != nil {
		slog.Error("store: connecting to PostgreSQL failed", "err", err)
		os.Exit(1)
	}
	defer pgStore.Close(context.TODO())

	err = pgStore.Ping(context.TODO())
	if err != nil {
		slog.Error("store: pinging to PostgreSQL failed", "err", err)
		os.Exit(1)
	}
	slog.Info("PostgreSQL connected")

	slog.Info("connecting to RabbitMQ")
	rabbitMQ, err := message_queue.NewRabbitMQ(config.ServerConf.MQ.URI)
	if err != nil {
		slog.Error("mq: connecting to RabbitMQ failed", "err", err)
		os.Exit(1)
	}
	defer rabbitMQ.Close(context.TODO())
	slog.Info("RabbitMQ connected")

	store.Cleaner(pgStore, 24*time.Hour)

	services := service.InitServices(pgStore)

	slog.Info("starting gRPC server")
	lis, err := net.Listen("tcp", ":"+config.ServerConf.GRPCPort)
	if err != nil {
		slog.Error("failed to listen", "err", err)
		os.Exit(1)
	}
	s := grpc.NewServer()
	grpcServer := ciserverGrpc.NewGRPCServer(pgStore, services)
	pb.RegisterPipelineReporterServer(s, grpcServer)
	go s.Serve(lis)
	slog.Info("gRPC server running", "port", config.ServerConf.GRPCPort)

	CSRF := csrf.Protect([]byte(config.ServerConf.SecretKey))

	loginHandler := handlers.NewLoginHandler(pgStore, services)
	logoutHandler := handlers.NewLogoutHandler()
	eventHandler := handlers.NewEventHandler(pgStore, rabbitMQ, services)
	oauth2Handler := handlers.NewOAuth2Handler(pgStore, services)
	repoHandler := handlers.NewRepoHandler(pgStore, services)

	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Handle("/", middleware.AuthMiddleware(pgStore)(http.HandlerFunc(handlers.IndexHandler)))
	r.HandleFunc("/login", loginHandler.HandleLoginPage)
	r.HandleFunc("/logout", logoutHandler.HandleLogout)
	r.HandleFunc("/event_handler/{service}", eventHandler.HandleEvent).Methods(http.MethodPost)

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

	reposAPIHandler := api.NewRepoAPI(pgStore, services)

	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware(pgStore))
	api.Use(middleware.ContentTypeMiddleware)

	reposAPI := api.PathPrefix("/repos").Subrouter()
	reposAPI.HandleFunc("", reposAPIHandler.GetRepos).Methods(http.MethodGet)
	reposAPI.HandleFunc("/refresh", reposAPIHandler.RefreshRepos).Methods(http.MethodPost)
	reposAPI.HandleFunc("/{repoID}/webhook", reposAPIHandler.CreateWebhook).Methods(http.MethodPost)
	reposAPI.HandleFunc("/{repoID}/webhook", reposAPIHandler.DeleteWebhook).Methods(http.MethodDelete)

	server := &http.Server{
		Addr:         ":" + config.ServerConf.Port,
		Handler:      r,
		ReadTimeout:  0,
		WriteTimeout: 0,
		IdleTimeout:  0,
	}
	slog.Info("server running", "addr", server.Addr)

	err = server.ListenAndServe()
	if err != nil {
		slog.Error("server crashed", "err", err)
		os.Exit(1)
	}
}
