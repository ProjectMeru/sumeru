package orm

import (
	"context"
	"fmt"
	"strings"
)

type ctxKeySkipRelated struct{}

// ContextSkipRelatedEnrichment prevents ApplyRelatedFields during nested Search used for batching.
func ContextSkipRelatedEnrichment(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeySkipRelated{}, true)
}

func skipRelatedEnrichment(ctx context.Context) bool {
	v, _ := ctx.Value(ctxKeySkipRelated{}).(bool)
	return v
}

// ApplyRelatedFields fills virtual related fields on rec (in place).
func ApplyRelatedFields(ctx context.Context, model string, rec map[string]interface{}) error {
	if rec == nil {
		return nil
	}
	return ApplyRelatedFieldsBatch(ctx, model, []map[string]interface{}{rec})
}

// ApplyRelatedFieldsBatch resolves related fields for many rows with one Search per relation hop.
func ApplyRelatedFieldsBatch(ctx context.Context, model string, records []map[string]interface{}) error {
	if len(records) == 0 {
		return nil
	}
	inst, ok := Registry[model]
	if !ok || inst == nil {
		return nil
	}
	type relatedSpec struct {
		fieldName   string
		relField    string
		targetField string
		relation    string
	}
	var specs []relatedSpec
	for _, fieldDef := range inst.Fields() {
		if fieldDef.Related == "" || fieldDef.RelatedStore {
			continue
		}
		parts := strings.Split(fieldDef.Related, ".")
		if len(parts) < 2 {
			return fmt.Errorf("invalid related path %q", fieldDef.Related)
		}
		relField := strings.TrimSpace(parts[0])
		targetField := strings.TrimSpace(parts[1])
		relationFieldDef := FieldDef(model, relField)
		if relationFieldDef == nil || relationFieldDef.Relation == "" {
			return fmt.Errorf("relation field %q not found on %s", relField, model)
		}
		specs = append(specs, relatedSpec{
			fieldName:   fieldDef.Name,
			relField:    relField,
			targetField: targetField,
			relation:    relationFieldDef.Relation,
		})
	}
	if len(specs) == 0 {
		return nil
	}

	// relation model -> ids to fetch
	idsByRelation := map[string]map[int64]struct{}{}
	for _, spec := range specs {
		set := idsByRelation[spec.relation]
		if set == nil {
			set = map[int64]struct{}{}
			idsByRelation[spec.relation] = set
		}
		for _, rec := range records {
			if id, ok := CoerceInt64(rec[spec.relField]); ok && id > 0 {
				set[id] = struct{}{}
			}
		}
	}

	ctx = ContextSkipRelatedEnrichment(ctx)
	cacheByRelation := map[string]map[int64]map[string]interface{}{}
	for relation, idSet := range idsByRelation {
		if len(idSet) == 0 {
			continue
		}
		values := make([]interface{}, 0, len(idSet))
		for id := range idSet {
			values = append(values, int(id))
		}
		rows, err := Search(ctx, relation, [][]interface{}{{"id", "in", values}})
		if err != nil {
			return err
		}
		byID := make(map[int64]map[string]interface{}, len(rows))
		for _, row := range rows {
			id, ok := CoerceInt64(row["id"])
			if !ok {
				continue
			}
			byID[id] = row
		}
		cacheByRelation[relation] = byID
	}

	for _, rec := range records {
		for _, spec := range specs {
			relID, ok := CoerceInt64(rec[spec.relField])
			if !ok || relID <= 0 {
				rec[spec.fieldName] = nil
				continue
			}
			target := cacheByRelation[spec.relation][relID]
			if target == nil {
				rec[spec.fieldName] = nil
				continue
			}
			rec[spec.fieldName] = target[spec.targetField]
		}
	}
	return nil
}

func resolveRelatedValue(ctx context.Context, model string, rec map[string]interface{}, path string) (interface{}, error) {
	// Kept for tests/callers; batch path is preferred.
	parts := strings.Split(path, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid related path %q", path)
	}
	tmp := map[string]interface{}{}
	for k, v := range rec {
		tmp[k] = v
	}
	// Find field name that has this related path
	inst, ok := Registry[model]
	if !ok {
		return nil, fmt.Errorf("model %s not found", model)
	}
	var fieldName string
	for _, fd := range inst.Fields() {
		if fd.Related == path {
			fieldName = fd.Name
			break
		}
	}
	if fieldName == "" {
		// synthesize
		relField := strings.TrimSpace(parts[0])
		targetField := strings.TrimSpace(parts[1])
		relationFieldDef := FieldDef(model, relField)
		if relationFieldDef == nil {
			return nil, fmt.Errorf("relation field %q not found on %s", relField, model)
		}
		relID, ok := CoerceInt64(rec[relField])
		if !ok || relID <= 0 {
			return nil, nil
		}
		target, err := SearchOne(ContextSkipRelatedEnrichment(ctx), relationFieldDef.Relation, map[string]interface{}{"id": int(relID)})
		if err != nil {
			return nil, err
		}
		if target == nil {
			return nil, nil
		}
		return target[targetField], nil
	}
	if err := ApplyRelatedFieldsBatch(ctx, model, []map[string]interface{}{tmp}); err != nil {
		return nil, err
	}
	return tmp[fieldName], nil
}
