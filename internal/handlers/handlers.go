package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/sgalsaleh-ai/statuspage/internal/centrifugo"
	"github.com/sgalsaleh-ai/statuspage/internal/email"
	"github.com/sgalsaleh-ai/statuspage/internal/middleware"
	"github.com/sgalsaleh-ai/statuspage/internal/models"
	"github.com/sgalsaleh-ai/statuspage/internal/sdk"
)

type Handler struct {
	db         *sql.DB
	centrifugo *centrifugo.Client
	email      *email.Client
	sdk        *sdk.Client
}

func New(db *sql.DB, cf *centrifugo.Client, em *email.Client, sc *sdk.Client) *Handler {
	return &Handler{db: db, centrifugo: cf, email: em, sdk: sc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Public API
	mux.HandleFunc("GET /api/status", h.GetPublicStatus)
	mux.HandleFunc("GET /api/incidents", h.ListIncidents)
	mux.HandleFunc("GET /api/incidents/{id}", h.GetIncident)
	mux.HandleFunc("POST /api/subscribers", h.CreateSubscriber)

	// Auth
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("POST /api/auth/setup", h.Setup)
	mux.HandleFunc("GET /api/auth/check", h.CheckAuth)

	// Admin API (auth required)
	auth := middleware.RequireAuth
	mux.Handle("GET /api/admin/components", auth(http.HandlerFunc(h.ListComponents)))
	mux.Handle("POST /api/admin/components", auth(http.HandlerFunc(h.CreateComponent)))
	mux.Handle("PUT /api/admin/components/{id}", auth(http.HandlerFunc(h.UpdateComponent)))
	mux.Handle("DELETE /api/admin/components/{id}", auth(http.HandlerFunc(h.DeleteComponent)))
	mux.Handle("GET /api/admin/incidents", auth(http.HandlerFunc(h.ListIncidents)))
	mux.Handle("POST /api/admin/incidents", auth(http.HandlerFunc(h.CreateIncident)))
	mux.Handle("PUT /api/admin/incidents/{id}", auth(http.HandlerFunc(h.UpdateIncident)))
	mux.Handle("DELETE /api/admin/incidents/{id}", auth(http.HandlerFunc(h.DeleteIncident)))
	mux.Handle("POST /api/admin/incidents/{id}/updates", auth(http.HandlerFunc(h.CreateIncidentUpdate)))
	mux.Handle("GET /api/admin/subscribers", auth(http.HandlerFunc(h.ListSubscribers)))
	mux.Handle("DELETE /api/admin/subscribers/{id}", auth(http.HandlerFunc(h.DeleteSubscriber)))

	// Health
	mux.HandleFunc("GET /healthz", h.Health)

	// Centrifugo connection info
	mux.HandleFunc("GET /api/centrifugo/config", h.CentrifugoConfig)

	// SDK proxy endpoints (frontend queries these)
	mux.HandleFunc("GET /api/sdk/license", h.GetLicense)
	mux.HandleFunc("GET /api/sdk/license/field/{name}", h.GetLicenseField)
	mux.HandleFunc("GET /api/sdk/updates", h.GetUpdates)
	mux.Handle("POST /api/admin/support-bundle", auth(http.HandlerFunc(h.GenerateSupportBundle)))
}

// --- Health ---

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	resp := models.HealthResponse{
		Status: "ok",
		Checks: map[string]string{},
	}

	if err := h.db.Ping(); err != nil {
		resp.Status = "degraded"
		resp.Checks["database"] = "error: " + err.Error()
	} else {
		resp.Checks["database"] = "ok"
	}

	if h.centrifugo.Healthy() {
		resp.Checks["centrifugo"] = "ok"
	} else {
		resp.Status = "degraded"
		resp.Checks["centrifugo"] = "unreachable"
	}

	status := http.StatusOK
	if resp.Status != "ok" {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, resp)
}

// --- Auth ---

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	var user models.User
	err := h.db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = $1", req.Username).
		Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	token, err := middleware.GenerateToken(user.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token, "username": user.Username})
}

