package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sumeru/core/modelmeta"
)

type SysBusEvent struct {
	modelmeta.ModelMeta `sumeru:"model=sys.bus.event"`

	Channel     modelmeta.String   `sumeru:"required,index"`
	PayloadJson modelmeta.Text     `sumeru:"column=payload_json"`
	CreatedAt   modelmeta.DateTime `sumeru:"required,index,column=created_at"`
}

const busNotifyChannel = "sumeru_bus"

// BusNotifyChannel returns the PostgreSQL NOTIFY channel name for bus fan-out.
func BusNotifyChannel() string {
	return busNotifyChannel
}

// InsertBusEvent persists an event and returns its id (0 if DB unavailable).
func InsertBusEvent(ctx context.Context, channel string, payload map[string]interface{}) (int64, error) {
	if DB == nil || channel == "" {
		return 0, fmt.Errorf("bus event: database or channel missing")
	}
	if _, ok := Registry["sys.bus.event"]; !ok {
		return 0, fmt.Errorf("sys.bus.event model not registered")
	}
	pj := ""
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return 0, err
		}
		pj = string(b)
	}
	inst := Registry["sys.bus.event"]
	id, err := Create(ctx, inst, map[string]interface{}{
		"channel":      channel,
		"payload_json": pj,
		"created_at":   time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return 0, err
	}
	return int64(id), nil
}

// LoadBusEventByID loads channel and payload for a persisted event id.
func LoadBusEventByID(ctx context.Context, eventID int64) (channel string, payload map[string]interface{}, err error) {
	if DB == nil || 0 >= eventID {
		return "", nil, fmt.Errorf("invalid bus event load")
	}
	tbl, ok := busEventTable()
	if !ok {
		return "", nil, fmt.Errorf("sys.bus.event table missing")
	}
	var ch string
	var pj string
	err = DB.QueryRowContext(ctx, `SELECT channel, COALESCE(payload_json,'') FROM `+tbl+` WHERE id = $1`, eventID).Scan(&ch, &pj)
	if err != nil {
		return "", nil, err
	}
	payload = map[string]interface{}{}
	if pj != "" {
		_ = json.Unmarshal([]byte(pj), &payload)
	}
	return ch, payload, nil
}

// ListBusEventsAfter returns events for channels with id > afterID (cap limit).
func ListBusEventsAfter(ctx context.Context, channels []string, afterID int64, limit int) ([]BusEventRow, error) {
	if DB == nil || 0 == len(channels) {
		return nil, nil
	}
	tbl, ok := busEventTable()
	if !ok {
		return nil, nil
	}
	if 0 >= limit || limit > 500 {
		limit = 500
	}
	// Build IN clause for channels — ponytail: small channel list per subscribe frame.
	placeholders := make([]string, len(channels))
	args := make([]interface{}, 0, len(channels)+1)
	args = append(args, afterID)
	for i, ch := range channels {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, ch)
	}
	inClause := strings.Join(placeholders, ",")
	query := fmt.Sprintf(
		`SELECT id, channel, COALESCE(payload_json,'') FROM `+tbl+` WHERE id > $1 AND channel IN (%s) ORDER BY id ASC LIMIT %d`,
		inClause,
		limit,
	)
	rows, err := DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BusEventRow
	for rows.Next() {
		var row BusEventRow
		var pj string
		if err := rows.Scan(&row.ID, &row.Channel, &pj); err != nil {
			return out, err
		}
		row.Payload = map[string]interface{}{}
		if pj != "" {
			_ = json.Unmarshal([]byte(pj), &row.Payload)
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

type BusEventRow struct {
	ID      int64
	Channel string
	Payload map[string]interface{}
}

func busEventTable() (string, bool) {
	if _, ok := Registry["sys.bus.event"]; !ok {
		return "", false
	}
	return MustQuotedTableName("sys.bus.event"), true
}
