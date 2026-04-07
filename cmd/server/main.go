package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sgalsaleh-ai/statuspage/internal/centrifugo"
	"github.com/sgalsaleh-ai/statuspage/internal/db"
	"github.com/sgalsaleh-ai/statuspage/internal/email"
	"github.com/sgalsaleh-ai/statuspage/internal/handlers"
	"github.com/sgalsaleh-ai/statuspage/internal/sdk"
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
	sc := sdk.New()

	// Start metrics reporter
	sc.StartMetricsReporter(func() map[string]any {
		var componentCount, incidentCount, subscriberCount int
		database.QueryRow("SELECT COUNT(*) FROM components").Scan(&componentCount)
		database.QueryRow("SELECT COUNT(*) FROM incidents").Scan(&incidentCount)
		database.QueryRow("SELECT COUNT(*) FROM subscribers").Scan(&subscriberCount)

		activeIncidents := 0
		database.QueryRow("SELECT COUNT(*) FROM incidents WHERE status != 'resolved'").Scan(&activeIncidents)

		return map[string]any{
			"component_count":       componentCount,
			"incident_count":        incidentCount,
			"active_incident_count": activeIncidents,
			"subscriber_count":      subscriberCount,
		}
	})

	// Set instance tags
	go func() {
		sc.SetInstanceTags(map[string]string{
			"name": "StatusPage Production",
			"env":  "production",
		})
	}()

	h := handlers.New(database, cf, em, sc)

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
