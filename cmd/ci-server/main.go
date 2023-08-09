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

	"github.com/shark-ci/shark-ci/ci-server/api"
	"github.com/shark-ci/shark-ci/ci-server/config"
	ciserverGrpc "github.com/shark-ci/shark-ci/ci-server/grpc"
	"github.com/shark-ci/shark-ci/ci-server/handler"
	"github.com/shark-ci/shark-ci/ci-server/middleware"
	"github.com/shark-ci/shark-ci/ci-server/service"
	"github.com/shark-ci/shark-ci/ci-server/session"
	"github.com/shark-ci/shark-ci/ci-server/store"
	"github.com/shark-ci/shark-ci/ci-server/template"
	"github.com/shark-ci/shark-ci/shared/message_queue"
	pb "github.com/shark-ci/shark-ci/shared/proto"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	conf, err := config.NewConfigFromEnv()
	if err != nil {
		slog.Error("creating config failed", "err", err)
		os.Exit(1)
	}
	config.Conf = conf

	session.InitSessionStore(conf.CIServer.SecretKey)

	template.LoadTemplates()

	slog.Info("connecting to PostgreSQL")
	pgStore, err := store.NewPostgresStore(conf.DB.URI)
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
	rabbitMQ, err := message_queue.NewRabbitMQ(conf.MQ.URI)
	if err != nil {
		slog.Error("mq: connecting to RabbitMQ failed", "err", err)
		os.Exit(1)
	}
	defer rabbitMQ.Close(context.TODO())
	slog.Info("RabbitMQ connected")

	store.Cleaner(pgStore, 24*time.Hour)

	services := service.InitServices(pgStore)

	slog.Info("starting gRPC server")
	lis, err := net.Listen("tcp", ":"+conf.CIServer.GRPCPort)
	if err != nil {
		slog.Error("failed to listen", "err", err)
		os.Exit(1)
	}
	s := grpc.NewServer()
	grpcServer := ciserverGrpc.NewGRPCServer(pgStore, services)
	pb.RegisterPipelineReporterServer(s, grpcServer)
	go s.Serve(lis)
	slog.Info("gRPC server running", "port", conf.CIServer.GRPCPort)

	CSRF := csrf.Protect([]byte(conf.CIServer.SecretKey))

	loginHandler := handler.NewLoginHandler(pgStore, services)
	logoutHandler := handler.NewLogoutHandler()
	eventHandler := handler.NewEventHandler(pgStore, rabbitMQ, services)
	oauth2Handler := handler.NewOAuth2Handler(pgStore, services)
	repoHandler := handler.NewRepoHandler(pgStore, services)

	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware)
	r.Handle("/", middleware.AuthMiddleware(pgStore)(http.HandlerFunc(handler.IndexHandler)))
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
		Addr:         ":" + conf.CIServer.Port,
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
