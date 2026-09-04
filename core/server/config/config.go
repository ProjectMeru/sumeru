package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	DbHost             string
	DbPort             string
	DbUser             string
	DbPass             string
	DbName             string
	DbSslMode          string
	HttpPort           string
	HttpInterface      string   // optional bind address; empty = all interfaces (:port)
	AddonsPath         string   // raw from file: comma-separated addon directory roots (see AddonPaths after AbsPaths)
	AddonPaths         []string // absolute addon roots from addons_path; filled by AbsPaths()
	SumeruHome         string   // optional: directory of standard sumeru repo (go.mod); used for default assets/templates if set
	AssetsPath         string   // static files (CSS/JS); default core/engine/assets
	TemplatesPath      string   // HTML templates; default core/engine/templates
	BrandCSS           string   // optional path to extra CSS (served as /static/brand.css)
	LogoPath           string   // optional image path (served as /static/app-logo)
	CompanyDisplayName string   // optional header label; else first core.company when module installed
	UserDisplayName    string   // optional header label; else first core.user when module installed
	LogFile            string   // optional log file path (absolutized in AbsPaths); see log_stdout / log_rolling
	LogStdout          bool     // when true, emit JSON logs to stdout (typical for Kubernetes)
	LogRolling         bool     // when true and log_file set, use size-based rotation (lumberjack); false for append-only or external rotation
	LogMaxSizeMB       int      // max megabytes per log file before rotation (default 100 when log_rolling)
	LogMaxBackups      int      // retained rotated files (0 = lumberjack default)
	LogMaxAgeDays      int      // delete rolled files older than N days (0 = no age limit)
	LogEnabled         bool     // log_enabled: when false, no slog sinks and L(ctx) is no-op; stdlib log discarded
	LogTimezone        string   // log_timezone: UTC, Local (default), or IANA (e.g. Asia/Kolkata) for timestamps
	DevMode            bool     // dev_mode INI key; parseBoolKey(..., false) — debug slog level and dev-only server paths
	DevFeatures        string   // dev_features INI: comma-separated sql, access, xml
	SetupToken         string   // secret for POST /setup/init; required when setup_localhost_only is false
	SetupLocalhostOnly bool     // when true (default), setup mode listens on 127.0.0.1 only; false requires setup_token
	DbMaxOpenConns     int      // db_max_open_conns; 0 = Go default
	DbMaxIdleConns     int      // db_max_idle_conns; 0 = Go default
	DbConnMaxLifetimeMin int    // db_conn_max_lifetime_minutes; 0 = no limit
	DbReadReplicaDSN   string   // optional libpq DSN for read replica (search/read_group RPC)
	RateLimitRPM       int      // rate_limit_rpm per client IP on /api/rpc and login; 0 = disabled
	TrustedProxies     string   // trusted_proxies: comma-separated CIDRs/IPs allowed to set X-Forwarded-For; empty = never trust XFF
	CSRFSecret         string   // csrf_secret: shared HMAC key for multi-instance; empty = ephemeral per process
	MetricsScrapeToken string   // metrics_scrape_token: Bearer token for unauthenticated /metrics scrape; empty = admin session only
	SMTPHost           string
	SMTPPort           int
	SMTPUser           string
	SMTPPassword       string
	SMTPFrom           string
}

var AppConfig Config

// ConfigFileDir is the absolute directory of the last successfully loaded INI (set by LoadConfig).
// Relative addons_path segments resolve against this directory.
var ConfigFileDir string

func LoadConfig(path string) error {
	AppConfig = Config{
		DbSslMode:          defaultDbSSLMode,
		LogStdout:          true,
		LogEnabled:         true,
		LogMaxSizeMB:       100,
		SetupLocalhostOnly: true,
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	ConfigFileDir = filepath.Dir(absPath)

	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, iniSectionPrefix) || strings.HasPrefix(line, iniCommentSemi) || strings.HasPrefix(line, iniCommentHash) {
			continue
		}

		parts := strings.SplitN(line, iniSeparator, iniKeyValueParts)
		if len(parts) != iniKeyValueParts {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case keyDbHost:
			AppConfig.DbHost = val
		case keyDbPort:
			AppConfig.DbPort = val
		case keyDbUser:
			AppConfig.DbUser = val
		case keyDbPassword:
			AppConfig.DbPass = val
		case keyDbName:
			AppConfig.DbName = val
		case keyDbSSLMode:
			AppConfig.DbSslMode = val
		case keyHTTPPort:
			AppConfig.HttpPort = val
		case keyHTTPInterface:
			AppConfig.HttpInterface = val
		case keyAddonsPath:
			AppConfig.AddonsPath = val
		case keySumeruHome:
			AppConfig.SumeruHome = val
		case keyAssetsPath:
			AppConfig.AssetsPath = val
		case keyTemplatesPath:
			AppConfig.TemplatesPath = val
		case keyBrandCSS:
			AppConfig.BrandCSS = val
		case keyLogoPath:
			AppConfig.LogoPath = val
		case keyCompanyDisplayName:
			AppConfig.CompanyDisplayName = val
		case keyUserDisplayName:
			AppConfig.UserDisplayName = val
		case keyLogFile:
			AppConfig.LogFile = val
		case keyLogStdout:
			AppConfig.LogStdout = parseBoolKey(val, true)
		case keyLogRolling:
			AppConfig.LogRolling = parseBoolKey(val, false)
		case keyLogMaxSizeMB:
			if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				AppConfig.LogMaxSizeMB = n
			}
		case keyLogMaxBackups:
			if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				AppConfig.LogMaxBackups = n
			}
		case keyLogMaxAgeDays:
			if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				AppConfig.LogMaxAgeDays = n
			}
		case keyLogEnabled:
			AppConfig.LogEnabled = parseBoolKey(val, true)
		case keyLogTimezone:
			AppConfig.LogTimezone = val
		case keyDevMode:
			AppConfig.DevMode = parseBoolKey(val, false)
		case keyDevFeatures:
			AppConfig.DevFeatures = val
		case keySetupToken:
			AppConfig.SetupToken = val
		case keySetupLocalhostOnly:
			AppConfig.SetupLocalhostOnly = parseBoolKey(val, true)
		case keyDbMaxOpenConns:
			if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				AppConfig.DbMaxOpenConns = n
			}
		case keyDbMaxIdleConns:
			if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				AppConfig.DbMaxIdleConns = n
			}
		case keyDbConnMaxLifetimeMin:
			if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				AppConfig.DbConnMaxLifetimeMin = n
			}
		case keyDbReadReplicaDSN:
			AppConfig.DbReadReplicaDSN = val
		case keyRateLimitRPM:
			if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				AppConfig.RateLimitRPM = n
			}
		case keyTrustedProxies:
			AppConfig.TrustedProxies = val
		case keyCSRFSecret:
			AppConfig.CSRFSecret = val
		case keyMetricsScrapeToken:
			AppConfig.MetricsScrapeToken = val
		case keySMTPHost:
			AppConfig.SMTPHost = val
		case keySMTPPort:
			if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				AppConfig.SMTPPort = n
			}
		case keySMTPUser:
			AppConfig.SMTPUser = val
		case keySMTPPassword:
			AppConfig.SMTPPassword = val
		case keySMTPFrom:
			AppConfig.SMTPFrom = val
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	ApplyEnvOverrides(&AppConfig)

	if err := validateRequired(&AppConfig, absPath); err != nil {
		return err
	}

	ApplyProductionSecurityDefaults(&AppConfig)

	// Default assets/templates paths are applied in AbsPaths() so sumeru_home can anchor
	// them under the standard tree; do not set repo-relative defaults here (would resolve
	// from the INI directory and break workspace configs next to ../sumeru).

	return nil
}

