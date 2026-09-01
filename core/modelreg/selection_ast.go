package modelreg

import (
	"go/ast"
	"go/build"
	"go/constant"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"

	"sumeru/core/modelmeta"
)

var (
	selectionASTMu sync.Mutex
	selectionAST   = map[string]map[string][][]string{} // pkgDir -> typeName -> options
)

func selectionOptionsFromPackage(pkgDir, typeName string) [][]string {
	if pkgDir == "" || typeName == "" {
		return nil
	}
	selectionASTMu.Lock()
	defer selectionASTMu.Unlock()
	if byType, ok := selectionAST[pkgDir]; ok {
		if opts, ok := byType[typeName]; ok {
			return opts
		}
	}
	if selectionAST[pkgDir] == nil {
		selectionAST[pkgDir] = map[string][][]string{}
	}
	opts := scanSelectionConsts(pkgDir, typeName)
	selectionAST[pkgDir][typeName] = opts
	return opts
}

func scanSelectionConsts(pkgDir, typeName string) [][]string {
	stringTypes := map[string]bool{}
	values := map[string][]string{}

	fset := token.NewFileSet()
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil
	}
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(ent.Name(), "_test.go") || ent.Name() == "zmodels.go" || ent.Name() == "zselections.go" {
			continue
		}
		path := filepath.Join(pkgDir, ent.Name())
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			continue
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			switch gd.Tok {
			case token.TYPE:
				for _, spec := range gd.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || !ts.Name.IsExported() {
						continue
					}
					id, ok := ts.Type.(*ast.Ident)
					if ok && id.Name == "string" {
						stringTypes[ts.Name.Name] = true
					}
				}
			case token.CONST:
				for _, spec := range gd.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					typedName := ""
					if vs.Type != nil {
						if id, ok := vs.Type.(*ast.Ident); ok {
							typedName = id.Name
						}
					}
					if typedName == "" || !stringTypes[typedName] {
						continue
					}
					for i, name := range vs.Names {
						if !name.IsExported() {
							continue
						}
						if i >= len(vs.Values) {
							continue
						}
						val := constStringValue(vs.Values[i])
						if val == "" {
							continue
						}
						values[typedName] = append(values[typedName], val)
					}
				}
			}
		}
	}

	raw := values[typeName]
	if len(raw) == 0 {
		return nil
	}
	sort.Strings(raw)
	out := make([][]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, key := range raw {
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, []string{key, modelmeta.LabelFromGo(key)})
	}
	return out
}

func constStringValue(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	v := constant.StringVal(constant.MakeFromLiteral(exprString(expr), token.STRING, 0))
	return strings.TrimSpace(v)
}

func exprString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Value
	default:
		return ""
	}
}

func pkgDirForStructType(st reflect.Type) string {
	if st == nil {
		return ""
	}
	pkgPath := st.PkgPath()
	if pkgPath == "" {
		return ""
	}
	pkg, err := build.Import(pkgPath, "", build.FindOnly)
	if err != nil {
		return ""
	}
	return pkg.Dir
}
