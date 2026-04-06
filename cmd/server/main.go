package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sgalsaleh-ai/statuspage/internal/centrifugo"
	"github.com/sgalsaleh-ai/statuspage/internal/db"
	"github.com/sgalsaleh-ai/statuspage/internal/email"
	"github.com/sgalsaleh-ai/statuspage/internal/handlers"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if err := db.WaitForDB(database, 30); err != nil {
		log.Fatalf("database not ready: %v", err)
	}

	if err := db.Migrate(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	cf := centrifugo.New()
	em := email.New()

	h := handlers.New(database, cf, em)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Serve React frontend static files
	frontendDir := os.Getenv("FRONTEND_DIR")
	if frontendDir == "" {
		frontendDir = "./frontend/dist"
	}

	// Serve static assets
	fs := http.FileServer(http.Dir(frontendDir))
	mux.Handle("GET /assets/", fs)
	mux.Handle("GET /favicon.ico", fs)

	// SPA fallback: serve index.html for all non-API, non-asset routes
	indexPath := frontendDir + "/index.html"
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, indexPath)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
