package main

import (
	"fmt"
	"log"
	"net/http"
	"qonto_project/producer/internal/config"
	"qonto_project/producer/internal/database"
	"qonto_project/producer/internal/services"
)

func main() {
	handler := http.NewServeMux()
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Could not load configuration: %v", err)
	}

	fmt.Println(cfg.Database.URL)

	db, err := database.ConnectDatabase(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Could not run migrations: %v", err)
	}

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to the homepage!")
	})

	handler.HandleFunc("/account", func(w http.ResponseWriter, r *http.Request) {
		services.CreateAccountHandler(w, r, db)
	})

	handler.HandleFunc("/transaction", func(w http.ResponseWriter, r *http.Request) {
		services.CreateTransactionHandler(w, r, db)
	})

	log.Println("server is running")

	if err := http.ListenAndServe(cfg.Port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
