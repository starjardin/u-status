package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/user/u-status/internal/models"
)

// ---- Users ----

func CreateUser(db *sql.DB, email, passwordHash string) (*models.User, error) {
	u := &models.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: passwordHash,
		Plan:         models.PlanFree,
		CreatedAt:    time.Now().UTC(),
	}
	_, err := db.Exec(
		`INSERT INTO users (id, email, password_hash, plan, created_at) VALUES (?, ?, ?, ?, ?)`,
		u.ID, u.Email, u.PasswordHash, u.Plan, u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {
	u := &models.User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, plan, created_at FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Plan, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func GetUserByID(db *sql.DB, id string) (*models.User, error) {
	u := &models.User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, plan, created_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Plan, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// ---- Monitors ----

func CountMonitorsByUser(db *sql.DB, userID string) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM monitors WHERE user_id = ?`, userID).Scan(&count)
	return count, err
}

func CreateMonitor(db *sql.DB, userID, name, url string, intervalSeconds int) (*models.Monitor, error) {
	m := &models.Monitor{
		ID:              uuid.NewString(),
		UserID:          userID,
		Name:            name,
		URL:             url,
		IntervalSeconds: intervalSeconds,
		Status:          "pending",
		AlertEmail:      true,
		IsPublic:        false,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	_, err := db.Exec(
		`INSERT INTO monitors (id, user_id, name, url, interval_seconds, status, alert_email, is_public, consecutive_failures, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		m.ID, m.UserID, m.Name, m.URL, m.IntervalSeconds, m.Status, boolToInt(m.AlertEmail), boolToInt(m.IsPublic), m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func ListMonitorsByUser(db *sql.DB, userID string) ([]*models.Monitor, error) {
	rows, err := db.Query(
		`SELECT id, user_id, name, url, interval_seconds, status, alert_email, is_public, consecutive_failures, created_at, updated_at
		 FROM monitors WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMonitors(rows)
}

func GetMonitor(db *sql.DB, id string) (*models.Monitor, error) {
	row := db.QueryRow(
		`SELECT id, user_id, name, url, interval_seconds, status, alert_email, is_public, consecutive_failures, created_at, updated_at
		 FROM monitors WHERE id = ?`, id,
	)
	return scanMonitor(row)
}

func GetAllActiveMonitors(db *sql.DB) ([]*models.Monitor, error) {
	rows, err := db.Query(
		`SELECT id, user_id, name, url, interval_seconds, status, alert_email, is_public, consecutive_failures, created_at, updated_at
		 FROM monitors ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMonitors(rows)
}

func UpdateMonitor(db *sql.DB, id, userID, name, url string, intervalSeconds int, alertEmail, isPublic bool) error {
	res, err := db.Exec(
		`UPDATE monitors SET name=?, url=?, interval_seconds=?, alert_email=?, is_public=?, updated_at=?
		 WHERE id=? AND user_id=?`,
		name, url, intervalSeconds, boolToInt(alertEmail), boolToInt(isPublic), time.Now().UTC(), id, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func DeleteMonitor(db *sql.DB, id, userID string) error {
	res, err := db.Exec(`DELETE FROM monitors WHERE id=? AND user_id=?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func UpdateMonitorStatus(db *sql.DB, id, status string, consecutiveFailures int) error {
	_, err := db.Exec(
		`UPDATE monitors SET status=?, consecutive_failures=?, updated_at=? WHERE id=?`,
		status, consecutiveFailures, time.Now().UTC(), id,
	)
	return err
}

// ---- Checks ----

func CreateCheck(db *sql.DB, monitorID string, statusCode *int, responseTimeMs *int, isUp bool, errStr *string) error {
	_, err := db.Exec(
		`INSERT INTO checks (id, monitor_id, status_code, response_time_ms, is_up, error, checked_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		uuid.NewString(), monitorID, statusCode, responseTimeMs, boolToInt(isUp), errStr, time.Now().UTC(),
	)
	return err
}

func ListChecks(db *sql.DB, monitorID string, hours int) ([]*models.Check, error) {
	rows, err := db.Query(
		`SELECT id, monitor_id, status_code, response_time_ms, is_up, error, checked_at
		 FROM checks WHERE monitor_id = ? AND checked_at >= datetime('now', ? || ' hours')
		 ORDER BY checked_at DESC LIMIT 500`,
		monitorID, fmt.Sprintf("-%d", hours),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []*models.Check
	for rows.Next() {
		c := &models.Check{}
		var isUpInt int
		if err := rows.Scan(&c.ID, &c.MonitorID, &c.StatusCode, &c.ResponseTimeMs, &isUpInt, &c.Error, &c.CheckedAt); err != nil {
			return nil, err
		}
		c.IsUp = isUpInt == 1
		checks = append(checks, c)
	}
	return checks, rows.Err()
}

func GetMonitorStats(db *sql.DB, monitorID string) (*models.MonitorStats, error) {
	stats := &models.MonitorStats{}
	for _, period := range []struct {
		hours int
		dest  *float64
	}{
		{24, &stats.Uptime1d},
		{168, &stats.Uptime7d},
		{720, &stats.Uptime30d},
	} {
		var total, up int
		err := db.QueryRow(
			`SELECT COUNT(*), SUM(CASE WHEN is_up=1 THEN 1 ELSE 0 END)
			 FROM checks WHERE monitor_id=? AND checked_at >= datetime('now', ? || ' hours')`,
			monitorID, fmt.Sprintf("-%d", period.hours),
		).Scan(&total, &up)
		if err != nil {
			return nil, err
		}
		if total > 0 {
			*period.dest = float64(up) / float64(total) * 100
		}
	}
	return stats, nil
}

func PurgeOldChecks(db *sql.DB) error {
	_, err := db.Exec(`DELETE FROM checks WHERE checked_at < datetime('now', '-90 days')`)
	return err
}

// ---- Incidents ----

func CreateIncident(db *sql.DB, monitorID, errStr string) (*models.Incident, error) {
	inc := &models.Incident{
		ID:        uuid.NewString(),
		MonitorID: monitorID,
		StartedAt: time.Now().UTC(),
		Error:     &errStr,
	}
	_, err := db.Exec(
		`INSERT INTO incidents (id, monitor_id, started_at, error) VALUES (?, ?, ?, ?)`,
		inc.ID, inc.MonitorID, inc.StartedAt, inc.Error,
	)
	if err != nil {
		return nil, err
	}
	return inc, nil
}

func ResolveIncident(db *sql.DB, monitorID string) error {
	_, err := db.Exec(
		`UPDATE incidents SET resolved_at=? WHERE monitor_id=? AND resolved_at IS NULL`,
		time.Now().UTC(), monitorID,
	)
	return err
}

func ListIncidents(db *sql.DB, monitorID string) ([]*models.Incident, error) {
	rows, err := db.Query(
		`SELECT id, monitor_id, started_at, resolved_at, error FROM incidents
		 WHERE monitor_id=? ORDER BY started_at DESC LIMIT 50`, monitorID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*models.Incident
	for rows.Next() {
		inc := &models.Incident{}
		if err := rows.Scan(&inc.ID, &inc.MonitorID, &inc.StartedAt, &inc.ResolvedAt, &inc.Error); err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}
	return incidents, rows.Err()
}

// ---- helpers ----

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func scanMonitor(row *sql.Row) (*models.Monitor, error) {
	m := &models.Monitor{}
	var alertEmailInt, isPublicInt int
	err := row.Scan(
		&m.ID, &m.UserID, &m.Name, &m.URL, &m.IntervalSeconds, &m.Status,
		&alertEmailInt, &isPublicInt, &m.ConsecutiveFailures, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	m.AlertEmail = alertEmailInt == 1
	m.IsPublic = isPublicInt == 1
	return m, nil
}

func scanMonitors(rows *sql.Rows) ([]*models.Monitor, error) {
	var monitors []*models.Monitor
	for rows.Next() {
		m := &models.Monitor{}
		var alertEmailInt, isPublicInt int
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.Name, &m.URL, &m.IntervalSeconds, &m.Status,
			&alertEmailInt, &isPublicInt, &m.ConsecutiveFailures, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		m.AlertEmail = alertEmailInt == 1
		m.IsPublic = isPublicInt == 1
		monitors = append(monitors, m)
	}
	return monitors, rows.Err()
}
