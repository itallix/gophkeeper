package service

import "time"

type AuditLog struct {
	// user string
	// operation string
	// path string
	// access bool
	// created_at time.Time
}

// AuditLogService logs all operations for auditing purposes.
type AuditLogService interface {
	Log(user string, operation string, path string, success bool) error
	GetLogs(startTime, endTime time.Time) ([]AuditLog, error)
}
