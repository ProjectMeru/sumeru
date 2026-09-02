// Command sumeru-db-check validates sumeru.conf and PostgreSQL connectivity.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"sumeru/core/server/cliboot"
	"sumeru/core/server/config"
)

func main() {
	cliboot.StripLeadingArgsSeparator()
	configPath := flag.String("c", "sumeru.conf", "Path to config file (INI)")
	flag.Parse()

	_, db, cancel, err := cliboot.OpenConfiguredDB(*configPath, 10*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer cancel()
	defer db.Close()

	c := config.AppConfig
	fmt.Printf("OK: database %s@%s:%s/%s reachable\n", c.DbUser, c.DbHost, c.DbPort, c.DbName)
}