// ApplyEnvOverrides applies SUMERU_* environment variables over INI values (container secrets).
func ApplyEnvOverrides(c *Config) {
	if c == nil {
		return
	}
	if v := strings.TrimSpace(os.Getenv("SUMERU_DB_PASSWORD")); v != "" {
		c.DbPass = v
	}
	if v := strings.TrimSpace(os.Getenv("SUMERU_DB_USER")); v != "" {
		c.DbUser = v
	}
	if v := strings.TrimSpace(os.Getenv("SUMERU_DB_HOST")); v != "" {
		c.DbHost = v
	}
	if v := strings.TrimSpace(os.Getenv("SUMERU_DB_NAME")); v != "" {
		c.DbName = v
	}
	if v := strings.TrimSpace(os.Getenv("SUMERU_SETUP_TOKEN")); v != "" {
		c.SetupToken = v
	}
	if v := strings.TrimSpace(os.Getenv("SUMERU_CSRF_SECRET")); v != "" {
		c.CSRFSecret = v
	}
	if v := strings.TrimSpace(os.Getenv("SUMERU_METRICS_SCRAPE_TOKEN")); v != "" {
		c.MetricsScrapeToken = v
	}
	if v := strings.TrimSpace(os.Getenv("SUMERU_SMTP_PASSWORD")); v != "" {
		c.SMTPPassword = v
	}
	if v := strings.TrimSpace(os.Getenv("SUMERU_SMTP_HOST")); v != "" {
		c.SMTPHost = v
	}
	if v := strings.TrimSpace(os.Getenv("SUMERU_SMTP_FROM")); v != "" {
		c.SMTPFrom = v
	}
}

const defaultProdRateLimitRPM = 120

// ApplyProductionSecurityDefaults sets safe defaults when not in dev_mode and returns operator warnings.
func ApplyProductionSecurityDefaults(c *Config) []string {
	if c == nil {
		return nil
	}
	var warns []string
	if !c.DevMode && c.RateLimitRPM == 0 {
		c.RateLimitRPM = defaultProdRateLimitRPM
		warns = append(warns, "rate_limit_rpm defaulted to 120 because dev_mode is false")
	}
	if c.DevMode {
		warns = append(warns, "dev_mode=true: session cookies are not Secure; do not use in production")
		if c.RateLimitRPM == 0 {
			warns = append(warns, "rate_limit_rpm=0: login/RPC rate limiting disabled")
		}
	}
	return warns
}

// parseBoolKey parses INI booleans; empty string returns defaultVal.
func parseBoolKey(val string, defaultVal bool) bool {
	s := strings.TrimSpace(strings.ToLower(val))
	if s == "" {
		return defaultVal
	}
	if s == "0" || s == "false" || s == "no" || s == "off" {
		return false
	}
	return s == "1" || s == "true" || s == "yes" || s == "on"
}

func validateRequired(c *Config, path string) error {
	if c.DbHost == "" {
		return fmt.Errorf(errFmtDbHostRequired, path)
	}
	if c.DbPort == "" {
		return fmt.Errorf(errFmtDbPortRequired, path)
	}
	if c.DbUser == "" {
		return fmt.Errorf(errFmtDbUserRequired, path)
	}
	if c.DbPass == "" {
		return fmt.Errorf(errFmtDbPasswordRequired, path)
	}
	if c.DbName == "" {
		return fmt.Errorf(errFmtDbNameRequired, path)
	}
	if c.HttpPort == "" {
		return fmt.Errorf(errFmtHTTPPortRequired, path)
	}
	if c.AddonsPath == "" {
		return fmt.Errorf(errFmtAddonsPathRequired, path)
	}
	return nil
}
