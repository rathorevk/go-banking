package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rathorevk/GoBanking/app/api"
	"github.com/rathorevk/GoBanking/app/database"
	"github.com/rathorevk/GoBanking/app/middleware"
)

func StartServer() {
	// Initialize database connection
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Run database migrations
	if err = database.RunMigrations(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize DB
	if _, err = database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Get server address and port from environment variables
	address := os.Getenv("SERVER_ADDRESS")
	port := os.Getenv("SERVER_PORT")

	// Create a new router
	router := mux.NewRouter()

	// Apply middleware
	router.Use(middleware.PanicHandler)
	router.Use(middleware.LoggingMiddleware)

	// Define routes
	router.HandleFunc("/user", api.CreateUserHandler).Methods("POST")
	router.HandleFunc("/user/{userId}", api.GetUserHandler).Methods("GET")
	router.HandleFunc("/user/{userId}/balance", api.GetBalanceHandler).Methods("GET")

	// transaction route with Source header validation
	tx_router := router.PathPrefix("/user/{userId}/transaction").Subrouter()
	tx_router.Use(middleware.SourceHeaderMatcher)
	tx_router.HandleFunc("", api.CreateTransactionHandler).Methods("POST")

	srv := &http.Server{
		Addr: fmt.Sprintf("%s:%v", address, port),
		// Set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router, // Pass our instance of gorilla/mux in.
	}

	log.Println("Server listening on:", fmt.Sprintf("%s:%v", address, port))

	log.Fatal(srv.ListenAndServe())
}
