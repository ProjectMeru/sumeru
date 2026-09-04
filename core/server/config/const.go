package config

const (
	defaultDbSSLMode = "disable"

	iniSectionPrefix = "["
	iniCommentSemi   = ";"
	iniCommentHash   = "#"
	iniSeparator     = "="
	iniKeyValueParts = 2

	keyDbHost             = "db_host"
	keyDbPort             = "db_port"
	keyDbUser             = "db_user"
	keyDbPassword         = "db_password"
	keyDbName             = "db_name"
	keyDbSSLMode          = "db_sslmode"
	keyHTTPPort           = "http_port"
	keyHTTPInterface      = "http_interface"
	keyAddonsPath         = "addons_path"
	keySumeruHome         = "sumeru_home"
	keyAssetsPath         = "assets_path"
	keyTemplatesPath      = "templates_path"
	keyBrandCSS           = "brand_css"
	keyLogoPath           = "logo_path"
	keyCompanyDisplayName = "company_display_name"
	keyUserDisplayName    = "user_display_name"
	keyLogFile            = "log_file"
	keyLogStdout          = "log_stdout"
	keyLogRolling         = "log_rolling"
	keyLogMaxSizeMB       = "log_max_size_mb"
	keyLogMaxBackups      = "log_max_backups"
	keyLogMaxAgeDays      = "log_max_age_days"
	keyLogEnabled         = "log_enabled"
	keyLogTimezone        = "log_timezone"
	keyDevMode            = "dev_mode"
	keyDevFeatures        = "dev_features"
	keySetupToken         = "setup_token"
	keySetupLocalhostOnly = "setup_localhost_only"
	keyDbMaxOpenConns     = "db_max_open_conns"
	keyDbMaxIdleConns     = "db_max_idle_conns"
	keyDbConnMaxLifetimeMin = "db_conn_max_lifetime_minutes"
	keyDbReadReplicaDSN   = "db_read_replica_dsn"
	keyRateLimitRPM       = "rate_limit_rpm"
	keyTrustedProxies     = "trusted_proxies"
	keyCSRFSecret         = "csrf_secret"
	keyMetricsScrapeToken = "metrics_scrape_token"
	keySMTPHost           = "smtp_host"
	keySMTPPort           = "smtp_port"
	keySMTPUser           = "smtp_user"
	keySMTPPassword       = "smtp_password"
	keySMTPFrom           = "smtp_from"
)

const (
	relPathDefaultAssets    = "core/engine/assets"
	relPathDefaultTemplates = "core/engine/templates"
)

const (
	fileGoMod = "go.mod"

	segCore      = "core"
	segEngine    = "engine"
	segAssets    = "assets"
	segTemplates = "templates"
)

const addonsPathDelimiter = ","

const (
	errFmtDbHostRequired     = "db_host is required in %s"
	errFmtDbPortRequired     = "db_port is required in %s"
	errFmtDbUserRequired     = "db_user is required in %s"
	errFmtDbPasswordRequired = "db_password is required in %s"
	errFmtDbNameRequired     = "db_name is required in %s"
	errFmtHTTPPortRequired   = "http_port is required in %s"
	errFmtAddonsPathRequired = "addons_path is required in %s"
)
