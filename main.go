package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/csrf"
	gorilla_handlers "github.com/gorilla/handlers"
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

// TODO: Change to postgres
func initDatabase() {
	var err error
	db.DB, err = gorm.Open(sqlite.Open("db.sqlite3"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}

	db.DB.AutoMigrate(&models.User{}, &models.OAuth2Token{}, &models.Webhook{})
}

func initGitServices() {
	GitHubService := os.Getenv("GITHUB_SERVICE") == "true"
	GitLabService := os.Getenv("GITLAB_SERVICE") == "true"
	if !GitHubService && !GitLabService {
		log.Fatal("error: at least one service (*_SERVICE) has to be set as `true`")
	}

	if GitHubService {
		clientID := os.Getenv("GITHUB_CLIENT_ID")
		clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
		services.NewGitHubOAuth2Config(clientID, clientSecret)
	}
}

func initTemplates() {
	configs.LoadTemplates()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	initTemplates()
	initDatabase()
	initGitServices()

	CSRF := csrf.Protect([]byte(os.Getenv("CSRF_KEY")))

	r := mux.NewRouter()
	r.Handle("/", middlewares.AuthMiddleware(http.HandlerFunc(handlers.Index)))
	r.HandleFunc("/login", handlers.Login)
	r.HandleFunc("/logout", handlers.Logout)

	sOAuth2 := r.PathPrefix("/oauth2").Subrouter()
	sOAuth2.HandleFunc("/callback", handlers.OAuth2Callback)

	sRepos := r.PathPrefix("/repositories").Subrouter()
	sRepos.Use(middlewares.AuthMiddleware)
	sRepos.HandleFunc("", handlers.Repos)
	sRepos.HandleFunc("/register", handlers.ReposRegister).Methods(http.MethodPost)
	sRepos.HandleFunc("/unregister", handlers.ReposUnregister).Methods(http.MethodPost)
	sRepos.HandleFunc("/activate", handlers.ReposActivate).Methods(http.MethodPost)
	sRepos.HandleFunc("/deactivate", handlers.ReposDeactivate).Methods(http.MethodPost)

	handler := gorilla_handlers.CombinedLoggingHandler(os.Stdout, r)
	handler = CSRF(handler)

	server := &http.Server{
		Addr:         ":" + os.Getenv("PORT"),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	log.Println("Server running on " + server.Addr)
	log.Fatal(server.ListenAndServe())
}
