package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	ciserver "github.com/shark-ci/shark-ci/ci-server"
	"github.com/shark-ci/shark-ci/ci-server/configs"
	"github.com/shark-ci/shark-ci/ci-server/handlers"
	"github.com/shark-ci/shark-ci/ci-server/middlewares"
	"github.com/shark-ci/shark-ci/ci-server/services"
	"github.com/shark-ci/shark-ci/ci-server/sessions"
	"github.com/shark-ci/shark-ci/ci-server/store"
	"github.com/shark-ci/shark-ci/config"
	"github.com/shark-ci/shark-ci/message_queue"
)

func initGitServices(store store.Storer, config config.CIServerConfig) services.ServiceMap {
	serviceMap := services.ServiceMap{}
	if config.GitHubClientID != "" {
		ghm := services.NewGitHubManager(config.GitHubClientID, config.GitHubClientSecret, store, config)
		serviceMap[ghm.ServiceName()] = ghm
	}
	return serviceMap
}

func main() {
	// TODO: Remove godotenv.
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}

	config, err := config.NewCIServerConfigFromEnv()
	if err != nil {
		log.Fatalln(err)
	}

	sessions.InitSessionStore(config.SecretKey)

	// TODO: Handle templates better.
	configs.LoadTemplates()

	mongoStore, err := store.NewMongoStore(config.MongoURI)
	if err != nil {
		log.Fatalln(err)
	}
	defer mongoStore.Close(context.TODO())

	rabbitMQ, err := message_queue.NewRabbitMQ(config.RabbitMQURI)
	if err != nil {
		log.Fatalln(err)
	}
	defer rabbitMQ.Close(context.TODO())

	serviceMap := initGitServices(mongoStore, config)

	CSRF := csrf.Protect([]byte(config.SecretKey))

	loginHandler := handlers.NewLoginHandler(mongoStore, serviceMap)
	logoutHandler := handlers.NewLogoutHandler()
	eventHandler := handlers.NewEventHandler(mongoStore, rabbitMQ, serviceMap)
	oauth2Handler := handlers.NewOAuth2Handler(mongoStore, serviceMap)
	repoHandler := handlers.NewRepoHandler(mongoStore, serviceMap)
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
		Addr:    ":" + config.Port,
		Handler: r,
		//ReadTimeout:  15 * time.Second,
		//WriteTimeout: 15 * time.Second,
		//IdleTimeout:  60 * time.Second,
	}
	log.Println("Server running on " + server.Addr)
	log.Fatalln(server.ListenAndServe())
}
