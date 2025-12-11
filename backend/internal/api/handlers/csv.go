package handlers

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/store"
)

type CSVHandler struct {
	recordStore *store.RecordStore
	fieldStore  *store.FieldStore
	tableStore  *store.TableStore
}

func NewCSVHandler(recordStore *store.RecordStore, fieldStore *store.FieldStore, tableStore *store.TableStore) *CSVHandler {
	return &CSVHandler{
		recordStore: recordStore,
		fieldStore:  fieldStore,
		tableStore:  tableStore,
	}
}

// CSVPreviewResponse contains preview data from an uploaded CSV
type CSVPreviewResponse struct {
	Columns []string            `json:"columns"`
	Rows    []map[string]string `json:"rows"`
	Total   int                 `json:"total"`
}

// CSVImportRequest contains the mapping configuration for import
type CSVImportRequest struct {
	Data     string            `json:"data"`     // CSV content as string
	Mappings map[string]string `json:"mappings"` // column name -> field ID
}

// CSVImportResponse contains the result of an import
type CSVImportResponse struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
	Errors   int `json:"errors"`
}

// Preview handles POST /tables/:tableId/import/preview
// Accepts CSV data and returns a preview of columns and first few rows
func (h *CSVHandler) Preview(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	tableID, err := uuid.Parse(chi.URLParam(r, "tableId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid table ID")
		return
	}

	// Verify user has access to this table
	_, err = h.fieldStore.ListFieldsForTable(r.Context(), tableID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found")
			return
		}
		log.Printf("Error checking table access: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to verify access")
		return
	}

	// Parse multipart form (max 10MB)
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		// Try parsing as JSON with data field
		var req struct {
			Data string `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request format")
			return
		}

		preview, err := h.parseCSVPreview(req.Data, 5)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_csv", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, preview)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "no_file", "No file uploaded")
		return
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "read_error", "Failed to read file")
		return
	}

	preview, err := h.parseCSVPreview(string(content), 5)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_csv", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, preview)
}

// Import handles POST /tables/:tableId/import
// Accepts CSV data with column->field mappings and creates records
func (h *CSVHandler) Import(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	tableID, err := uuid.Parse(chi.URLParam(r, "tableId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid table ID")
		return
	}

	// Get table fields to validate mappings
	fields, err := h.fieldStore.ListFieldsForTable(r.Context(), tableID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found")
			return
		}
		log.Printf("Error listing fields: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get table fields")
		return
	}

	// Build field lookup
	fieldMap := make(map[string]models.Field)
	for _, f := range fields {
		fieldMap[f.ID.String()] = f
	}

	var req CSVImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if req.Data == "" {
		writeError(w, http.StatusBadRequest, "no_data", "CSV data is required")
		return
	}

	// Parse CSV
	reader := csv.NewReader(bytes.NewBufferString(req.Data))
	allRows, err := reader.ReadAll()
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_csv", "Failed to parse CSV")
		return
	}

	if len(allRows) < 2 {
		writeError(w, http.StatusBadRequest, "no_data", "CSV must have at least a header row and one data row")
		return
	}

	headers := allRows[0]
	dataRows := allRows[1:]

	// Build header index
	headerIndex := make(map[string]int)
	for i, h := range headers {
		headerIndex[h] = i
	}

	// Prepare records
	var recordValues []json.RawMessage
	skipped := 0
	errCount := 0

	for _, row := range dataRows {
		values := make(map[string]interface{})
		hasData := false

		for colName, fieldID := range req.Mappings {
			if fieldID == "" {
				continue // Skip unmapped columns
			}

			colIdx, ok := headerIndex[colName]
			if !ok || colIdx >= len(row) {
				continue
			}

			field, ok := fieldMap[fieldID]
			if !ok {
				continue
			}

			cellValue := row[colIdx]
			if cellValue == "" {
				continue
			}

			hasData = true
			convertedValue := h.convertCellValue(cellValue, field)
			values[fieldID] = convertedValue
		}

		if !hasData {
			skipped++
			continue
		}

		jsonValues, err := json.Marshal(values)
		if err != nil {
			errCount++
			continue
		}
		recordValues = append(recordValues, jsonValues)
	}

	// Bulk create records
	if len(recordValues) > 0 {
		_, err = h.recordStore.BulkCreateRecords(r.Context(), tableID, recordValues, user.ID)
		if err != nil {
			if errors.Is(err, store.ErrForbidden) {
				writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to create records")
				return
			}
			log.Printf("Error bulk creating records: %v", err)
			writeError(w, http.StatusInternalServerError, "server_error", "Failed to create records")
			return
		}
	}

	writeJSON(w, http.StatusOK, CSVImportResponse{
		Imported: len(recordValues),
		Skipped:  skipped,
		Errors:   errCount,
	})
}

// Export handles GET /tables/:tableId/export
// Returns table data as CSV
func (h *CSVHandler) Export(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	tableID, err := uuid.Parse(chi.URLParam(r, "tableId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid table ID")
		return
	}

	// Get table info for filename
	table, err := h.tableStore.GetTable(r.Context(), tableID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found")
			return
		}
		log.Printf("Error getting table: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get table")
		return
	}

	// Get fields
	fields, err := h.fieldStore.ListFieldsForTable(r.Context(), tableID, user.ID)
	if err != nil {
		log.Printf("Error listing fields: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get fields")
		return
	}

	// Get records
	records, err := h.recordStore.ListRecordsForTable(r.Context(), tableID, user.ID)
	if err != nil {
		log.Printf("Error listing records: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get records")
		return
	}

	// Build CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := make([]string, len(fields))
	for i, f := range fields {
		header[i] = f.Name
	}
	if err := writer.Write(header); err != nil {
		log.Printf("Error writing CSV header: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to generate CSV")
		return
	}

	// Write data rows
	for _, record := range records {
		var recordValues map[string]interface{}
		if err := json.Unmarshal(record.Values, &recordValues); err != nil {
			recordValues = make(map[string]interface{})
		}

		row := make([]string, len(fields))
		for i, f := range fields {
			val := recordValues[f.ID.String()]
			row[i] = h.formatCellValue(val, f)
		}
		if err := writer.Write(row); err != nil {
			log.Printf("Error writing CSV row: %v", err)
			continue
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Printf("CSV writer error: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to generate CSV")
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("%s.csv", table.Name)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

// parseCSVPreview parses CSV content and returns a preview
func (h *CSVHandler) parseCSVPreview(content string, maxRows int) (*CSVPreviewResponse, error) {
	reader := csv.NewReader(bytes.NewBufferString(content))

	allRows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(allRows) == 0 {
		return nil, fmt.Errorf("CSV is empty")
	}

	columns := allRows[0]
	dataRows := allRows[1:]
	total := len(dataRows)

	// Limit preview rows
	if len(dataRows) > maxRows {
		dataRows = dataRows[:maxRows]
	}

	rows := make([]map[string]string, 0, len(dataRows))
	for _, row := range dataRows {
		rowMap := make(map[string]string)
		for i, col := range columns {
			if i < len(row) {
				rowMap[col] = row[i]
			} else {
				rowMap[col] = ""
			}
		}
		rows = append(rows, rowMap)
	}

	return &CSVPreviewResponse{
		Columns: columns,
		Rows:    rows,
		Total:   total,
	}, nil
}

// convertCellValue converts a string cell value to the appropriate type based on field type
func (h *CSVHandler) convertCellValue(value string, field models.Field) interface{} {
	switch field.FieldType {
	case models.FieldTypeNumber:
		// Try to parse as float
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
		return value
	case models.FieldTypeCheckbox:
		// Accept various true/false representations
		lower := value
		return lower == "true" || lower == "1" || lower == "yes" || lower == "TRUE" || lower == "Yes" || lower == "Y" || lower == "y"
	case models.FieldTypeDate:
		// Return date string as-is (frontend will handle parsing)
		return value
	case models.FieldTypeLinkedRecord:
		// For linked records, try to parse as JSON array
		var ids []string
		if err := json.Unmarshal([]byte(value), &ids); err == nil {
			return ids
		}
		// If not JSON, treat as single ID
		if value != "" {
			return []string{value}
		}
		return []string{}
	default:
		return value
	}
}

// formatCellValue formats a cell value for CSV export
func (h *CSVHandler) formatCellValue(value interface{}, field models.Field) string {
	if value == nil {
		return ""
	}

	switch field.FieldType {
	case models.FieldTypeCheckbox:
		if b, ok := value.(bool); ok {
			if b {
				return "true"
			}
			return "false"
		}
	case models.FieldTypeNumber:
		if f, ok := value.(float64); ok {
			return strconv.FormatFloat(f, 'f', -1, 64)
		}
	case models.FieldTypeLinkedRecord:
		// Export linked records as JSON array
		if arr, ok := value.([]interface{}); ok {
			if b, err := json.Marshal(arr); err == nil {
				return string(b)
			}
		}
	}

	return fmt.Sprintf("%v", value)
}
