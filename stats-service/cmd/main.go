package main

import (
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"stats-service/internal/handler"
	"stats-service/internal/repository"
)

func main() {
	connStr := os.Getenv("DB_CONNECTION")
	if connStr == "" {
		connStr = "postgres://user:password@localhost:5432/gymdb?sslmode=disable"
	}

	repo, err := repository.NewPostgresRepo(connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	statsHandler := handler.NewStatsHandler(repo)

	r := mux.NewRouter()
	r.HandleFunc("/health", statsHandler.HealthCheck).Methods("GET")
	r.HandleFunc("/api/stats/club/{clubId}/today", statsHandler.GetTodayVisits).Methods("GET")
	r.HandleFunc("/api/stats/members/top", statsHandler.GetTopMembers).Methods("GET")

	handler := cors.Default().Handler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Stats service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
