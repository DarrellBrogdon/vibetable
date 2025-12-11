package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/formula"
	"github.com/vibetable/backend/internal/models"
)

// ComputedFieldService handles computation of formula, rollup, and lookup fields
type ComputedFieldService struct {
	db         DBTX
	evaluator  *formula.Evaluator
}

// NewComputedFieldService creates a new computed field service
func NewComputedFieldService(db DBTX) *ComputedFieldService {
	return &ComputedFieldService{
		db:        db,
		evaluator: formula.NewEvaluator(),
	}
}

// ComputeFieldsForRecords computes all computed field values for a list of records
func (s *ComputedFieldService) ComputeFieldsForRecords(ctx context.Context, records []models.Record, fields []models.Field) ([]models.Record, error) {
	if len(records) == 0 {
		return records, nil
	}

	// Separate computed fields
	var formulaFields, rollupFields, lookupFields []models.Field
	fieldMap := make(map[string]models.Field)

	for _, field := range fields {
		fieldMap[field.ID.String()] = field
		switch field.FieldType {
		case models.FieldTypeFormula:
			formulaFields = append(formulaFields, field)
		case models.FieldTypeRollup:
			rollupFields = append(rollupFields, field)
		case models.FieldTypeLookup:
			lookupFields = append(lookupFields, field)
		}
	}

	// If no computed fields, return as-is
	if len(formulaFields) == 0 && len(rollupFields) == 0 && len(lookupFields) == 0 {
		return records, nil
	}

	// Create a copy of records to avoid modifying the original
	result := make([]models.Record, len(records))
	for i, r := range records {
		result[i] = r

		// Parse values
		var values map[string]interface{}
		if err := json.Unmarshal(r.Values, &values); err != nil {
			values = make(map[string]interface{})
		}

		// Compute formula fields
		for _, field := range formulaFields {
			value, err := s.computeFormula(ctx, field, values, fieldMap)
			if err != nil {
				// Log error but continue - set to nil
				values[field.ID.String()] = nil
			} else {
				values[field.ID.String()] = value
			}
		}

		// Compute lookup fields first (rollups may depend on them)
		for _, field := range lookupFields {
			value, err := s.computeLookup(ctx, field, values, fieldMap)
			if err != nil {
				values[field.ID.String()] = nil
			} else {
				values[field.ID.String()] = value
			}
		}

		// Compute rollup fields
		for _, field := range rollupFields {
			value, err := s.computeRollup(ctx, field, values, fieldMap)
			if err != nil {
				values[field.ID.String()] = nil
			} else {
				values[field.ID.String()] = value
			}
		}

		// Marshal back to JSON
		newValues, err := json.Marshal(values)
		if err != nil {
			return nil, err
		}
		result[i].Values = newValues
	}

	return result, nil
}

// computeFormula evaluates a formula field
func (s *ComputedFieldService) computeFormula(ctx context.Context, field models.Field, values map[string]interface{}, fieldMap map[string]models.Field) (interface{}, error) {
	var opts models.FieldOptions
	if err := json.Unmarshal(field.Options, &opts); err != nil {
		return nil, fmt.Errorf("invalid formula options: %w", err)
	}

	if opts.Expression == nil || *opts.Expression == "" {
		return nil, nil
	}

	// Create a field resolver
	resolver := func(fieldRef string) (interface{}, error) {
		// Try to find field by name first
		for _, f := range fieldMap {
			if f.Name == fieldRef {
				return values[f.ID.String()], nil
			}
		}
		// Try by ID
		if _, ok := fieldMap[fieldRef]; ok {
			return values[fieldRef], nil
		}
		return nil, fmt.Errorf("field not found: %s", fieldRef)
	}

	result, err := s.evaluator.Evaluate(*opts.Expression, resolver)
	if err != nil {
		return nil, err
	}

	// Coerce result to expected type
	if opts.ResultType != nil {
		result = s.coerceResultType(result, *opts.ResultType)
	}

	return result, nil
}

// computeLookup gets values from linked records
func (s *ComputedFieldService) computeLookup(ctx context.Context, field models.Field, values map[string]interface{}, fieldMap map[string]models.Field) (interface{}, error) {
	var opts models.FieldOptions
	if err := json.Unmarshal(field.Options, &opts); err != nil {
		return nil, fmt.Errorf("invalid lookup options: %w", err)
	}

	if opts.LookupLinkedFieldID == nil || opts.LookupFieldID == nil {
		return nil, nil
	}

	// Get the linked record IDs from the linked field
	linkedRecordIDs := s.getLinkedRecordIDs(values[*opts.LookupLinkedFieldID])
	if len(linkedRecordIDs) == 0 {
		return nil, nil
	}

	// Fetch the lookup values from linked records
	lookupValues, err := s.fetchLookupValues(ctx, linkedRecordIDs, *opts.LookupFieldID)
	if err != nil {
		return nil, err
	}

	// Return single value if only one, otherwise array
	if len(lookupValues) == 1 {
		return lookupValues[0], nil
	}
	return lookupValues, nil
}

