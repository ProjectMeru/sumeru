package importgen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"sumeru/core/modelmeta"
	"sumeru/core/module"
)

type modelRef struct {
	GoName         string
	TechnicalModel string
	ImportPath     string
	ImportAlias    string
	HeuristicName  string
	UseAlias       bool
	PhantomName    string
	IsExtend       bool
}

type exportedRef struct {
	Name           string
	Kind           string
	TechnicalModel string
	ImportPath     string
	ImportAlias    string
	SourceGoName   string
}

func scanModelsInDir(dir string) ([]modelRef, error) {
	st, err := os.Stat(dir)
	if err != nil || !st.IsDir() {
		return nil, nil
	}

	importPath, err := module.PackageImportPath(dir)
	if err != nil {
		return nil, err
	}

	models, err := parseDirModels(dir)
	if err != nil {
		return nil, err
	}

	var refs []modelRef
	for _, m := range models {
		if m.ModelName == "" || m.ModelName == "-" {
			continue
		}
		heuristic := modelmeta.HeuristicGoName(m.ModelName)
		ref := modelRef{
			GoName:         m.GoName,
			TechnicalModel: m.ModelName,
			ImportPath:     importPath,
			HeuristicName:  heuristic,
			IsExtend:       m.Extend,
		}
		if ref.GoName == heuristic ||
			strings.EqualFold(ref.GoName, heuristic) ||
			modelmeta.ModelNameFromGo(ref.GoName) == m.ModelName {
			ref.UseAlias = true
		} else if heuristic != "" {
			ref.PhantomName = heuristic
		} else {
			ref.UseAlias = true
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

func scanUsedRelationTypes(modelsDir string) (map[string]struct{}, error) {
	used := map[string]struct{}{}
	local := map[string]struct{}{}

	st, err := os.Stat(modelsDir)
	if err != nil || !st.IsDir() {
		return used, nil
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, modelsDir, isSourceGo, 0) //nolint:staticcheck // SA1019: ParseDir adequate for model tag scan; go/packages migration is separate
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			for _, decl := range f.Decls {
				gd, ok := decl.(*ast.GenDecl)
				if !ok || gd.Tok != token.TYPE {
					continue
				}
				for _, spec := range gd.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || !ts.Name.IsExported() {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok || !structEmbedsModel(st) {
						continue
					}
					local[ts.Name.Name] = struct{}{}
				}
			}
			ast.Inspect(f, func(n ast.Node) bool {
				idx, ok := n.(*ast.IndexExpr)
				if !ok {
					return true
				}
				name, ok := relationTypeArg(idx)
				if !ok || name == "" || name == "Any" {
					return true
				}
				used[name] = struct{}{}
				return true
			})
		}
	}
	for name := range local {
		delete(used, name)
	}
	return used, nil
}

func relationTypeArg(idx *ast.IndexExpr) (string, bool) {
	sel, ok := idx.X.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	x, ok := sel.X.(*ast.Ident)
	if !ok || (x.Name != "sdk" && x.Name != "modelmeta") {
		return "", false
	}
	switch sel.Sel.Name {
	case "Many2One", "One2Many", "Many2Many":
	default:
		return "", false
	}
	switch t := idx.Index.(type) {
	case *ast.Ident:
		return t.Name, true
	case *ast.SelectorExpr:
		return t.Sel.Name, true
	default:
		return "", false
	}
}

func collectModelRefs(workspace, sumeruRoot, addonsRoot string, modules []string) ([]modelRef, error) {
	importAliases := map[string]string{}
	var all []modelRef

	for _, mod := range modules {
		mod = strings.TrimSpace(mod)
		if mod == "" {
			continue
		}
		for _, dir := range dirsForModule(workspace, sumeruRoot, addonsRoot, mod) {
			refs, err := scanModelsInDir(dir)
			if err != nil {
				return nil, fmt.Errorf("scan %s: %w", dir, err)
			}
			if len(refs) == 0 {
				continue
			}
			for i := range refs {
				alias := importAliases[refs[i].ImportPath]
				if alias == "" {
					alias = importAliasForPath(refs[i].ImportPath)
					importAliases[refs[i].ImportPath] = alias
				}
				refs[i].ImportAlias = alias
				all = append(all, refs[i])
			}
		}
	}
	return all, nil
}

func filterRefsForUsage(refs []modelRef, used map[string]struct{}) []modelRef {
	if len(used) == 0 {
		return nil
	}
	byTech := map[string]modelRef{}
	for _, ref := range refs {
		if ref.TechnicalModel == "" {
			continue
		}
		existing, ok := byTech[ref.TechnicalModel]
		if !ok {
			byTech[ref.TechnicalModel] = ref
			continue
		}
		// Prefer the defining model= struct over inherit= extensions.
		if existing.IsExtend && !ref.IsExtend {
			byTech[ref.TechnicalModel] = ref
		}
	}

	var out []modelRef
	seen := map[string]struct{}{}
	for name := range used {
		for _, ref := range byTech {
			if ref.GoName == name || ref.PhantomName == name || ref.HeuristicName == name {
				if _, ok := seen[ref.TechnicalModel]; ok {
					break
				}
				seen[ref.TechnicalModel] = struct{}{}
				out = append(out, ref)
				break
			}
		}
	}
	return out
}

func buildExportedRefs(refs []modelRef, used map[string]struct{}) []exportedRef {
	seen := map[string]struct{}{}
	var out []exportedRef

	add := func(r exportedRef) {
		if r.Name == "" {
			return
		}
		if _, ok := seen[r.Name]; ok {
			return
		}
		seen[r.Name] = struct{}{}
		out = append(out, r)
	}

	for _, ref := range refs {
		needed := func(name string) bool {
			_, ok := used[name]
			return ok
		}
		if ref.UseAlias && needed(ref.GoName) {
			add(exportedRef{
				Name: ref.GoName, Kind: "alias",
				TechnicalModel: ref.TechnicalModel, ImportPath: ref.ImportPath,
				ImportAlias: ref.ImportAlias, SourceGoName: ref.GoName,
			})
			continue
		}
		if ref.PhantomName != "" && needed(ref.PhantomName) {
			add(exportedRef{
				Name: ref.PhantomName, Kind: "phantom",
				TechnicalModel: ref.TechnicalModel, ImportPath: ref.ImportPath,
				ImportAlias: ref.ImportAlias, SourceGoName: ref.GoName,
			})
		}
	}

	return out
}
