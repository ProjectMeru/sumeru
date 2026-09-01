package importgen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"sort"
	"strings"

	"sumeru/core/modelmeta"
)

func isSourceGo(fi os.FileInfo) bool {
	name := fi.Name()
	return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") &&
		name != "zmodels.go" && name != "zrefs.go"
}

func structEmbedsModel(st *ast.StructType) bool {
	if st == nil {
		return false
	}
	for _, field := range st.Fields.List {
		if field.Type == nil {
			continue
		}
		switch t := field.Type.(type) {
		case *ast.Ident:
			if t.Name == "Model" || t.Name == "ModelMeta" {
				return true
			}
		case *ast.SelectorExpr:
			if id, ok := t.X.(*ast.Ident); ok && (id.Name == "sdk" || id.Name == "modelmeta") && (t.Sel.Name == "Model" || t.Sel.Name == "ModelMeta") {
				return true
			}
		}
	}
	return false
}

func modelTagFromStruct(st *ast.StructType) string {
	technical, _ := modelSpecFromStruct(st)
	return technical
}

// modelSpecFromStruct returns the ORM model name and whether the struct extends via inherit=.
func modelSpecFromStruct(st *ast.StructType) (technical string, isExtend bool) {
	if st == nil {
		return "", false
	}
	for _, field := range st.Fields.List {
		if field.Type == nil {
			continue
		}
		switch field.Type.(type) {
		case *ast.Ident, *ast.SelectorExpr:
			if field.Tag == nil {
				continue
			}
			tag, ok := reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Lookup("sumeru")
			if !ok {
				continue
			}
			tags, err := modelmeta.ParseFieldTag(tag)
			if err != nil {
				continue
			}
			if tags.Inherit != "" {
				return tags.Inherit, true
			}
			if tags.Model != "" {
				return tags.Model, false
			}
		}
	}
	return "", false
}

func scanPackageForModels(dir string) ([]string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, isSourceGo, 0)
	if err != nil {
		return nil, err
	}
	var types []string
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
					types = append(types, ts.Name.Name)
				}
			}
		}
	}
	sort.Strings(types)
	return types, nil
}

func packageNameForDir(dir string) string {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, isSourceGo, parser.PackageClauseOnly)
	if err != nil || len(pkgs) == 0 {
		return "models"
	}
	for name := range pkgs {
		return name
	}
	return "models"
}
