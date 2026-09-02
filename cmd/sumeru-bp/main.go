// Command sumeru-bp scaffolds a new addon under addons/ following strict Sumeru conventions.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/module"
)

func main() {
	namePtr := flag.String("name", "", "technical module name (folder + manifest name)")
	bpPtr := flag.String("bp", "", "alias for -name")

	outDir := flag.String("out", "", "parent directory for the new module (default: detected workspace addons/)")
	flag.Parse()

	name := *namePtr
	if name == "" {
		name = *bpPtr
	}

	if strings.TrimSpace(name) == "" {
		fmt.Fprintln(os.Stderr, "usage: sumeru-bp -name <my_module> [-out <path>]")
		fmt.Fprintln(os.Stderr, "alias: -bp")
		os.Exit(2)
	}
	if !module.ModuleNamePattern.MatchString(name) {
		fmt.Fprintf(os.Stderr, "invalid module name %q (must match %s)\n", name, module.ModuleNamePattern.String())
		os.Exit(2)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Try to find the repo root
	repoRoot, err := module.FindRepoRoot(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sumeru-bp: %v\n", err)
		os.Exit(1)
	}

	modPath, err := module.ReadModulePath(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sumeru-bp: %v\n", err)
		os.Exit(1)
	}

	// Workspace detection and outDir defaulting
	resolvedOut := *outDir
	if resolvedOut == "" {
		resolvedOut = "addons"
	}

	cfg := &ScaffoldConfig{
		Name:        name,
		HumanTitle:  humanTitle(name),
		TypeName:    typeName(name),
		ModelDotted: strings.ReplaceAll(name, "_", "."),
		ModelImport: modPath + "/" + filepath.ToSlash(filepath.Join(resolvedOut, name, "models")),
		OutDir:      resolvedOut,
		RepoRoot:    repoRoot,
		ModPath:     modPath,
	}

	RunScaffold(cfg)
}
