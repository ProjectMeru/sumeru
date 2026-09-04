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

type scannedModel struct {
	GoName    string
	ModelName string
	Extend    bool
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

func embeddedModelMetaTag(st *ast.StructType) (string, bool) {
	if st == nil {
		return "", false
	}
	for _, field := range st.Fields.List {
		if field.Type == nil || field.Tag == nil {
			continue
		}
		switch field.Type.(type) {
		case *ast.Ident, *ast.SelectorExpr:
			tag, ok := reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Lookup("sumeru")
			if !ok {
				continue
			}
			return tag, true
		}
	}
	return "", false
}

// modelSpecFromStruct returns the ORM model spec aligned with modelmeta.ModelSpecFromStruct.
func modelSpecFromStruct(st *ast.StructType, goName string) (modelmeta.ModelSpec, error) {
	if st == nil {
		return modelmeta.ModelSpec{}, nil
	}
	tag, ok := embeddedModelMetaTag(st)
	if !ok {
		return modelmeta.ModelSpec{Name: modelmeta.ModelNameFromGo(goName), Extend: false}, nil
	}
	tags, err := modelmeta.ParseFieldTag(tag)
	if err != nil {
		return modelmeta.ModelSpec{}, err
	}
	return modelmeta.ModelSpecFromTags(tags, goName)
}

func scanPackageModels(pkgs map[string]*ast.Package) []scannedModel { //nolint:staticcheck // SA1019: ParseDir/ast.Package adequate for model tag scan; go/packages migration is separate
	var out []scannedModel
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
					modelSpec, err := modelSpecFromStruct(st, ts.Name.Name)
					if err != nil {
						continue
					}
					out = append(out, scannedModel{
						GoName:    ts.Name.Name,
						ModelName: modelSpec.Name,
						Extend:    modelSpec.Extend,
					})
				}
			}
		}
	}
	return out
}

func parseDirModels(dir string) ([]scannedModel, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, isSourceGo, 0) //nolint:staticcheck // SA1019: see scanPackageModels
	if err != nil {
		return nil, err
	}
	return scanPackageModels(pkgs), nil
}

func scanPackageForModels(dir string) ([]string, error) {
	models, err := parseDirModels(dir)
	if err != nil {
		return nil, err
	}
	var types []string
	for _, m := range models {
		if m.ModelName == "-" {
			continue
		}
		types = append(types, m.GoName)
	}
	sort.Strings(types)
	return types, nil
}

func packageNameForDir(dir string) string {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, isSourceGo, parser.PackageClauseOnly) //nolint:staticcheck // SA1019: see scanPackageModels
	if err != nil || len(pkgs) == 0 {
		return "models"
	}
	for name := range pkgs {
		return name
	}
	return "models"
}
