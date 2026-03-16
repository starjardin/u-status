package models

import "time"

type Plan string

const (
	PlanFree   Plan = "free"
	PlanPro    Plan = "pro"
	PlanAgency Plan = "agency"
)

var PlanLimits = map[Plan]int{
	PlanFree:   3,
	PlanPro:    20,
	PlanAgency: -1, // unlimited
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Plan         Plan      `json:"plan"`
	CreatedAt    time.Time `json:"created_at"`
}

type Monitor struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"user_id"`
	Name                 string    `json:"name"`
	URL                  string    `json:"url"`
	IntervalSeconds      int       `json:"interval_seconds"`
	Status               string    `json:"status"` // "up" | "down" | "pending"
	AlertEmail           bool      `json:"alert_email"`
	IsPublic             bool      `json:"is_public"`
	ConsecutiveFailures  int       `json:"consecutive_failures"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type Check struct {
	ID             string    `json:"id"`
	MonitorID      string    `json:"monitor_id"`
	StatusCode     *int      `json:"status_code"`
	ResponseTimeMs *int      `json:"response_time_ms"`
	IsUp           bool      `json:"is_up"`
	Error          *string   `json:"error,omitempty"`
	CheckedAt      time.Time `json:"checked_at"`
}

type Incident struct {
	ID         string     `json:"id"`
	MonitorID  string     `json:"monitor_id"`
	StartedAt  time.Time  `json:"started_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	Error      *string    `json:"error,omitempty"`
}

type MonitorStats struct {
	Uptime1d  float64 `json:"uptime_1d"`
	Uptime7d  float64 `json:"uptime_7d"`
	Uptime30d float64 `json:"uptime_30d"`
}
