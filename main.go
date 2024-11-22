package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/mbsoft31/screens/client_manager"
	"github.com/mbsoft31/screens/db"
	"github.com/mbsoft31/screens/handlers"
	"github.com/mbsoft31/screens/models"
)

// Middleware to inject the database into the request context
func dbMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "db", db.DB)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Simple logger middleware
func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize the database and seed data
	db.Init()
	if err := models.Seed(db.DB); err != nil {
		log.Fatal("Cannot seed database")
	}

	// Start the TCP server for clients
	manager, err := client_manager.NewManager(":8081")
	if err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}

	fmt.Println("Starting the TCP server...")
	go manager.Listen()
	//go client_manager.StartTCPServer()

	// Define HTTP handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/screens", handlers.ScreensHandler)
	mux.HandleFunc("/playlists", handlers.PlaylistsHandler)
	mux.HandleFunc("/users", handlers.UsersHandler)
	mux.HandleFunc("/ws", client_manager.HandleWebSocket)

	// Apply middleware
	handler := loggerMiddleware(dbMiddleware(mux))

	// Start the HTTP server
	log.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
