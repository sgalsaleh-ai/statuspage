package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

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

	// Set admin password from config screen if provided
	if adminPass := os.Getenv("ADMIN_PASSWORD"); adminPass != "" {
		if err := db.SetAdminPassword(database, adminPass); err != nil {
			log.Printf("failed to set admin password: %v", err)
		}
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

	// Reverse proxy to Centrifugo for WebSocket connections
	centrifugoWSURL := os.Getenv("CENTRIFUGO_PUBLIC_URL")
	if centrifugoWSURL == "" {
		centrifugoWSURL = "http://localhost:8000"
	}
	cfURL, _ := url.Parse(centrifugoWSURL)
	centrifugoProxy := httputil.NewSingleHostReverseProxy(cfURL)
	mux.HandleFunc("GET /connection/websocket", func(w http.ResponseWriter, r *http.Request) {
		centrifugoProxy.ServeHTTP(w, r)
	})

	// Serve static assets
	fs := http.FileServer(http.Dir(frontendDir))
	mux.Handle("GET /assets/", fs)
	mux.Handle("GET /favicon.ico", fs)

	// SPA fallback: inject config into index.html and serve
	indexBytes, err := os.ReadFile(frontendDir + "/index.html")
	if err != nil {
		log.Fatalf("failed to read index.html: %v", err)
	}
	pageTitle := os.Getenv("PAGE_TITLE")
	if pageTitle == "" {
		pageTitle = "StatusPage"
	}
	pageHeaderText := os.Getenv("PAGE_HEADER_TEXT")
	enableSupportBundle := os.Getenv("ENABLE_SUPPORT_BUNDLE") != "false"
	configScript := fmt.Sprintf(`<script>window.__CONFIG__={pageTitle:%q,pageHeaderText:%q,enableSupportBundle:%t}</script>`, pageTitle, pageHeaderText, enableSupportBundle)
	indexHTML := strings.Replace(string(indexBytes), "</head>", configScript+"</head>", 1)

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(indexHTML))
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
