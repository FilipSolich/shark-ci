package main

import (
	"log"
	"net/http"
	"os"
	"time"

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
	configs.LoadTemplates("templates/*.html")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	initTemplates()
	initDatabase()
	initGitServices()

	r := mux.NewRouter()
	r.HandleFunc("/", middlewares.AuthMiddleware(handlers.IndexHandler))
	r.HandleFunc("/login", handlers.LoginHandler)
	r.HandleFunc("/logout", handlers.LogoutHandler)

	s := r.PathPrefix("/oauth2").Subrouter()
	s.HandleFunc("/callback", handlers.OAuth2CallbackHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	log.Println("Server running on " + server.Addr)
	log.Fatal(server.ListenAndServe())
}
