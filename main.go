package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/FilipSolich/ci-server/configs"
	"github.com/FilipSolich/ci-server/db"
	"github.com/FilipSolich/ci-server/handlers"
	"github.com/FilipSolich/ci-server/middlewares"
	"github.com/FilipSolich/ci-server/models"
	"github.com/FilipSolich/ci-server/services"
)

func initDatabase() {
	var err error
	db.DB, err = gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}

	db.DB.AutoMigrate(
		&models.User{},
		&models.UserIdentity{},
		&models.OAuth2Token{},
		&models.OAuth2State{},
		&models.Repository{},
		&models.Webhook{},
		&models.Job{},
	)
}

func initGitServices() {
	if configs.GitHubEnabled {
		services.NewGitHubManager(configs.GitHubClientID, configs.GitHubClientSecret)
		services.Services[services.GitHub.GetServiceName()] = &services.GitHub
	}

	// TODO: Add GitLab service.
}

func initTemplates() {
	configs.LoadTemplates()
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

	initTemplates()

	initDatabase() // TODO: Delete
	disconnect, err := db.InitDatabase(
		configs.MongoURI,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer disconnect(context.Background())

	initGitServices()
	//messageQueue, err := mq.NewMQ(
	//	configs.RabbitMQHost,
	//	configs.RabbitMQPort,
	//	configs.RabbitMQUsername,
	//	configs.RabbitMQPassword,
	//)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//mq.MQ = messageQueue
	//defer messageQueue.Close()

	CSRF := csrf.Protect([]byte(configs.CSRFSecret))

	r := mux.NewRouter()
	r.Use(middlewares.LoggingMiddleware)
	r.Handle("/", middlewares.AuthMiddleware(http.HandlerFunc(handlers.IndexHandler)))
	r.HandleFunc("/login", handlers.LoginHandler)
	r.HandleFunc("/logout", handlers.LogoutHandler)
	r.HandleFunc(configs.EventHandlerPath+"/{service}", handlers.EventHandler).Methods(http.MethodPost)

	sOAuth2 := r.PathPrefix("/oauth2").Subrouter()
	sOAuth2.HandleFunc("/callback", handlers.OAuth2CallbackHandler)

	sRepos := r.PathPrefix("/repositories").Subrouter()
	sRepos.Use(CSRF)
	sRepos.Use(middlewares.AuthMiddleware)
	sRepos.HandleFunc("", handlers.ReposHandler)
	//sRepos.HandleFunc("/register", handlers.ReposRegisterHandler).Methods(http.MethodPost)
	//sRepos.HandleFunc("/unregister", handlers.ReposUnregisterHandler).Methods(http.MethodPost)
	//sRepos.HandleFunc("/activate", handlers.ReposActivateHandler).Methods(http.MethodPost)
	//sRepos.HandleFunc("/deactivate", handlers.ReposDeactivateHandler).Methods(http.MethodPost)

	server := &http.Server{
		Addr:         ":" + configs.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Println("Server running on " + server.Addr)
	log.Fatal(server.ListenAndServe())
}
