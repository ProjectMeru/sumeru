package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

var (
	templateCache   map[string]*template.Template
	templateCacheMu sync.Once
)

type ScaffoldConfig struct {
	Name        string
	HumanTitle  string
	TypeName    string
	ModelDotted string
	ModelImport string
	OutDir      string
	RepoRoot    string
	ModPath     string
}

func RunScaffold(cfg *ScaffoldConfig) {
	target := filepath.Join(cfg.RepoRoot, filepath.FromSlash(cfg.OutDir), cfg.Name)
	if fi, err := os.Stat(target); err == nil && fi.IsDir() {
		fmt.Fprintf(os.Stderr, "sumeru-bp: %q already exists\n", target)
		os.Exit(1)
	}

	dirs := []string{
		"models",
		"views",
		"security",
		"static/src/img",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(target, filepath.FromSlash(d)), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	renderTemplate(cfg, "manifest.json.tmpl", filepath.Join(target, "manifest.json"))
	renderTemplate(cfg, "init.go.tmpl", filepath.Join(target, "init.go"))
	renderTemplate(cfg, "models.go.tmpl", filepath.Join(target, "models", "models.go"))
	renderTemplate(cfg, "security.xml.tmpl", filepath.Join(target, "security", "security.xml"))
	renderTemplate(cfg, "sys.access.csv.tmpl", filepath.Join(target, "security", "sys.access.csv"))
	renderTemplate(cfg, "actions.xml.tmpl", filepath.Join(target, "views", "actions.xml"))
	renderTemplate(cfg, "form_view.xml.tmpl", filepath.Join(target, "views", "form_view.xml"))
	renderTemplate(cfg, "list_view.xml.tmpl", filepath.Join(target, "views", "list_view.xml"))
	renderTemplate(cfg, "kanban_view.xml.tmpl", filepath.Join(target, "views", "kanban_view.xml"))
	renderTemplate(cfg, "menus.xml.tmpl", filepath.Join(target, "views", "menus.xml"))

	fmt.Printf("Successfully created premium modular addon at: %s\n", target)
	fmt.Printf("Next steps:\n")
	if cfg.ModPath == "sumeru_custom_addons" {
		fmt.Printf("  1. Run 'make generate' in the workspace\n")
		fmt.Printf("  2. Run 'make install MODULES=%s'\n", cfg.Name)
	} else {
		fmt.Printf("  1. Run 'make generate'\n")
		fmt.Printf("  2. Run 'go run ./cmd/sumeru -- -i %s'\n", cfg.Name)
	}
}

func loadTemplates() {
	templateCacheMu.Do(func() {
		names, err := templateFS.ReadDir("templates")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading templates: %v\n", err)
			os.Exit(1)
		}
		templateCache = make(map[string]*template.Template, len(names))
		for _, entry := range names {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			t, err := template.ParseFS(templateFS, "templates/"+name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error parsing template %s: %v\n", name, err)
				os.Exit(1)
			}
			templateCache[name] = t
		}
	})
}

func renderTemplate(cfg *ScaffoldConfig, tmplName, outPath string) {
	loadTemplates()
	t := templateCache[tmplName]
	if t == nil {
		fmt.Fprintf(os.Stderr, "missing template %s\n", tmplName)
		os.Exit(1)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error executing template %s: %v\n", tmplName, err)
		os.Exit(1)
	}

	writeOrDie(outPath, buf.String())
}
