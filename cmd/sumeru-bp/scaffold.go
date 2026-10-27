package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// moduleScaffold is the data rendered into a new addon's files.
type moduleScaffold struct {
	ModuleName   string // technical name: folder, package, and manifest name
	DisplayName  string // human title, for example "Sale Order"
	StructName   string // Go struct name, for example "SaleOrder"
	ModelName    string // dotted model name, for example "sale.order"
	ModelsImport string // Go import path of the models package
	ParentDir    string // -out directory, relative to the repo or absolute
	RepoRoot     string
	GoModule     string // module path from go.mod
}

var scaffoldDirectories = []string{
	"models",
	"views",
	"security",
	"static/src/img",
	"i18n",
}

var scaffoldFiles = []struct {
	template string
	relPath  string
}{
	{"manifest.json.tmpl", "manifest.json"},
	{"init.go.tmpl", "init.go"},
	{"models.go.tmpl", filepath.Join("models", "models.go")},
	{"security.xml.tmpl", filepath.Join("security", "security.xml")},
	{"sys.access.csv.tmpl", filepath.Join("security", "sys.access.csv")},
	{"actions.xml.tmpl", filepath.Join("views", "actions.xml")},
	{"form_view.xml.tmpl", filepath.Join("views", "form_view.xml")},
	{"list_view.xml.tmpl", filepath.Join("views", "list_view.xml")},
	{"kanban_view.xml.tmpl", filepath.Join("views", "kanban_view.xml")},
	{"menus.xml.tmpl", filepath.Join("views", "menus.xml")},
	{"en_US.po.tmpl", filepath.Join("i18n", "en_US.po")},
}

func scaffoldModule(module moduleScaffold) error {
	directory, err := moduleDirectory(module)
	if err != nil {
		return err
	}
	for _, rel := range scaffoldDirectories {
		if err := os.MkdirAll(filepath.Join(directory, filepath.FromSlash(rel)), 0o755); err != nil {
			return fmt.Errorf("create %s: %w", rel, err)
		}
	}
	templates, err := parseTemplates()
	if err != nil {
		return err
	}
	for _, file := range scaffoldFiles {
		parsed := templates[file.template]
		if parsed == nil {
			return fmt.Errorf("template %q is not embedded", file.template)
		}
		var rendered bytes.Buffer
		if err := parsed.Execute(&rendered, module); err != nil {
			return fmt.Errorf("render %s: %w", file.template, err)
		}
		dest := filepath.Join(directory, file.relPath)
		if err := writeFile(dest, rendered.String()); err != nil {
			return err
		}
	}

	fmt.Printf("Created %s\n", directory)
	fmt.Println("Next:")
	if module.GoModule == "sumeru_custom_addons" {
		fmt.Printf("  1. Run 'make generate' in the workspace\n")
		fmt.Printf("  2. Run 'make install MODULES=%s'\n", module.ModuleName)
	} else {
		fmt.Println("  1. Run 'make generate'")
		fmt.Printf("  2. Run 'go run ./cmd/sumeru -- -i %s'\n", module.ModuleName)
	}
	return nil
}

// moduleDirectory is the new addon folder. A relative ParentDir is under RepoRoot.
// An empty ParentDir means addons/. An existing path is refused.
func moduleDirectory(module moduleScaffold) (string, error) {
	parent := module.ParentDir
	if parent == "" {
		parent = "addons"
	}
	if !filepath.IsAbs(parent) {
		parent = filepath.Join(module.RepoRoot, filepath.FromSlash(parent))
	}
	directory := filepath.Join(parent, module.ModuleName)
	info, err := os.Stat(directory)
	if err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("%q exists and is not a directory", directory)
		}
		return "", fmt.Errorf("%q already exists", directory)
	}
	if !os.IsNotExist(err) {
		return "", err
	}
	return directory, nil
}

func parseTemplates() (map[string]*template.Template, error) {
	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		return nil, fmt.Errorf("read templates: %w", err)
	}
	parsed := make(map[string]*template.Template, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		t, err := template.ParseFS(templateFS, "templates/"+name)
		if err != nil {
			return nil, fmt.Errorf("parse template %s: %w", name, err)
		}
		parsed[name] = t
	}
	return parsed, nil
}
