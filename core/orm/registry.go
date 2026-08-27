package orm

import (
	"context"
	"fmt"
	"strings"
)

var Registry = map[string]Model{}

// modelDeclaringModule maps each technical model name to the declaring addon (sys.module.name).
// Kernel / platform models use "base" so metadata and DDL follow one module graph.
var modelDeclaringModule = make(map[string]string)

// RegisterModelWithModule registers a model and records which addon owns its table (catalog module linkage).
func RegisterModelWithModule(model Model, declaringModule string) {
	if model == nil {
		return
	}
	name := model.ModelName()
	if err := ValidateModelName(name); err != nil {
		panic(fmt.Sprintf("RegisterModelWithModule: %v", err))
	}
	for _, f := range model.Fields() {
		if f.Name == "" || f.Name == "id" {
			continue
		}
		if err := ValidateFieldName(f.Name); err != nil {
			panic(fmt.Sprintf("RegisterModelWithModule %s: %v", name, err))
		}
	}
	Registry[name] = model
	modelDeclaringModule[name] = strings.TrimSpace(declaringModule)
}

// RegistryModel returns a registered model by technical name, or nil.
func RegistryModel(name string) Model {
	return Registry[name]
}

func SyncModels() error {
	if DB == nil {
		return nil
	}
	ctx := ContextWithBypass(context.Background(), true)
	installed, err := InstalledModuleNames(ctx)
	if err != nil {
		return err
	}
	for _, model := range Registry {
		name := model.ModelName()
		if len(installed) == 0 {
			owner := DeclaringModule(name)
			if owner != "" && !IsPlatformModule(owner) {
				continue
			}
		} else if !ShouldMaterializeModel(name, installed) {
			continue
		}
		if err := createTable(model); err != nil {
			return err
		}
	}
	return nil
}

// ColumnTypeSQL returns the PostgreSQL column type fragment for f (without NOT NULL / UNIQUE / DEFAULT).
// Many2One is stored as BIGINT. Unknown types return ok == false.
func ColumnTypeSQL(f FieldDefinition) (string, bool) {
	switch f.Type {
	case Char:
		size := f.Size
		if size <= 0 {
			size = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", size), true
	case Text:
		return "TEXT", true
	case Integer:
		return "BIGINT", true
	case Float:
		return "REAL", true
	case Float64:
		return "DOUBLE PRECISION", true
	case Numeric:
		prec, scale := f.Precision, f.Scale
		if prec <= 0 {
			prec = 16
		}
		if scale < 0 {
			scale = 4
		}
		return fmt.Sprintf("NUMERIC(%d, %d)", prec, scale), true
	case Boolean:
		return "BOOLEAN", true
	case Date:
		return "DATE", true
	case DateTime:
		return "TIMESTAMPTZ", true
	case Selection:
		return "VARCHAR(50)", true
	case Json:
		return "JSONB", true
	case Many2One:
		return "BIGINT", true
	default:
		return "", false
	}
}

func createTable(model Model) error {
	physical, err := ModelToTableName(model.ModelName())
	if err != nil {
		return err
	}
	tableName, err := QuotedTableName(model.ModelName())
	if err != nil {
		return err
	}
	exists, err := tableExists(physical)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	var columns []string
	columns = append(columns, quoteIdent("id")+" BIGSERIAL PRIMARY KEY")

	for _, field := range model.Fields() {
		if field.Name == "id" || IsVirtualField(field) {
			continue
		}
		baseType, ok := ColumnTypeSQL(field)
		if !ok {
			continue
		}
		if err := ValidateFieldName(field.Name); err != nil {
			return err
		}
		colType := baseType
		if field.Required {
			colType += " NOT NULL"
		}
		if field.Unique {
			colType += " UNIQUE"
		}

		defVal := field.DefaultVal
		if defVal != nil {
			if s, ok := defVal.(string); ok {
				if strings.Contains(s, ";") || strings.Contains(s, "--") || strings.Contains(s, "/*") {
					return fmt.Errorf("unsafe string default on %s.%s", model.ModelName(), field.Name)
				}
			}
			switch defVal.(type) {
			case string, bool, int, int64, float64:
				if lit, ok := sqlDefaultLiteral(defVal); ok {
					colType += " DEFAULT " + lit
				}
			}
		}

		columns = append(columns, fmt.Sprintf("%s %s", quoteIdent(field.Name), colType))
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, strings.Join(columns, ", "))
	if _, err := DB.Exec(query); err != nil {
		return err
	}
	return ensureModelIndexes(model.ModelName(), physical, tableName, model)
}
