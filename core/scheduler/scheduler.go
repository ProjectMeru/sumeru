package scheduler

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"sumeru/core/applog"
	"sumeru/core/errcode"
	"sumeru/core/event"
	"sumeru/core/orm"
)

var (
	mu     sync.Mutex
	cancel context.CancelFunc
)

// Start begins a background ticker that evaluates due cron rows.
func Start(parent context.Context, every time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	if cancel != nil {
		return
	}
	if every <= 0 {
		every = time.Minute
	}
	ctx, c := context.WithCancel(parent)
	cancel = c
	go loop(ctx, every)
}

// Stop halts the background ticker.
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if cancel != nil {
		cancel()
		cancel = nil
	}
}

func loop(ctx context.Context, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			runDue(ctx)
		}
	}
}

func runDue(ctx context.Context) {
	if orm.DB == nil {
		return
	}
	if _, ok := orm.Registry["sys.cron"]; !ok {
		return
	}
	bypass := orm.ContextWithBypass(ctx, true)
	tbl := orm.MustQuotedTableName("sys.cron")
	now := time.Now().UTC()

	tx, err := orm.DB.BeginTx(bypass, nil)
	if err != nil {
		return
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.QueryContext(bypass,
		`SELECT id, name, COALESCE(event_name,''), COALESCE(code,'') FROM `+tbl+
			` WHERE active = true AND (next_call IS NULL OR next_call <= $1)
			  ORDER BY id FOR UPDATE SKIP LOCKED LIMIT 20`, now,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	type cronRow struct {
		id        int64
		name      string
		eventName string
		code      string
	}
	var due []cronRow
	for rows.Next() {
		var row cronRow
		if err := rows.Scan(&row.id, &row.name, &row.eventName, &row.code); err != nil {
			continue
		}
		due = append(due, row)
	}
	if err := rows.Err(); err != nil {
		return
	}

	for _, row := range due {
		executeCron(bypass, CronRunInput{ID: row.id, Name: row.name, EventName: row.eventName, Code: row.code})
		interval := cronIntervalTx(bypass, tx, row.id)
		next := now.Add(interval)
		if _, err := tx.ExecContext(bypass,
			`UPDATE `+tbl+` SET next_call = $1, last_call = $2 WHERE id = $3`,
			next, now, row.id,
		); err != nil {
			applog.WarnCode(bypass, errcode.CronUpdateFailed, "cron next_call update failed", applog.Event{
				Component: "scheduler",
				Operation: "run_due",
				Status:    "failure",
				Context:   map[string]interface{}{"cron_id": row.id},
				Err:       err,
			})
			return
		}
	}
	if err := tx.Commit(); err != nil {
		applog.WarnCode(bypass, errcode.CronCommitFailed, "cron transaction commit failed", applog.Event{
			Component: "scheduler",
			Operation: "run_due",
			Status:    "failure",
			Err:       err,
		})
	}
}

func cronIntervalTx(ctx context.Context, tx orm.TxWrapper, id int64) time.Duration {
	var mins sql.NullInt64
	if err := tx.QueryRowContext(ctx,
		`SELECT interval_number FROM `+orm.MustQuotedTableName("sys.cron")+` WHERE id = $1`, id,
	).Scan(&mins); err != nil {
		applog.WarnCode(ctx, errcode.InternalError, "cron interval read failed; using default 60m", applog.Event{
			Component: "scheduler",
			Operation: "cron_interval",
			Status:    "partial",
			Context:   map[string]interface{}{"cron_id": id},
			Err:       err,
		})
		return time.Hour
	}
	n := int(mins.Int64)
	if n <= 0 {
		n = 60
	}
	return time.Duration(n) * time.Minute
}

type CronRunInput struct {
	ID        int64
	Name      string
	EventName string
	Code      string
}

func executeCron(ctx context.Context, in CronRunInput) {
	applog.Info(ctx, applog.Event{
		Message:   "cron job started",
		Component: "scheduler",
		Operation: "cron_run",
		Status:    "success",
		Context: map[string]interface{}{
			"cron_id":   in.ID,
			"cron_name": in.Name,
			"cron_code": in.Code,
		},
	})
	payload := map[string]interface{}{"cron_id": in.ID, "cron_name": in.Name, "code": in.Code}
	_ = event.Publish(ctx, event.Event{Name: "cron.tick", Payload: payload})
	if eventName := strings.TrimSpace(in.EventName); eventName != "" {
		_ = event.Publish(ctx, event.Event{Name: eventName, Payload: payload})
	}
	if fn := lookupCronHandler(in.Code); fn != nil {
		if err := fn(ctx, payload); err != nil {
			applog.WarnCode(ctx, errcode.CronHandlerFailed, "cron handler failed", applog.Event{
				Component: "scheduler",
				Operation: "cron_handler",
				Status:    "failure",
				Context:   map[string]interface{}{"cron_id": in.ID, "cron_code": in.Code},
				Err:       err,
			})
		}
	}
}
