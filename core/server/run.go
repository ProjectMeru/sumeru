package server

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"sumeru/core/applog"
	_ "sumeru/core/ormmodels"
	"sumeru/core/orm"
	"sumeru/core/runtime"
	"sumeru/core/scheduler"
	"sumeru/core/server/config"
	"sumeru/core/server/web"
)

// Run parses flags, loads configuration, initializes persistence and modules,
// registers HTTP routes, and blocks in ListenAndServe.
//
// CLI flags (see README): -c, -d/--database, -i, -u, --http-port/-p, --stop-after-init.
func Run() {
	configPath := flag.String("c", "sumeru.conf", "Path to config file (INI)")
	installMods := flag.String("i", "", "Install modules (comma-separated). One or many: -i sales  OR  -i sales,crm")
	updateMods := flag.String("u", "", "Update modules (comma-separated). Use -u all for every installed module, or -u sales  OR  -u sales,crm")
	dbName := flag.String("d", "", "Database name; overrides db_name in config")
	dbNameLong := flag.String("database", "", "Database name (long form); same as -d if set")
	httpPort := flag.String("http-port", "", "HTTP listen port; overrides http_port in config")
	httpPortShort := flag.String("p", "", "HTTP port shorthand (same as --http-port)")
	stopAfterInit := flag.Bool("stop-after-init", false, "After -i / -u, exit without starting HTTP")
	flag.Parse()

	if err := LoadConfig(*configPath); err != nil {
		applog.BootstrapFatal("Critical Error: Failed to load configuration: %v", err)
	}
	if err := AbsPaths(); err != nil {
		applog.BootstrapFatal("Resolve paths: %v", err)
	}

	if err := applog.SetupFromConfig(&config.AppConfig); err != nil {
		applog.BootstrapFatal("Logging: %v", err)
	}
	defer applog.Sync()
	applog.RegisterUIDResolver(orm.UIDFromContext)
	applog.RegisterCompanyIDResolver(orm.CompanyIDFromContext)
	ctx := context.Background()

	if s := strings.TrimSpace(*dbNameLong); s != "" {
		config.AppConfig.DbName = s
	} else if s := strings.TrimSpace(*dbName); s != "" {
		config.AppConfig.DbName = s
	}
	if s := strings.TrimSpace(*httpPortShort); s != "" {
		config.AppConfig.HttpPort = s
	} else if s := strings.TrimSpace(*httpPort); s != "" {
		config.AppConfig.HttpPort = s
	}
	if strings.TrimSpace(*dbName) != "" || strings.TrimSpace(*dbNameLong) != "" {
		applog.InfoMsg(ctx, "server", "config", "Database name override applied",
			map[string]interface{}{"db": config.AppConfig.DbName})
	}
	if strings.TrimSpace(*httpPort) != "" || strings.TrimSpace(*httpPortShort) != "" {
		applog.InfoMsg(ctx, "server", "config", "HTTP port override applied",
			map[string]interface{}{"port": config.AppConfig.HttpPort})
	}

	applog.InfoMsg(ctx, "server", "startup", "Addon roots configured",
		map[string]interface{}{"paths": config.AppConfig.AddonPaths})

	databaseSource := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.AppConfig.DbHost, config.AppConfig.DbPort, config.AppConfig.DbUser,
		config.AppConfig.DbPass, config.AppConfig.DbName, config.AppConfig.DbSslMode)
	InitDatabase(databaseSource)

	if orm.IsInitialized() {
		if err := SyncModels(); err != nil {
			applog.Fatal(ctx, "Error syncing models", "err", err)
		}
		if err := orm.SyncRegistrySchema(); err != nil {
			applog.Fatal(ctx, "Schema sync failed", "err", err)
		}
	}

	if err := LoadAddonPaths(config.AppConfig.AddonPaths); err != nil {
		applog.Fatal(ctx, "Addon load failed", "err", err)
	}

	if !orm.IsInitialized() {
		applog.InfoMsg(ctx, "server", "startup", "Database is not initialized; starting in setup mode", nil)
		applog.InfoMsg(ctx, "server", "startup", "Visit setup URL to initialize the system",
			map[string]interface{}{"url": "http://localhost:" + config.AppConfig.HttpPort + "/setup"})

		registerBrandingAndStatic()
		registerSetupRoutes()

		listenHost := setupListenAddr(config.AppConfig)
		applog.InfoMsg(ctx, "server", "listen", "Server starting in setup mode",
			map[string]interface{}{"port": config.AppConfig.HttpPort, "bind": listenHost})
		setupHandler := web.SecurityMiddleware(nil)
		if err := http.ListenAndServe(listenHost, setupHandler); err != nil {
			applog.Fatal(ctx, "Server failed in setup mode", "err", err)
		}
		return
	}

	if err := orm.EnsureDefaultGroupsAndImplied(); err != nil {
		applog.Fatal(ctx, "Default security groups failed", "err", err)
	}

	if err := RunModuleCLI(*installMods, *updateMods); err != nil {
		applog.Fatal(ctx, "Module CLI failed", "err", err)
	}

	if err := orm.EnsureBootstrapSecurity(); err != nil {
		applog.Fatal(ctx, "Security bootstrap failed", "err", err)
	}

	hadModuleOperations := strings.TrimSpace(*installMods) != "" || strings.TrimSpace(*updateMods) != ""
	if *stopAfterInit && hadModuleOperations {
		applog.InfoMsg(ctx, "server", "shutdown", "stop-after-init: module operations finished, exiting", nil)
		os.Exit(0)
	}

	registerBrandingAndStatic()
	registerAppRoutes()
	web.InitRateLimit()
	scheduler.Start(context.Background(), time.Minute)
	orm.StartOutboxDrain(context.Background(), 5*time.Second)

	listenHost := listenAddr(config.AppConfig.HttpInterface, config.AppConfig.HttpPort)
	applog.InfoMsg(ctx, "server", "listen", "Server starting",
		map[string]interface{}{"port": config.AppConfig.HttpPort, "bind": listenHost})
	runtime.SyncFromGlobals()
	appHandler := web.SecurityMiddleware(nil)
	srv := &http.Server{
		Addr:         listenHost,
		Handler:      appHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		applog.Fatal(ctx, "Server failed", "err", err)
	}
}

func registerAppRoutes() {
	web.RegisterAppRoutes(nil)
}

func registerSetupRoutes() {
	web.RegisterSetupRoutes(nil)
}

func listenAddr(host, port string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ":" + port
	}
	return host + ":" + port
}

func setupListenAddr(cfg config.Config) string {
	if strings.TrimSpace(cfg.HttpInterface) != "" {
		return listenAddr(cfg.HttpInterface, cfg.HttpPort)
	}
	if cfg.SetupLocalhostOnly {
		return "127.0.0.1:" + cfg.HttpPort
	}
	return listenAddr("", cfg.HttpPort)
}
