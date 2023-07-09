package main

import (
	"context"
	std_log "log"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	ciserver "github.com/FilipSolich/shark-ci/ci-server"
	"github.com/FilipSolich/shark-ci/ci-server/config"
	"github.com/FilipSolich/shark-ci/ci-server/handlers"
	"github.com/FilipSolich/shark-ci/ci-server/log"
	"github.com/FilipSolich/shark-ci/ci-server/middlewares"
	"github.com/FilipSolich/shark-ci/ci-server/service"
	"github.com/FilipSolich/shark-ci/ci-server/sessions"
	"github.com/FilipSolich/shark-ci/ci-server/store"
	"github.com/FilipSolich/shark-ci/ci-server/template"
	"github.com/FilipSolich/shark-ci/message_queue"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		std_log.Fatal(err)
	}
	defer logger.Sync()
	log.L = logger.Sugar()

	// TODO: Remove godotenv.
	err = godotenv.Load()
	if err != nil {
		log.L.Fatal(err)
	}

	config, err := config.NewConfigFromEnv()
	if err != nil {
		log.L.Fatal(err)
	}

	sessions.InitSessionStore(config.CIServer.SecretKey)

	template.LoadTemplates()

	mongoStore, err := store.NewMongoStore(config.DB.URI)
	if err != nil {
		log.L.Fatal(err)
	}
	defer mongoStore.Close(context.TODO())

	rabbitMQ, err := message_queue.NewRabbitMQ(config.MQ.URI)
	if err != nil {
		log.L.Fatal(err)
	}
	defer rabbitMQ.Close(context.TODO())

	serviceMap := service.InitServices(mongoStore, config)

	CSRF := csrf.Protect([]byte(config.CIServer.SecretKey))

	loginHandler := handlers.NewLoginHandler(mongoStore, serviceMap)
	logoutHandler := handlers.NewLogoutHandler()
	eventHandler := handlers.NewEventHandler(log.L, mongoStore, rabbitMQ, serviceMap)
	oauth2Handler := handlers.NewOAuth2Handler(mongoStore, serviceMap)
	repoHandler := handlers.NewRepoHandler(log.L, mongoStore, serviceMap)
	jobHandler := handlers.NewJobHandler(mongoStore, serviceMap)

	r := mux.NewRouter()
	r.Use(middlewares.LoggingMiddleware)
	r.Handle("/", middlewares.AuthMiddleware(mongoStore)(http.HandlerFunc(handlers.IndexHandler)))
	r.HandleFunc("/login", loginHandler.HandleLogin)
	r.HandleFunc("/logout", logoutHandler.HandleLogout)
	r.HandleFunc(ciserver.EventHandlerPath+"/{service}", eventHandler.HandleEvent).Methods(http.MethodPost)

	// OAuth2 subrouter.
	OAuth2 := r.PathPrefix("/oauth2").Subrouter()
	OAuth2.HandleFunc("/callback", oauth2Handler.HandleCallback)

	// Repositories subrouter.
	repos := r.PathPrefix("/repositories").Subrouter()
	repos.Use(CSRF)
	repos.Use(middlewares.AuthMiddleware(mongoStore))
	repos.HandleFunc("", repoHandler.HandleRepos)
	repos.HandleFunc("/register", repoHandler.HandleRegisterRepo).Methods(http.MethodPost)
	repos.HandleFunc("/unregister", repoHandler.HandleUnregisterRepo).Methods(http.MethodPost)
	repos.HandleFunc("/activate", repoHandler.HandleActivateRepo).Methods(http.MethodPost)
	repos.HandleFunc("/deactivate", repoHandler.HandleDeactivateRepo).Methods(http.MethodPost)

	// Jobs subrouter.
	jobs := r.PathPrefix(ciserver.JobsPath).Subrouter()
	jobs.Handle("/{id}", middlewares.AuthMiddleware(mongoStore)(http.HandlerFunc(jobHandler.HandleJob)))
	jobs.HandleFunc(ciserver.JobsReportStatusHandlerPath+"/{id}", jobHandler.HandleStatusReport).Methods(http.MethodPost)
	jobs.HandleFunc(ciserver.JobsPublishLogsHandlerPath+"/{id}", jobHandler.HandleLogReport).Methods(http.MethodPost)

	server := &http.Server{
		Addr:         ":" + config.CIServer.Port,
		Handler:      r,
		ReadTimeout:  0,
		WriteTimeout: 0,
		IdleTimeout:  0,
	}
	log.L.Info("Server running on " + server.Addr)
	log.L.Fatal(server.ListenAndServe())
}
