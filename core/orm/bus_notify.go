package orm

import (
	"fmt"
	"strconv"
)

// NotifyBusEvent sends PostgreSQL NOTIFY for multi-process fan-out.
func NotifyBusEvent(eventID int64) error {
	if DB == nil || 0 >= eventID {
		return nil
	}
	_, err := DB.Exec(`SELECT pg_notify($1, $2)`, BusNotifyChannel(), strconv.FormatInt(eventID, 10))
	if err != nil {
		return fmt.Errorf("pg_notify bus: %w", err)
	}
	return nil
}
