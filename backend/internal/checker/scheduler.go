package checker

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	appdb "github.com/user/u-status/internal/db"
	"github.com/user/u-status/internal/models"
)

type Scheduler struct {
	db       *sql.DB
	mu       sync.Mutex
	monitors map[string]context.CancelFunc // monitorID -> cancel
	notifyCh <-chan string                  // receives new monitor IDs from API
	deleteCh <-chan string                  // receives deleted monitor IDs from API
}

func NewScheduler(db *sql.DB, notifyCh <-chan string, deleteCh <-chan string) *Scheduler {
	return &Scheduler{
		db:       db,
		monitors: make(map[string]context.CancelFunc),
		notifyCh: notifyCh,
		deleteCh: deleteCh,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	// Load all existing monitors on startup
	monitors, err := appdb.GetAllActiveMonitors(s.db)
	if err != nil {
		log.Printf("scheduler: failed to load monitors: %v", err)
	}
	for _, m := range monitors {
		s.startMonitor(ctx, m)
	}

	log.Printf("scheduler: started %d monitors", len(monitors))

	// Watch for new/deleted monitors from API handlers
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case id := <-s.notifyCh:
				m, err := appdb.GetMonitor(s.db, id)
				if err != nil {
					log.Printf("scheduler: could not load new monitor %s: %v", id, err)
					continue
				}
				s.startMonitor(ctx, m)
				log.Printf("scheduler: started new monitor %s (%s)", m.ID, m.URL)
			case id := <-s.deleteCh:
				s.stopMonitor(id)
				log.Printf("scheduler: stopped monitor %s", id)
			}
		}
	}()

	// Periodic cleanup of old checks
	go func() {
		t := time.NewTicker(24 * time.Hour)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := appdb.PurgeOldChecks(s.db); err != nil {
					log.Printf("scheduler: purge error: %v", err)
				}
			}
		}
	}()
}

func (s *Scheduler) startMonitor(parentCtx context.Context, m *models.Monitor) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop existing goroutine if any (e.g. restart after update)
	if cancel, ok := s.monitors[m.ID]; ok {
		cancel()
	}

	ctx, cancel := context.WithCancel(parentCtx)
	s.monitors[m.ID] = cancel

	go s.runMonitor(ctx, m)
}

func (s *Scheduler) stopMonitor(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cancel, ok := s.monitors[id]; ok {
		cancel()
		delete(s.monitors, id)
	}
}

func (s *Scheduler) runMonitor(ctx context.Context, m *models.Monitor) {
	interval := time.Duration(m.IntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run an immediate check on start
	s.doCheck(m.ID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.doCheck(m.ID)
		}
	}
}

func (s *Scheduler) doCheck(monitorID string) {
	// Reload monitor to get latest URL and settings
	m, err := appdb.GetMonitor(s.db, monitorID)
	if err != nil {
		log.Printf("checker: monitor %s not found, stopping", monitorID)
		return
	}

	result := CheckURL(m.URL)

	// Persist check
	if err := appdb.CreateCheck(s.db, m.ID, result.StatusCode, result.ResponseTimeMs, result.IsUp, result.Error); err != nil {
		log.Printf("checker: failed to save check for %s: %v", m.ID, err)
	}

	// Incident state machine
	s.updateIncidentState(m, result)
}

func (s *Scheduler) updateIncidentState(m *models.Monitor, result CheckResult) {
	if result.IsUp {
		if m.Status == "down" {
			// Recovery
			if err := appdb.ResolveIncident(s.db, m.ID); err != nil {
				log.Printf("checker: failed to resolve incident for %s: %v", m.ID, err)
			}
			if err := appdb.UpdateMonitorStatus(s.db, m.ID, "up", 0); err != nil {
				log.Printf("checker: failed to update status for %s: %v", m.ID, err)
			}
			log.Printf("checker: %s RECOVERED", m.URL)
		} else {
			// Was already up or pending
			if err := appdb.UpdateMonitorStatus(s.db, m.ID, "up", 0); err != nil {
				log.Printf("checker: failed to update status for %s: %v", m.ID, err)
			}
		}
	} else {
		newFailures := m.ConsecutiveFailures + 1
		if newFailures >= 2 && m.Status != "down" {
			// Transition to DOWN after 2 consecutive failures
			errStr := ""
			if result.Error != nil {
				errStr = *result.Error
			}
			if _, err := appdb.CreateIncident(s.db, m.ID, errStr); err != nil {
				log.Printf("checker: failed to create incident for %s: %v", m.ID, err)
			}
			if err := appdb.UpdateMonitorStatus(s.db, m.ID, "down", newFailures); err != nil {
				log.Printf("checker: failed to update status for %s: %v", m.ID, err)
			}
			log.Printf("checker: %s is DOWN", m.URL)
		} else {
			// Still accumulating failures
			if err := appdb.UpdateMonitorStatus(s.db, m.ID, m.Status, newFailures); err != nil {
				log.Printf("checker: failed to update failure count for %s: %v", m.ID, err)
			}
		}
	}
}