func (h *Handler) Setup(w http.ResponseWriter, r *http.Request) {
	// Check if admin already has a real password
	var hash string
	err := h.db.QueryRow("SELECT password_hash FROM users WHERE username = 'admin'").Scan(&hash)
	if err == nil && hash != "$2a$10$placeholder" {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "admin already configured"})
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "password required"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to hash password"})
		return
	}

	_, err = h.db.Exec("UPDATE users SET password_hash = $1 WHERE username = 'admin'", string(hashed))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update password"})
		return
	}

	token, err := middleware.GenerateToken("admin")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token, "username": "admin"})
}

func (h *Handler) CheckAuth(w http.ResponseWriter, r *http.Request) {
	// Check if setup is needed
	var hash string
	err := h.db.QueryRow("SELECT password_hash FROM users WHERE username = 'admin'").Scan(&hash)
	needsSetup := err != nil || hash == "$2a$10$placeholder"
	writeJSON(w, http.StatusOK, map[string]bool{"needs_setup": needsSetup})
}

// --- Components ---

func (h *Handler) ListComponents(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, name, description, status, position, group_name, created_at, updated_at FROM components ORDER BY position, id")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list components"})
		return
	}
	defer rows.Close()

	components := []models.Component{}
	for rows.Next() {
		var c models.Component
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Status, &c.Position, &c.GroupName, &c.CreatedAt, &c.UpdatedAt); err != nil {
			continue
		}
		components = append(components, c)
	}
	writeJSON(w, http.StatusOK, components)
}

func (h *Handler) CreateComponent(w http.ResponseWriter, r *http.Request) {
	var c models.Component
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if c.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name required"})
		return
	}
	if c.Status == "" {
		c.Status = "operational"
	}

	err := h.db.QueryRow(
		"INSERT INTO components (name, description, status, position, group_name) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at",
		c.Name, c.Description, c.Status, c.Position, c.GroupName,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create component"})
		return
	}

	h.publishEvent("component:created", c)
	writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) UpdateComponent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var c models.Component
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	err = h.db.QueryRow(
		"UPDATE components SET name = $1, description = $2, status = $3, position = $4, group_name = $5, updated_at = NOW() WHERE id = $6 RETURNING id, name, description, status, position, group_name, created_at, updated_at",
		c.Name, c.Description, c.Status, c.Position, c.GroupName, id,
	).Scan(&c.ID, &c.Name, &c.Description, &c.Status, &c.Position, &c.GroupName, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "component not found"})
		return
	}

	h.publishEvent("component:updated", c)
	writeJSON(w, http.StatusOK, c)
}

func (h *Handler) DeleteComponent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	result, err := h.db.Exec("DELETE FROM components WHERE id = $1", id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete component"})
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "component not found"})
		return
	}

	h.publishEvent("component:deleted", map[string]int{"id": id})
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Incidents ---

func (h *Handler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, title, status, impact, created_at, updated_at FROM incidents ORDER BY created_at DESC")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list incidents"})
		return
	}
	defer rows.Close()

	incidents := []models.Incident{}
	for rows.Next() {
		var i models.Incident
		if err := rows.Scan(&i.ID, &i.Title, &i.Status, &i.Impact, &i.CreatedAt, &i.UpdatedAt); err != nil {
			continue
		}
		incidents = append(incidents, i)
	}
	writeJSON(w, http.StatusOK, incidents)
}

func (h *Handler) GetIncident(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var i models.Incident
	err = h.db.QueryRow("SELECT id, title, status, impact, created_at, updated_at FROM incidents WHERE id = $1", id).
		Scan(&i.ID, &i.Title, &i.Status, &i.Impact, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "incident not found"})
		return
	}

	rows, err := h.db.Query("SELECT id, incident_id, status, message, created_at FROM incident_updates WHERE incident_id = $1 ORDER BY created_at DESC", id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var u models.IncidentUpdate
			if err := rows.Scan(&u.ID, &u.IncidentID, &u.Status, &u.Message, &u.CreatedAt); err != nil {
				continue
			}
			i.Updates = append(i.Updates, u)
		}
	}

	writeJSON(w, http.StatusOK, i)
}

