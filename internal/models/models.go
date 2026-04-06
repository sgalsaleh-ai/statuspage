package models

import "time"

type Component struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // operational, degraded, major_outage, maintenance
	Position    int       `json:"position"`
	GroupName   string    `json:"group_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Incident struct {
	ID        int              `json:"id"`
	Title     string           `json:"title"`
	Status    string           `json:"status"` // investigating, identified, monitoring, resolved
	Impact    string           `json:"impact"` // none, minor, major, critical
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Updates   []IncidentUpdate `json:"updates,omitempty"`
}

type IncidentUpdate struct {
	ID         int       `json:"id"`
	IncidentID int       `json:"incident_id"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

type Subscriber struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Verified  bool      `json:"verified"`
	Token     string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}