// computeRollup aggregates values from linked records
func (s *ComputedFieldService) computeRollup(ctx context.Context, field models.Field, values map[string]interface{}, fieldMap map[string]models.Field) (interface{}, error) {
	var opts models.FieldOptions
	if err := json.Unmarshal(field.Options, &opts); err != nil {
		return nil, fmt.Errorf("invalid rollup options: %w", err)
	}

	if opts.RollupLinkedFieldID == nil || opts.AggregationFunction == nil {
		return nil, nil
	}

	// Get the linked record IDs from the linked field
	linkedRecordIDs := s.getLinkedRecordIDs(values[*opts.RollupLinkedFieldID])
	if len(linkedRecordIDs) == 0 {
		// For COUNT, return 0 even if no records
		if strings.ToUpper(*opts.AggregationFunction) == "COUNT" {
			return 0.0, nil
		}
		return nil, nil
	}

	// For COUNT, we don't need to fetch values
	aggFunc := strings.ToUpper(*opts.AggregationFunction)
	if aggFunc == "COUNT" {
		return float64(len(linkedRecordIDs)), nil
	}

	// Fetch the values to aggregate
	if opts.RollupFieldID == nil {
		return nil, fmt.Errorf("rollup_field_id required for aggregation function %s", aggFunc)
	}

	lookupValues, err := s.fetchLookupValues(ctx, linkedRecordIDs, *opts.RollupFieldID)
	if err != nil {
		return nil, err
	}

	// Apply aggregation function
	return s.aggregate(aggFunc, lookupValues)
}

// getLinkedRecordIDs extracts record IDs from a linked record field value
func (s *ComputedFieldService) getLinkedRecordIDs(value interface{}) []uuid.UUID {
	var ids []uuid.UUID

	switch v := value.(type) {
	case []interface{}:
		for _, item := range v {
			if idStr, ok := item.(string); ok {
				if id, err := uuid.Parse(idStr); err == nil {
					ids = append(ids, id)
				}
			}
		}
	case []string:
		for _, idStr := range v {
			if id, err := uuid.Parse(idStr); err == nil {
				ids = append(ids, id)
			}
		}
	case string:
		if id, err := uuid.Parse(v); err == nil {
			ids = append(ids, id)
		}
	}

	return ids
}

// fetchLookupValues fetches field values from a list of records
func (s *ComputedFieldService) fetchLookupValues(ctx context.Context, recordIDs []uuid.UUID, fieldID string) ([]interface{}, error) {
	if len(recordIDs) == 0 {
		return nil, nil
	}

	// Build query to fetch values for all records
	rows, err := s.db.Query(ctx, `
		SELECT values->$1
		FROM records
		WHERE id = ANY($2)
	`, fieldID, recordIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []interface{}
	for rows.Next() {
		var valueJSON json.RawMessage
		if err := rows.Scan(&valueJSON); err != nil {
			return nil, err
		}

		var value interface{}
		if err := json.Unmarshal(valueJSON, &value); err == nil && value != nil {
			values = append(values, value)
		}
	}

	return values, rows.Err()
}

// aggregate applies an aggregation function to a list of values
func (s *ComputedFieldService) aggregate(aggFunc string, values []interface{}) (interface{}, error) {
	switch aggFunc {
	case "COUNT":
		return float64(len(values)), nil

	case "COUNTA":
		// Count non-empty values
		count := 0
		for _, v := range values {
			if v != nil && v != "" {
				count++
			}
		}
		return float64(count), nil

	case "SUM":
		sum := 0.0
		for _, v := range values {
			sum += toFloat(v)
		}
		return sum, nil

	case "AVG", "AVERAGE":
		if len(values) == 0 {
			return nil, nil
		}
		sum := 0.0
		for _, v := range values {
			sum += toFloat(v)
		}
		return sum / float64(len(values)), nil

	case "MIN":
		if len(values) == 0 {
			return nil, nil
		}
		min := toFloat(values[0])
		for _, v := range values[1:] {
			if f := toFloat(v); f < min {
				min = f
			}
		}
		return min, nil

	case "MAX":
		if len(values) == 0 {
			return nil, nil
		}
		max := toFloat(values[0])
		for _, v := range values[1:] {
			if f := toFloat(v); f > max {
				max = f
			}
		}
		return max, nil

	default:
		return nil, fmt.Errorf("unknown aggregation function: %s", aggFunc)
	}
}

// coerceResultType converts a result to the expected type
func (s *ComputedFieldService) coerceResultType(value interface{}, resultType string) interface{} {
	switch resultType {
	case "number":
		return toFloat(value)
	case "text":
		return fmt.Sprintf("%v", value)
	case "boolean":
		return toBool(value)
	default:
		return value
	}
}

// Helper functions

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	default:
		return 0
	}
}

func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case int:
		return val != 0
	case string:
		return val != "" && strings.ToLower(val) != "false"
	default:
		return v != nil
	}
}
