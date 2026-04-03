package db

import (
	"database/sql"
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
		IsAdmin:      false,
		CreatedAt:    time.Now().UTC(),
	}
	_, err := db.Exec(
		`INSERT INTO users (id, email, password_hash, plan, is_admin, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		u.ID, u.Email, u.PasswordHash, u.Plan, u.IsAdmin, u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {
	u := &models.User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, plan, is_admin, created_at FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Plan, &u.IsAdmin, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func GetUserByID(db *sql.DB, id string) (*models.User, error) {
	u := &models.User{}
	err := db.QueryRow(
		`SELECT id, email, password_hash, plan, is_admin, created_at FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Plan, &u.IsAdmin, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func ListAllUsers(db *sql.DB) ([]*models.User, error) {
	rows, err := db.Query(
		`SELECT id, email, password_hash, plan, is_admin, created_at FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Plan, &u.IsAdmin, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// ---- Monitors ----

func CountMonitorsByUser(db *sql.DB, userID string) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM monitors WHERE user_id = $1`, userID).Scan(&count)
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
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, $9, $10)`,
		m.ID, m.UserID, m.Name, m.URL, m.IntervalSeconds, m.Status, m.AlertEmail, m.IsPublic, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func ListMonitorsByUser(db *sql.DB, userID string) ([]*models.Monitor, error) {
	rows, err := db.Query(
		`SELECT id, user_id, name, url, interval_seconds, status, alert_email, is_public, consecutive_failures, created_at, updated_at
		 FROM monitors WHERE user_id = $1 ORDER BY created_at DESC`, userID,
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
		 FROM monitors WHERE id = $1`, id,
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
		`UPDATE monitors SET name=$1, url=$2, interval_seconds=$3, alert_email=$4, is_public=$5, updated_at=$6
		 WHERE id=$7 AND user_id=$8`,
		name, url, intervalSeconds, alertEmail, isPublic, time.Now().UTC(), id, userID,
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
	res, err := db.Exec(`DELETE FROM monitors WHERE id=$1 AND user_id=$2`, id, userID)
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
		`UPDATE monitors SET status=$1, consecutive_failures=$2, updated_at=$3 WHERE id=$4`,
		status, consecutiveFailures, time.Now().UTC(), id,
	)
	return err
}

// ---- Checks ----

func CreateCheck(db *sql.DB, monitorID string, statusCode *int, responseTimeMs *int, isUp bool, errStr *string) error {
	_, err := db.Exec(
		`INSERT INTO checks (id, monitor_id, status_code, response_time_ms, is_up, error, checked_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		uuid.NewString(), monitorID, statusCode, responseTimeMs, isUp, errStr, time.Now().UTC(),
	)
	return err
}

func ListChecks(db *sql.DB, monitorID string, hours int) ([]*models.Check, error) {
	rows, err := db.Query(
		`SELECT id, monitor_id, status_code, response_time_ms, is_up, error, checked_at
		 FROM checks WHERE monitor_id = $1 AND checked_at >= NOW() - make_interval(hours => $2)
		 ORDER BY checked_at DESC LIMIT 500`,
		monitorID, hours,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []*models.Check
	for rows.Next() {
		c := &models.Check{}
		if err := rows.Scan(&c.ID, &c.MonitorID, &c.StatusCode, &c.ResponseTimeMs, &c.IsUp, &c.Error, &c.CheckedAt); err != nil {
			return nil, err
		}
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
		var total int
		var up sql.NullInt64
		err := db.QueryRow(
			`SELECT COUNT(*), SUM(CASE WHEN is_up THEN 1 ELSE 0 END)
			 FROM checks WHERE monitor_id=$1 AND checked_at >= NOW() - make_interval(hours => $2)`,
			monitorID, period.hours,
		).Scan(&total, &up)
		if err != nil {
			return nil, err
		}
		if total > 0 {
			*period.dest = float64(up.Int64) / float64(total) * 100
		}
	}
	return stats, nil
}

func PurgeOldChecks(db *sql.DB) error {
	_, err := db.Exec(`DELETE FROM checks WHERE checked_at < NOW() - INTERVAL '90 days'`)
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
		`INSERT INTO incidents (id, monitor_id, started_at, error) VALUES ($1, $2, $3, $4)`,
		inc.ID, inc.MonitorID, inc.StartedAt, inc.Error,
	)
	if err != nil {
		return nil, err
	}
	return inc, nil
}

func ResolveIncident(db *sql.DB, monitorID string) error {
	_, err := db.Exec(
		`UPDATE incidents SET resolved_at=$1 WHERE monitor_id=$2 AND resolved_at IS NULL`,
		time.Now().UTC(), monitorID,
	)
	return err
}

func ListIncidents(db *sql.DB, monitorID string) ([]*models.Incident, error) {
	rows, err := db.Query(
		`SELECT id, monitor_id, started_at, resolved_at, error FROM incidents
		 WHERE monitor_id=$1 ORDER BY started_at DESC LIMIT 50`, monitorID,
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

func scanMonitor(row *sql.Row) (*models.Monitor, error) {
	m := &models.Monitor{}
	err := row.Scan(
		&m.ID, &m.UserID, &m.Name, &m.URL, &m.IntervalSeconds, &m.Status,
		&m.AlertEmail, &m.IsPublic, &m.ConsecutiveFailures, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func scanMonitors(rows *sql.Rows) ([]*models.Monitor, error) {
	var monitors []*models.Monitor
	for rows.Next() {
		m := &models.Monitor{}
		if err := rows.Scan(
			&m.ID, &m.UserID, &m.Name, &m.URL, &m.IntervalSeconds, &m.Status,
			&m.AlertEmail, &m.IsPublic, &m.ConsecutiveFailures, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		monitors = append(monitors, m)
	}
	return monitors, rows.Err()
}
