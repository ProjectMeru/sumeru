// Command sumeru-module manages addons without starting the HTTP server.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"sumeru/core/module"
	"sumeru/core/server/cliboot"
)

func main() {
	cliboot.StripLeadingArgsSeparator()
	configPath := flag.String("c", "sumeru.conf", "Path to config file (INI)")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}
	cmd := strings.ToLower(strings.TrimSpace(args[0]))
	subArgs := args[1:]

	switch cmd {
	case "list":
		runList(*configPath)
	case "depends-tree", "depends":
		if len(subArgs) == 0 {
			fmt.Fprintln(os.Stderr, "usage: sumeru-module depends-tree MODULE")
			os.Exit(1)
		}
		runDepends(*configPath, subArgs[0])
	case "install":
		if len(subArgs) == 0 {
			fmt.Fprintln(os.Stderr, "usage: sumeru-module install mod1,mod2")
			os.Exit(1)
		}
		runInstall(*configPath, strings.Join(subArgs, " "))
	case "update":
		if len(subArgs) == 0 {
			fmt.Fprintln(os.Stderr, "usage: sumeru-module update mod|all")
			os.Exit(1)
		}
		runUpdate(*configPath, strings.Join(subArgs, " "))
	case "uninstall":
		if len(subArgs) == 0 {
			fmt.Fprintln(os.Stderr, "usage: sumeru-module uninstall MODULE")
			os.Exit(1)
		}
		runUninstall(*configPath, subArgs[0])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: sumeru-module [-c sumeru.conf] COMMAND

Commands:
  list                         List discovered modules and DB state
  depends-tree MODULE          Print dependency tree
  install mod1,mod2            Install modules
  update mod|all               Update module metadata from disk
  uninstall MODULE             Uninstall module
`)
}

func runList(configPath string) {
	ctx, err := cliboot.InitOptionalDB(configPath, true)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	infos, err := module.ListModuleInfo(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, m := range infos {
		app := ""
		if m.Application {
			app = " app"
		}
		fmt.Printf("%-20s %-12s v%s%s\n", m.Name, m.State, m.Version, app)
	}
}

func runDepends(configPath, mod string) {
	_, err := cliboot.InitOptionalDB(configPath, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	out, err := module.DependsTree(mod)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(out)
}

func runInstall(configPath, csv string) {
	if _, err := cliboot.Init(configPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := module.RunModuleCLI(csv, ""); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Install complete.")
}

func runUpdate(configPath, csv string) {
	if _, err := cliboot.Init(configPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := module.RunModuleCLI("", csv); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Update complete.")
}

func runUninstall(configPath, mod string) {
	ctx, err := cliboot.Init(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := module.UninstallModuleByName(ctx, strings.TrimSpace(mod)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Uninstalled %s\n", mod)
}
