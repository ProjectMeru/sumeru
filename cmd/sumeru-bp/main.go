// Command sumeru-bp scaffolds a new addon under addons/ following strict Sumeru conventions.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"sumeru/core/module"
)

func main() {
	nameFlag := flag.String("name", "", "technical module name (folder and manifest name)")
	bpFlag := flag.String("bp", "", "alias for -name")
	parentFlag := flag.String("out", "", "parent directory for the new module (default: addons/)")
	flag.Parse()

	moduleName := strings.TrimSpace(*nameFlag)
	if moduleName == "" {
		moduleName = strings.TrimSpace(*bpFlag)
	}
	if moduleName == "" {
		fmt.Fprintln(os.Stderr, "usage: sumeru-bp -name <my_module> [-out <path>]")
		fmt.Fprintln(os.Stderr, "alias: -bp")
		os.Exit(2)
	}
	if !module.ModuleNamePattern.MatchString(moduleName) {
		fmt.Fprintf(os.Stderr, "sumeru-bp: invalid module name %q (must match %s)\n", moduleName, module.ModuleNamePattern.String())
		os.Exit(2)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "sumeru-bp: %v\n", err)
		os.Exit(1)
	}
	repoRoot, err := module.FindRepoRoot(workingDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sumeru-bp: %v\n", err)
		os.Exit(1)
	}
	goModule, err := module.ReadModulePath(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sumeru-bp: %v\n", err)
		os.Exit(1)
	}

	parentDir := strings.TrimSpace(*parentFlag)
	if parentDir == "" {
		parentDir = "addons"
	}
	err = scaffoldModule(moduleScaffold{
		ModuleName:   moduleName,
		DisplayName:  displayName(moduleName),
		StructName:   goStructName(moduleName),
		ModelName:    strings.ReplaceAll(moduleName, "_", "."),
		ModelsImport: modelsImportPath(goModule, repoRoot, parentDir, moduleName),
		ParentDir:    parentDir,
		RepoRoot:     repoRoot,
		GoModule:     goModule,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "sumeru-bp: %v\n", err)
		os.Exit(1)
	}
}
