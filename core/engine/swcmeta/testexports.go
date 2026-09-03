package swcmeta

import (
	"context"

	"sumeru/core/engine/parser"
	"reflect"
)

func SerializeSheetForTest(ctx context.Context, model string, s *parser.Sheet) *ArchSheet {
	return serializeSheet(ctx, model, s)
}

func SerializeGroupForTest(ctx context.Context, model string, g parser.Group) ArchGroup {
	return serializeGroup(ctx, model, g)
}

func SerializeDivForTest(ctx context.Context, model string, d parser.Div) ArchDiv {
	return serializeDiv(ctx, model, d)
}

func FormMetaForModelForTest(model string) *FormMeta { return formMetaForModel(model) }

func SerializeFieldsForTest(ctx context.Context, fields []parser.Field) []ArchField {
	return serializeFields(ctx, fields)
}

func EnrichFieldForTest(model string, f ArchField) ArchField { return enrichField(model, f) }

func WorkspacePayloadTypeForTest() reflect.Type { return reflect.TypeOf(WorkspacePayload{}) }
