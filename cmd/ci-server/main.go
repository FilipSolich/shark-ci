package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/shark-ci/shark-ci/ci-server/configs"
	"github.com/shark-ci/shark-ci/ci-server/handlers"
	"github.com/shark-ci/shark-ci/ci-server/middlewares"
	"github.com/shark-ci/shark-ci/ci-server/services"
	"github.com/shark-ci/shark-ci/ci-server/store"
	"github.com/shark-ci/shark-ci/mq"
)

func initGitServices(store store.Storer) services.ServiceMap {
	serviceMap := services.ServiceMap{}
	if configs.GitHubEnabled {
		ghm := services.NewGitHubManager(configs.GitHubClientID, configs.GitHubClientSecret, store)
		serviceMap[ghm.ServiceName()] = ghm
	}
	return serviceMap
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	err = configs.LoadEnv()
	if err != nil {
		log.Fatal(err)
	}

	configs.LoadTemplates()

	mongoStore, err := store.NewMongoStore(configs.MongoURI)
	if err != nil {
		log.Fatal(err)
	}
	defer mongoStore.Close(context.TODO())

	serviceMap := initGitServices(mongoStore)

	closeMQ, err := mq.InitMQ(configs.RabbitMQHost, configs.RabbitMQPort, configs.RabbitMQUsername, configs.RabbitMQPassword)
	if err != nil {
		log.Fatal(err)
	}
	defer closeMQ()

	CSRF := csrf.Protect([]byte(configs.CSRFSecret))

	loginHandler := handlers.NewLoginHandler(mongoStore, serviceMap)
	logoutHandler := handlers.NewLogoutHandler()
	eventHandler := handlers.NewEventHandler(mongoStore, serviceMap)
	oauth2Handler := handlers.NewOAuth2Handler(mongoStore, serviceMap)
	repoHandler := handlers.NewRepoHandler(mongoStore, serviceMap)
	jobHandler := handlers.NewJobHandler(mongoStore, serviceMap)

	r := mux.NewRouter()
	r.Use(middlewares.LoggingMiddleware)
	r.Handle("/", middlewares.AuthMiddleware(mongoStore)(http.HandlerFunc(handlers.IndexHandler)))
	r.HandleFunc("/login", loginHandler.HandleLogin)
	r.HandleFunc("/logout", logoutHandler.HandleLogout)
	r.HandleFunc(configs.EventHandlerPath+"/{service}", eventHandler.HandleEvent).Methods(http.MethodPost)

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
	jobs := r.PathPrefix(configs.JobsPath).Subrouter()
	jobs.Handle("/{id}", middlewares.AuthMiddleware(mongoStore)(http.HandlerFunc(jobHandler.HandleJob)))
	jobs.HandleFunc(configs.JobsReportStatusHandlerPath+"/{id}", jobHandler.HandleStatusReport).Methods(http.MethodPost)
	jobs.HandleFunc(configs.JobsPublishLogsHandlerPath+"/{id}", jobHandler.HandleLogReport).Methods(http.MethodPost)

	server := &http.Server{
		Addr:    ":" + configs.Port,
		Handler: r,
		//ReadTimeout:  15 * time.Second,
		//WriteTimeout: 15 * time.Second,
		//IdleTimeout:  60 * time.Second,
	}
	log.Println("Server running on " + server.Addr)
	log.Fatal(server.ListenAndServe())
}
