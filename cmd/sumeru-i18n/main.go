// Command sumeru-i18n imports or exports sys.translation CSV rows.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"sumeru/core/orm"
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

	switch args[0] {
	case "import":
		runImport(*configPath, args[1:])
	case "export":
		runExport(*configPath, args[1:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: sumeru-i18n [-c sumeru.conf] import|export [flags]

  import [-i translations.csv]
  export [-o translations.csv]
`)
}

func runImport(configPath string, subArgs []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	inPath := fs.String("i", "translations.csv", "Input CSV path")
	_ = fs.Parse(subArgs)

	f, err := os.Open(*inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	header, rows, err := orm.ParseTranslationCSV(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "csv: %v\n", err)
		os.Exit(1)
	}
	if len(rows) == 0 {
		fmt.Println("No data rows.")
		return
	}

	ctx, db, cancel, err := cliboot.OpenConfiguredDB(configPath, 2*time.Minute)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer cancel()
	defer db.Close()

	imported, err := orm.ImportTranslationsCSV(ctx, db, header, rows)
	if err != nil {
		fmt.Fprintf(os.Stderr, "import: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Imported %d translation rows from %s\n", imported, *inPath)
}

func runExport(configPath string, subArgs []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	outPath := fs.String("o", "translations.csv", "Output CSV path")
	_ = fs.Parse(subArgs)

	ctx, db, cancel, err := cliboot.OpenConfiguredDB(configPath, 30*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer cancel()
	defer db.Close()

	count, err := orm.ExportTranslationsCSV(ctx, db, *outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "export: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Exported %d translation rows to %s\n", count, *outPath)
}
