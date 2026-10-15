package applog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"

	"sumeru/core/server/config"
)

var root *slog.Logger

type handlerResult struct {
	handler    slog.Handler
	logPath    string
	maxSize    int
	maxBackups int
	maxAge     int
}

// SetupFromConfig builds the global slog logger for JSON logs to stdout and optional file.
// Call after config.LoadConfig and config.AbsPaths so log_file paths are absolute.
func SetupFromConfig(c *config.Config) error {
	root = nil
	logEnabled = c.LogEnabled
	loc, tzLabel := parseLogTimezone(c.LogTimezone)
	logLocation = loc
	logTzName = tzLabel

	if !c.LogEnabled {
		root = slog.New(slog.DiscardHandler)
		slog.SetDefault(root)
		return nil
	}

	level := slog.LevelInfo
	if c.DevMode {
		level = slog.LevelDebug
	}

	hr, err := buildHandler(c, level)
	if err != nil {
		return err
	}
	root = slog.New(hr.handler)
	slog.SetDefault(root)

	Info(context.Background(), Event{
		Message:   "Logging initialized",
		Component: "applog",
		Operation: "init",
		Status:    "success",
		Context: map[string]interface{}{
			"log_enabled":      c.LogEnabled,
			"log_timezone":     logTzName,
			"log_stdout":       c.LogStdout,
			"log_rolling":      c.LogRolling,
			"log_file":         hr.logPath,
			"log_max_size_mb":  hr.maxSize,
			"log_max_backups":  hr.maxBackups,
			"log_max_age_days": hr.maxAge,
		},
	})
	return nil
}

func buildHandler(c *config.Config, level slog.Level) (handlerResult, error) {
	logPath := strings.TrimSpace(c.LogFile)
	if c.LogRolling && logPath == "" {
		return handlerResult{}, fmt.Errorf("log_rolling=true requires log_file")
	}

	maxSize := c.LogMaxSizeMB
	if 0 >= maxSize {
		maxSize = 100
	}
	maxBackups := c.LogMaxBackups
	if maxBackups < 0 {
		maxBackups = 0
	}
	maxAge := c.LogMaxAgeDays
	if maxAge < 0 {
		maxAge = 0
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			return a
		},
	}
	writers := []io.Writer{}
	if c.LogStdout {
		writers = append(writers, os.Stdout)
	}
	if logPath != "" {
		w, err := openLogFile(logPath, c.LogRolling, maxSize, maxBackups, maxAge)
		if err != nil {
			return handlerResult{}, err
		}
		writers = append(writers, w)
	}
	if 0 == len(writers) {
		return handlerResult{}, fmt.Errorf("no log sink configured: enable log_stdout and/or set log_file")
	}
	var out io.Writer
	if len(writers) == 1 {
		out = writers[0]
	} else {
		out = io.MultiWriter(writers...)
	}
	return handlerResult{
		handler:    slog.NewJSONHandler(out, opts),
		logPath:    logPath,
		maxSize:    maxSize,
		maxBackups: maxBackups,
		maxAge:     maxAge,
	}, nil
}

func openLogFile(path string, rolling bool, maxSize, maxBackups, maxAge int) (io.Writer, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir log dir: %w", err)
	}
	if rolling {
		return &lumberjack.Logger{
			Filename:   path,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			LocalTime:  true,
		}, nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file %q: %w", path, err)
	}
	return f, nil
}

// Sync flushes buffered log entries (no-op for slog JSON handlers; kept for API compatibility).
func Sync() {}