func (h *Handler) GetPublicStatus(w http.ResponseWriter, r *http.Request) {
	// Get all components
	compRows, err := h.db.Query("SELECT id, name, description, status, position, group_name, created_at, updated_at FROM components ORDER BY position, id")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get status"})
		return
	}
	defer compRows.Close()

	components := []models.Component{}
	for compRows.Next() {
		var c models.Component
		if err := compRows.Scan(&c.ID, &c.Name, &c.Description, &c.Status, &c.Position, &c.GroupName, &c.CreatedAt, &c.UpdatedAt); err != nil {
			continue
		}
		components = append(components, c)
	}

	// Get active incidents (not resolved)
	incRows, err := h.db.Query("SELECT id, title, status, impact, created_at, updated_at FROM incidents WHERE status != 'resolved' ORDER BY created_at DESC")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get incidents"})
		return
	}
	defer incRows.Close()

	activeIncidents := []models.Incident{}
	for incRows.Next() {
		var i models.Incident
		if err := incRows.Scan(&i.ID, &i.Title, &i.Status, &i.Impact, &i.CreatedAt, &i.UpdatedAt); err != nil {
			continue
		}
		activeIncidents = append(activeIncidents, i)
	}

	// Overall status
	overallStatus := "operational"
	for _, c := range components {
		if c.Status == "major_outage" {
			overallStatus = "major_outage"
			break
		}
		if c.Status == "degraded" && overallStatus != "major_outage" {
			overallStatus = "degraded"
		}
		if c.Status == "maintenance" && overallStatus == "operational" {
			overallStatus = "maintenance"
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"overall_status":   overallStatus,
		"components":       components,
		"active_incidents": activeIncidents,
	})
}

func (h *Handler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title   string `json:"title"`
		Status  string `json:"status"`
		Impact  string `json:"impact"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "title required"})
		return
	}
	if req.Status == "" {
		req.Status = "investigating"
	}
	if req.Impact == "" {
		req.Impact = "none"
	}

	var i models.Incident
	err := h.db.QueryRow(
		"INSERT INTO incidents (title, status, impact) VALUES ($1, $2, $3) RETURNING id, title, status, impact, created_at, updated_at",
		req.Title, req.Status, req.Impact,
	).Scan(&i.ID, &i.Title, &i.Status, &i.Impact, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create incident"})
		return
	}

	// Create initial update
	if req.Message == "" {
		req.Message = "Incident created"
	}
	var u models.IncidentUpdate
	h.db.QueryRow(
		"INSERT INTO incident_updates (incident_id, status, message) VALUES ($1, $2, $3) RETURNING id, incident_id, status, message, created_at",
		i.ID, req.Status, req.Message,
	).Scan(&u.ID, &u.IncidentID, &u.Status, &u.Message, &u.CreatedAt)
	i.Updates = []models.IncidentUpdate{u}

	h.publishEvent("incident:created", i)
	h.notifySubscribers(i.Title, req.Status, req.Message)
	writeJSON(w, http.StatusCreated, i)
}

func (h *Handler) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var req struct {
		Status string `json:"status"`
		Impact string `json:"impact"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	var i models.Incident
	err = h.db.QueryRow(
		"UPDATE incidents SET status = $1, impact = $2, updated_at = NOW() WHERE id = $3 RETURNING id, title, status, impact, created_at, updated_at",
		req.Status, req.Impact, id,
	).Scan(&i.ID, &i.Title, &i.Status, &i.Impact, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "incident not found"})
		return
	}

	h.publishEvent("incident:updated", i)
	writeJSON(w, http.StatusOK, i)
}

func (h *Handler) DeleteIncident(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	result, err := h.db.Exec("DELETE FROM incidents WHERE id = $1", id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete incident"})
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "incident not found"})
		return
	}

	h.publishEvent("incident:deleted", map[string]int{"id": id})
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) CreateIncidentUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var req struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if req.Message == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message required"})
		return
	}

	// Update incident status if provided
	if req.Status != "" {
		h.db.Exec("UPDATE incidents SET status = $1, updated_at = NOW() WHERE id = $2", req.Status, id)
	}

	var u models.IncidentUpdate
	err = h.db.QueryRow(
		"INSERT INTO incident_updates (incident_id, status, message) VALUES ($1, $2, $3) RETURNING id, incident_id, status, message, created_at",
		id, req.Status, req.Message,
	).Scan(&u.ID, &u.IncidentID, &u.Status, &u.Message, &u.CreatedAt)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create update"})
		return
	}

	// Get incident title for notification
	var title string
	h.db.QueryRow("SELECT title FROM incidents WHERE id = $1", id).Scan(&title)

	h.publishEvent("incident:update:created", u)
	h.notifySubscribers(title, req.Status, req.Message)
	writeJSON(w, http.StatusCreated, u)
}

