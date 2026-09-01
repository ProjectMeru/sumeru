package server

import (
	"strings"
	"time"

	"sumeru/core/mail"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

// InitDatabase opens the primary pool (and optional read replica) using INI pool settings.
func InitDatabase(primaryDSN string) {
	cfg := config.AppConfig
	orm.InitDevFeatures(cfg.DevFeatures)
	pool := orm.DBPoolSettings{
		MaxOpenConns: cfg.DbMaxOpenConns,
		MaxIdleConns: cfg.DbMaxIdleConns,
	}
	if cfg.DbConnMaxLifetimeMin > 0 {
		pool.ConnMaxLifetime = time.Duration(cfg.DbConnMaxLifetimeMin) * time.Minute
	}
	orm.InitDBWithPool(primaryDSN, pool)
	if dsn := strings.TrimSpace(cfg.DbReadReplicaDSN); dsn != "" {
		orm.InitReadReplica(dsn, pool)
	}
	mail.Configure(mail.SMTPConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		User:     cfg.SMTPUser,
		Password: cfg.SMTPPassword,
		From:     cfg.SMTPFrom,
	})
}