// --- Subscribers ---

func (h *Handler) CreateSubscriber(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid email required"})
		return
	}

	token := generateToken()
	_, err := h.db.Exec(
		"INSERT INTO subscribers (email, token, verified) VALUES ($1, $2, TRUE) ON CONFLICT (email) DO NOTHING",
		req.Email, token,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to subscribe"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "subscribed"})
}

func (h *Handler) ListSubscribers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, email, verified, created_at FROM subscribers ORDER BY created_at DESC")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list subscribers"})
		return
	}
	defer rows.Close()

	subscribers := []models.Subscriber{}
	for rows.Next() {
		var s models.Subscriber
		if err := rows.Scan(&s.ID, &s.Email, &s.Verified, &s.CreatedAt); err != nil {
			continue
		}
		subscribers = append(subscribers, s)
	}
	writeJSON(w, http.StatusOK, subscribers)
}

func (h *Handler) DeleteSubscriber(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	result, err := h.db.Exec("DELETE FROM subscribers WHERE id = $1", id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete subscriber"})
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "subscriber not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Centrifugo Config ---

func (h *Handler) CentrifugoConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"url": h.centrifugo.PublicURL(),
	})
}

// --- SDK ---

func (h *Handler) GetLicense(w http.ResponseWriter, r *http.Request) {
	info, err := h.sdk.GetLicenseInfo()
	if err != nil {
		log.Printf("sdk license error: %v", err)
		writeJSON(w, http.StatusOK, map[string]any{
			"available": false,
			"error":     err.Error(),
		})
		return
	}

	// Check expiry
	expiresAt := ""
	expired := false
	field, err := h.sdk.GetLicenseField("expires_at")
	if err == nil && field.Value != nil {
		if s, ok := field.Value.(string); ok && s != "" {
			expiresAt = s
			t, err := time.Parse(time.RFC3339, s)
			if err == nil && t.Before(time.Now()) {
				expired = true
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"available":    true,
		"licenseID":    info.LicenseID,
		"customerName": info.CustomerName,
		"licenseType":  info.LicenseType,
		"expiresAt":    expiresAt,
		"expired":      expired,
	})
}

func (h *Handler) GetLicenseField(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	field, err := h.sdk.GetLicenseField(name)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"name":  name,
			"value": nil,
			"error": err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, field)
}

func (h *Handler) GetUpdates(w http.ResponseWriter, r *http.Request) {
	updates, err := h.sdk.GetUpdates()
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"updates":   []any{},
			"available": false,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"updates":   updates,
		"available": len(updates) > 0,
	})
}

// --- Support Bundle ---

func (h *Handler) GenerateSupportBundle(w http.ResponseWriter, r *http.Request) {
	result, err := h.sdk.GenerateSupportBundle()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// --- Helpers ---

func (h *Handler) publishEvent(eventType string, data any) {
	event := map[string]any{
		"type": eventType,
		"data": data,
	}
	if err := h.centrifugo.Publish("status", event); err != nil {
		log.Printf("failed to publish %s event: %v", eventType, err)
	}
}

func (h *Handler) notifySubscribers(incidentTitle, status, message string) {
	if !h.email.Enabled() {
		return
	}

	rows, err := h.db.Query("SELECT email FROM subscribers WHERE verified = TRUE")
	if err != nil {
		log.Printf("failed to query subscribers: %v", err)
		return
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			continue
		}
		emails = append(emails, email)
	}

	if len(emails) == 0 {
		return
	}

	subject := fmt.Sprintf("[StatusPage] %s - %s", incidentTitle, status)
	body := fmt.Sprintf("<h2>%s</h2><p><strong>Status:</strong> %s</p><p>%s</p>", incidentTitle, status, message)

	go func() {
		if err := h.email.SendIncidentNotification(context.Background(), emails, subject, body); err != nil {
			log.Printf("failed to send notifications: %v", err)
		}
	}()
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
