# VibeTable Feature Implementation Plan

This plan covers 18 features to expand VibeTable beyond the MVP, organized into implementation phases based on dependencies and complexity.

---

## 📊 Completion Status

| Phase | Status | Features |
|-------|--------|----------|
| Phase 1: Quick Wins | ✅ COMPLETE | 5/5 |
| Phase 2: New View Types | ✅ COMPLETE | 4/4 |
| Phase 3: Collaboration Features | ✅ COMPLETE | 2/2 |
| Phase 4: Advanced Field Types | ✅ COMPLETE | 4/4 |
| Phase 5: Real-time & Automation | ⏳ IN PROGRESS | 2/3 |

**Overall Progress: 17/18 features complete (94%)**

---

## Architecture Overview

Before diving into features, here's how they map to the existing architecture:

### Backend Extensions Needed

| Feature Category | New Stores | New Handlers | New Models | Migrations |
|-----------------|------------|--------------|------------|------------|
| CSV Import/Export | - | CSVHandler | - | - |
| Duplicate Base/Table | - | Clone methods | - | - |
| Record Color Coding | - | - | ViewCondition | Yes |
| Undo/Redo | ActionStore | ActionHandler | Action | Yes |
| Form View | FormStore | FormHandler | Form, FormField | Yes |
| Calendar/Gallery View | - | - | ViewConfig updates | - |
| Comments | CommentStore | CommentHandler | Comment | Yes |
| Activity Log | ActivityStore | ActivityHandler | Activity | Yes |
| Public Views | - | PublicViewHandler | ViewShare | Yes |
| Formula Fields | - | - | FieldType expansion | - |
| Rollup/Lookup Fields | - | - | FieldType expansion | - |
| File Attachments | AttachmentStore | AttachmentHandler | Attachment | Yes |
| Real-time | - | WebSocket handler | - | - |
| Automations | AutomationStore | AutomationHandler | Automation, Trigger, Action | Yes |
| Webhooks/API Keys | APIKeyStore, WebhookStore | APIHandler | APIKey, Webhook | Yes |

### Frontend Extensions Needed

| Feature Category | New Components | Store Changes | New Routes |
|-----------------|----------------|---------------|------------|
| CSV Import/Export | ImportModal, ExportButton | - | - |
| Duplicate | DuplicateModal | - | - |
| Color Coding | ConditionalFormatModal | viewStore | - |
| Keyboard Shortcuts | KeyboardHandler | - | - |
| Undo/Redo | - | undoStore | - |
| Form View | FormBuilder, FormRenderer | formStore | /forms/[id] |
| Calendar View | Calendar.svelte | - | - |
| Gallery View | Gallery.svelte | - | - |
| Comments | CommentThread.svelte | commentStore | - |
| Activity Log | ActivityFeed.svelte | - | - |
| Public Views | ShareModal updates | - | /public/[token] |
| Formula Fields | FormulaEditor.svelte | - | - |
| Rollup/Lookup | FieldConfigModal updates | - | - |
| Attachments | FileUploader, FilePreviewer | - | - |
| Real-time | PresenceIndicator | realtimeStore | - |
| Automations | AutomationBuilder | automationStore | /automations |
| Webhooks | APIKeyManager, WebhookConfig | - | /settings/api |

---

## Phase 1: Quick Wins (Low Effort, High Value) ✅ COMPLETE

### 1.1 CSV Import/Export ✅

**Backend Implementation:**

1. Create `/backend/internal/api/handlers/csv.go`:
```go
type CSVHandler struct {
    recordStore *store.RecordStore
    fieldStore  *store.FieldStore
    tableStore  *store.TableStore
}

// POST /tables/:tableId/import - Upload CSV
// GET /tables/:tableId/export - Download CSV (respects view filters)
```

2. Import flow:
   - Parse CSV with `encoding/csv`
   - Return preview (first 5 rows + detected columns)
   - Accept column→field mapping
   - Batch insert records (use existing `BulkCreateRecords` pattern)

3. Export flow:
   - Accept optional `viewId` query param for filters
   - Query records with filters applied
   - Stream CSV response with proper headers

**API Endpoints:**
```
POST /tables/:tableId/import/preview    # Upload CSV, get column preview
POST /tables/:tableId/import            # Execute import with mapping
GET  /tables/:tableId/export            # Download CSV
GET  /tables/:tableId/export?viewId=x   # Download filtered CSV
```

**Frontend Implementation:**

1. Create `ImportModal.svelte`:
   - Drag-and-drop file upload
   - Column mapping interface (dropdown for each CSV column → table field)
   - Preview of data to import
   - Progress indicator for large files

2. Add export button to Grid toolbar:
   - Simple download link with auth token
   - Option to export current view (with filters) vs. all data

**Data Flow:**
```
User uploads CSV → Preview endpoint parses first N rows
User maps columns → Import endpoint validates types and creates records
Records created → Grid refreshes via existing pattern
```

---

### 1.2 Keyboard Shortcuts ✅

**Frontend Implementation:**

1. Create `KeyboardHandler.svelte` (or action):
   - Global keydown listener
   - Context-aware shortcuts (different in grid vs modal)
   - Prevent default browser behavior where needed

2. Shortcut map:
```typescript
const shortcuts = {
    grid: {
        'ArrowUp': () => moveSelection(0, -1),
        'ArrowDown': () => moveSelection(0, 1),
        'ArrowLeft': () => moveSelection(-1, 0),
        'ArrowRight': () => moveSelection(1, 0),
        'Enter': () => startEditing(),
        'Escape': () => cancelEditing(),
        'Tab': () => moveSelection(1, 0),
        'Shift+Tab': () => moveSelection(-1, 0),
        'Meta+c': () => copySelection(),
        'Meta+v': () => pasteSelection(),
        'Meta+z': () => undo(),
        'Meta+Shift+z': () => redo(),
        'Delete': () => clearCell(),
        'Backspace': () => clearCell(),
    },
    global: {
        'Meta+/': () => showShortcutModal(),
        'Escape': () => closeModal(),
    }
};
```

3. Grid.svelte changes:
   - Track selected cell position (already exists)
   - Add `tabindex="0"` for focus management
   - Emit events for copy/paste operations

4. Create `ShortcutsModal.svelte`:
   - Display all available shortcuts
   - Group by category (navigation, editing, selection)

**Considerations:**
- Mac vs Windows key detection (`navigator.platform`)
- Prevent shortcuts when typing in input fields
- Accessibility: ensure focus states are visible

---

### 1.3 Undo/Redo ✅

**Backend Implementation:**

1. Create migration `008_create_actions.sql`:
```sql
CREATE TABLE actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_id UUID NOT NULL REFERENCES bases(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action_type VARCHAR(50) NOT NULL, -- 'create_record', 'update_record', etc.
    entity_type VARCHAR(50) NOT NULL, -- 'record', 'field', 'table'
    entity_id UUID NOT NULL,
    previous_state JSONB, -- State before action (for undo)
    new_state JSONB,      -- State after action (for redo)
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_actions_base_user ON actions(base_id, user_id, created_at DESC);
```

2. Create `/backend/internal/store/action.go`:
```go
type ActionStore struct {
    db DBTX
}

func (s *ActionStore) RecordAction(ctx context.Context, action *models.Action) error
func (s *ActionStore) GetUndoStack(ctx context.Context, baseID, userID uuid.UUID, limit int) ([]models.Action, error)
func (s *ActionStore) GetRedoStack(ctx context.Context, baseID, userID uuid.UUID, limit int) ([]models.Action, error)
```

3. Modify existing stores to record actions:
   - RecordStore: Capture state before/after create, update, delete
   - FieldStore: Same pattern
   - Wrap modifications in action recording

4. Create `/backend/internal/api/handlers/action.go`:
```go
// POST /bases/:baseId/undo - Undo last action
// POST /bases/:baseId/redo - Redo last undone action
```

**API Endpoints:**
```
POST /bases/:baseId/undo    # Execute undo, return restored state
POST /bases/:baseId/redo    # Execute redo, return new state
GET  /bases/:baseId/actions # List recent actions (optional, for history view)
```

**Frontend Implementation:**

1. Create `undoStore.ts`:
```typescript
interface UndoState {
    canUndo: boolean;
    canRedo: boolean;
}

function createUndoStore() {
    const { subscribe, set } = writable<UndoState>({ canUndo: false, canRedo: false });

    return {
        subscribe,
        async undo(baseId: string) { ... },
        async redo(baseId: string) { ... },
        updateState(canUndo: boolean, canRedo: boolean) { ... }
    };
}
```

2. Wire keyboard shortcuts (Cmd+Z, Cmd+Shift+Z) to undo/redo actions

3. Add undo/redo buttons to toolbar (optional)

**Considerations:**
- Stack depth limit (50 actions default)
- Cleanup old actions (cron job or on insert)
- Handle conflicts when undoing collaborative edits
- Batch related actions (e.g., bulk delete should undo as one)

---

### 1.4 Duplicate Base/Table ✅

**Backend Implementation:**

1. Add methods to existing stores:

`/backend/internal/store/table.go`:
```go
func (s *TableStore) DuplicateTable(ctx context.Context, tableID uuid.UUID,
    userID uuid.UUID, includeRecords bool) (*models.Table, error) {
    // 1. Get original table
    // 2. Create new table with name "Copy of X"
    // 3. Copy all fields
    // 4. If includeRecords, copy all records (map old field IDs to new)
    // 5. Copy all views
    return newTable, nil
}
```

`/backend/internal/store/base.go`:
```go
func (s *BaseStore) DuplicateBase(ctx context.Context, baseID uuid.UUID,
    userID uuid.UUID, includeRecords bool) (*models.Base, error) {
    // 1. Get original base
    // 2. Create new base with name "Copy of X"
    // 3. For each table, call DuplicateTable
    // 4. Handle linked records (remap IDs to new tables)
    return newBase, nil
}
```

2. Add endpoints to existing handlers:

`/backend/internal/api/handlers/table.go`:
```go
// POST /tables/:id/duplicate
func (h *TableHandler) DuplicateTable(w http.ResponseWriter, r *http.Request)
```

`/backend/internal/api/handlers/base.go`:
```go
// POST /bases/:id/duplicate
func (h *BaseHandler) DuplicateBase(w http.ResponseWriter, r *http.Request)
```

**API Endpoints:**
```
POST /tables/:id/duplicate              # Body: { "includeRecords": true }
POST /bases/:id/duplicate               # Body: { "includeRecords": true }
```

**Frontend Implementation:**

1. Add "Duplicate" option to table/base context menus
2. Create simple `DuplicateModal.svelte`:
   - New name input (prefilled with "Copy of X")
   - Checkbox: Include records
   - Progress indicator for large bases

---

### 1.5 Record Color Coding ✅

**Backend Implementation:**

1. Extend ViewConfig in models:
```go
type ViewConfig struct {
    Filters         []ViewFilter         `json:"filters,omitempty"`
    Sorts           []ViewSort           `json:"sorts,omitempty"`
    GroupByFieldID  string               `json:"groupByFieldId,omitempty"`
    VisibleFields   []string             `json:"visibleFields,omitempty"`
    ConditionalColors []ConditionalColor `json:"conditionalColors,omitempty"` // NEW
}

type ConditionalColor struct {
    ID        string       `json:"id"`
    FieldID   string       `json:"fieldId"`
    Operator  string       `json:"operator"` // 'equals', 'contains', 'isEmpty', etc.
    Value     interface{}  `json:"value"`
    Color     string       `json:"color"` // hex or preset name
    ColorType string       `json:"colorType"` // 'row' or 'indicator'
}
```

No new migrations needed - ViewConfig is already JSONB.

**Frontend Implementation:**

1. Create `ConditionalFormatModal.svelte`:
   - List of existing rules
   - Add rule: Select field → Select operator → Enter value → Pick color
   - Preview of affected rows
   - Drag to reorder (first match wins)

2. Modify `Grid.svelte`:
   - Evaluate conditional colors for each row
   - Apply background color or left indicator bar
   - Cache evaluation results for performance

```typescript
function evaluateRowColor(record: Record, rules: ConditionalColor[]): string | null {
    for (const rule of rules) {
        if (evaluateCondition(record.values[rule.fieldId], rule.operator, rule.value)) {
            return rule.color;
        }
    }
    return null;
}
```

3. Color palette:
   - Predefined Airtable-like colors: red, orange, yellow, green, blue, purple, pink, gray
   - Each with light variant for row background

---

## Phase 2: New View Types (Medium Effort, High Visual Impact) ✅ COMPLETE

### 2.1 Form View ✅

**Backend Implementation:**

1. Create migration `009_create_forms.sql`:
```sql
CREATE TABLE forms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id UUID NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    public_token VARCHAR(64) UNIQUE, -- For public access
    is_active BOOLEAN DEFAULT true,
    success_message TEXT DEFAULT 'Thank you for your submission!',
    redirect_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE form_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    field_id UUID NOT NULL REFERENCES fields(id) ON DELETE CASCADE,
    label VARCHAR(255), -- Override field name
    help_text TEXT,
    is_required BOOLEAN DEFAULT false,
    position INT NOT NULL,
    UNIQUE(form_id, field_id)
);

CREATE INDEX idx_forms_table ON forms(table_id);
CREATE INDEX idx_forms_token ON forms(public_token);
```

2. Create `/backend/internal/store/form.go`:
```go
type FormStore struct {
    db          DBTX
    tableStore  *TableStore
    recordStore *RecordStore
}

func (s *FormStore) CreateForm(ctx context.Context, tableID uuid.UUID, userID uuid.UUID, name string) (*models.Form, error)
func (s *FormStore) GetForm(ctx context.Context, formID uuid.UUID, userID uuid.UUID) (*models.Form, error)
func (s *FormStore) GetFormByToken(ctx context.Context, token string) (*models.Form, error) // Public access
func (s *FormStore) UpdateForm(ctx context.Context, formID uuid.UUID, userID uuid.UUID, updates FormUpdates) (*models.Form, error)
func (s *FormStore) UpdateFormFields(ctx context.Context, formID uuid.UUID, userID uuid.UUID, fields []FormFieldConfig) error
func (s *FormStore) SubmitForm(ctx context.Context, token string, values map[string]interface{}) (*models.Record, error)
```

3. Create `/backend/internal/api/handlers/form.go`:
```go
type FormHandler struct {
    store *store.FormStore
}

// Protected endpoints (require auth)
// GET    /tables/:tableId/forms
// POST   /tables/:tableId/forms
// GET    /forms/:id
// PATCH  /forms/:id
// DELETE /forms/:id
// PATCH  /forms/:id/fields

// Public endpoint (no auth required)
// GET    /public/forms/:token       # Get form config
// POST   /public/forms/:token       # Submit form
```

**API Endpoints:**
```
# Authenticated
GET    /tables/:tableId/forms           # List forms for table
POST   /tables/:tableId/forms           # Create form
GET    /forms/:id                       # Get form with fields
PATCH  /forms/:id                       # Update form settings
DELETE /forms/:id                       # Delete form
PATCH  /forms/:id/fields                # Update field configuration

# Public (no auth)
GET    /public/forms/:token             # Get form for display
POST   /public/forms/:token             # Submit form response
```

**Frontend Implementation:**

1. Create `FormBuilder.svelte`:
   - Drag-and-drop field ordering
   - Toggle fields on/off
   - Edit labels and help text
   - Set required fields
   - Preview mode
   - Settings: success message, redirect URL

2. Create `FormRenderer.svelte`:
   - Public-facing form display
   - Field type appropriate inputs
   - Validation (required, type checking)
   - Submit handling with loading state
   - Success/error states

3. Create new route `/frontend/src/routes/forms/[token]/+page.svelte`:
   - Public form submission page
   - No auth required
   - Clean, focused design

4. Add form management to base UI:
   - "Forms" tab or section
   - List of forms for each table
   - Link to form builder

---

### 2.2 Calendar View ✅

**Backend Implementation:**

No new migrations needed - extend ViewConfig:
```go
type ViewConfig struct {
    // ... existing fields
    CalendarConfig *CalendarConfig `json:"calendarConfig,omitempty"`
}

type CalendarConfig struct {
    DateFieldID    string `json:"dateFieldId"`    // Required: which field determines date
    EndDateFieldID string `json:"endDateFieldId"` // Optional: for multi-day events
    TitleFieldID   string `json:"titleFieldId"`   // Which field to display as title
}
```

**Frontend Implementation:**

1. Create `Calendar.svelte`:
   - Month view grid (7 columns, 5-6 rows)
   - Week view (7 columns, 24 hour rows)
   - Day indicators with record counts
   - Records shown as colored blocks
   - Click date to create record with that date
   - Click record to open detail modal
   - Drag record to change date

2. Calendar state management:
```typescript
interface CalendarState {
    currentDate: Date;
    viewMode: 'month' | 'week';
    selectedDate: Date | null;
}
```

3. Date positioning logic:
```typescript
function getRecordsForDate(records: Record[], date: Date, config: CalendarConfig): Record[] {
    return records.filter(r => {
        const recordDate = new Date(r.values[config.dateFieldId]);
        return isSameDay(recordDate, date);
    });
}
```

4. Integration:
   - Add "Calendar" to view type selector
   - Show only when table has date fields
   - Configuration panel to select date field

---

### 2.3 Gallery View ✅

**Backend Implementation:**

Extend ViewConfig:
```go
type GalleryConfig struct {
    CoverFieldID   string   `json:"coverFieldId"`   // Attachment or URL field
    TitleFieldID   string   `json:"titleFieldId"`
    VisibleFields  []string `json:"visibleFields"`  // Fields to show on card
    CardSize       string   `json:"cardSize"`       // 'small', 'medium', 'large'
}
```

**Frontend Implementation:**

1. Create `Gallery.svelte`:
   - CSS Grid layout (responsive columns)
   - Card component with:
     - Cover image (with fallback placeholder)
     - Title
     - Selected field values
   - Click card to open record detail
   - Hover effects

2. Card sizing:
```css
.gallery-grid {
    display: grid;
    gap: 1rem;
}
.gallery-grid.small { grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); }
.gallery-grid.medium { grid-template-columns: repeat(auto-fill, minmax(250px, 1fr)); }
.gallery-grid.large { grid-template-columns: repeat(auto-fill, minmax(350px, 1fr)); }
```

3. Image handling:
   - Lazy loading for performance
   - Aspect ratio container
   - Graceful fallback (icon or color block)

---

### 2.4 Public Shared Views ✅

**Backend Implementation:**

1. Create migration `010_create_view_shares.sql`:
```sql
ALTER TABLE views ADD COLUMN public_token VARCHAR(64) UNIQUE;
ALTER TABLE views ADD COLUMN is_public BOOLEAN DEFAULT false;
ALTER TABLE views ADD COLUMN public_password_hash VARCHAR(255);
```

2. Add to ViewStore:
```go
func (s *ViewStore) MakePublic(ctx context.Context, viewID uuid.UUID, userID uuid.UUID) (string, error)
func (s *ViewStore) MakePrivate(ctx context.Context, viewID uuid.UUID, userID uuid.UUID) error
func (s *ViewStore) SetPublicPassword(ctx context.Context, viewID uuid.UUID, userID uuid.UUID, password string) error
func (s *ViewStore) GetPublicView(ctx context.Context, token string, password string) (*models.View, error)
```

3. Create public endpoint in handlers:
```go
// GET /public/views/:token - Get view with records (no auth, optional password)
```

**API Endpoints:**
```
# Authenticated
POST   /views/:id/make-public     # Generate token, return share URL
POST   /views/:id/make-private    # Remove public access
PATCH  /views/:id/public-settings # Set/remove password

# Public
GET    /public/views/:token       # Get view config + records
POST   /public/views/:token/auth  # Verify password if protected
```

**Frontend Implementation:**

1. Update `ShareModal.svelte`:
   - "Share publicly" toggle
   - Generated link display with copy button
   - Optional password protection
   - Preview of what public users see

2. Create `/frontend/src/routes/public/views/[token]/+page.svelte`:
   - Password gate if required
   - Read-only view display
   - VibeTable branding footer
   - "Create your own" CTA

---

## Phase 3: Collaboration Features (Medium Effort) ✅ COMPLETE

### 3.1 Comments on Records ✅

**Backend Implementation:**

1. Create migration `011_create_comments.sql`:
```sql
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    record_id UUID NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE, -- For threading
    is_resolved BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_comments_record ON comments(record_id, created_at);
```

2. Create `/backend/internal/store/comment.go`:
```go
type CommentStore struct {
    db         DBTX
    baseStore  *BaseStore
    tableStore *TableStore
}

func (s *CommentStore) CreateComment(ctx context.Context, recordID uuid.UUID, userID uuid.UUID, content string, parentID *uuid.UUID) (*models.Comment, error)
func (s *CommentStore) ListComments(ctx context.Context, recordID uuid.UUID, userID uuid.UUID) ([]models.Comment, error)
func (s *CommentStore) UpdateComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID, content string) (*models.Comment, error)
func (s *CommentStore) DeleteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error
func (s *CommentStore) ResolveComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error
```

3. Create `/backend/internal/api/handlers/comment.go`

**API Endpoints:**
```
GET    /records/:recordId/comments      # List comments
POST   /records/:recordId/comments      # Create comment
PATCH  /comments/:id                    # Update comment
DELETE /comments/:id                    # Delete comment
POST   /comments/:id/resolve            # Toggle resolved
```

**Frontend Implementation:**

1. Create `CommentThread.svelte`:
   - List of comments with avatars and timestamps
   - Reply functionality
   - Edit/delete own comments
   - Resolve/unresolve (with visual indicator)
   - @mention autocomplete (future)

2. Create `RecordDetailModal.svelte`:
   - Full record view (all fields)
   - Comment thread
   - Activity history (if implemented)

3. Add comment indicator to Grid:
   - Small icon/badge on rows with comments
   - Comment count tooltip

---

### 3.2 Activity Log ✅

**Backend Implementation:**

1. Create migration `012_create_activity.sql`:
```sql
CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_id UUID NOT NULL REFERENCES bases(id) ON DELETE CASCADE,
    table_id UUID REFERENCES tables(id) ON DELETE SET NULL,
    record_id UUID REFERENCES records(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL, -- 'create', 'update', 'delete'
    entity_type VARCHAR(50) NOT NULL, -- 'record', 'field', 'table', 'view'
    entity_name VARCHAR(255), -- Snapshot of name at time of action
    changes JSONB, -- What changed (for updates)
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_activity_base ON activities(base_id, created_at DESC);
CREATE INDEX idx_activity_record ON activities(record_id, created_at DESC);
CREATE INDEX idx_activity_user ON activities(user_id, created_at DESC);
```

2. Create `/backend/internal/store/activity.go`:
```go
type ActivityStore struct {
    db DBTX
}

func (s *ActivityStore) Log(ctx context.Context, activity *models.Activity) error
func (s *ActivityStore) ListForBase(ctx context.Context, baseID uuid.UUID, filters ActivityFilters, limit, offset int) ([]models.Activity, error)
func (s *ActivityStore) ListForRecord(ctx context.Context, recordID uuid.UUID, limit int) ([]models.Activity, error)
```

3. Integrate activity logging into existing stores:
   - RecordStore.CreateRecord → Log create
   - RecordStore.UpdateRecord → Log update with changes
   - RecordStore.DeleteRecord → Log delete
   - Same for Field, Table, View stores

**API Endpoints:**
```
GET /bases/:baseId/activity              # Base-level activity
GET /bases/:baseId/activity?userId=x     # Filter by user
GET /bases/:baseId/activity?action=x     # Filter by action
GET /records/:recordId/activity          # Record-level activity
```

**Frontend Implementation:**

1. Create `ActivityFeed.svelte`:
   - Chronological list of activities
   - User avatars and names
   - Human-readable descriptions ("John updated 3 fields")
   - Expandable change details
   - Filters: user, action type, date range

2. Integration points:
   - Base-level activity panel (sidebar or tab)
   - Record detail modal (show recent activity)

---

## Phase 4: Advanced Field Types (High Effort, High Value) ✅ COMPLETE

### 4.1 Formula Fields ✅

**Backend Implementation:**

1. Extend field types in models:
```go
const (
    FieldTypeFormula FieldType = "formula"
)

type FormulaOptions struct {
    Expression   string `json:"expression"`   // The formula expression
    ResultType   string `json:"resultType"`   // 'text', 'number', 'date', 'boolean'
}
```

2. Create formula evaluation engine `/backend/internal/formula/`:
```go
package formula

type Evaluator struct{}

func (e *Evaluator) Parse(expression string) (*AST, error)
func (e *Evaluator) Validate(ast *AST, fields []models.Field) error
func (e *Evaluator) Evaluate(ast *AST, values map[string]interface{}) (interface{}, error)

// Supported functions
var functions = map[string]Function{
    // Text
    "CONCAT":     concatFunc,
    "LEFT":       leftFunc,
    "RIGHT":      rightFunc,
    "LEN":        lenFunc,
    "UPPER":      upperFunc,
    "LOWER":      lowerFunc,
    "TRIM":       trimFunc,

    // Numeric
    "SUM":        sumFunc,
    "AVERAGE":    averageFunc,
    "MIN":        minFunc,
    "MAX":        maxFunc,
    "ROUND":      roundFunc,
    "ABS":        absFunc,

    // Logic
    "IF":         ifFunc,
    "AND":        andFunc,
    "OR":         orFunc,
    "NOT":        notFunc,
    "SWITCH":     switchFunc,

    // Date
    "TODAY":      todayFunc,
    "NOW":        nowFunc,
    "DATEADD":    dateaddFunc,
    "DATEDIFF":   datediffFunc,
    "YEAR":       yearFunc,
    "MONTH":      monthFunc,
    "DAY":        dayFunc,
}
```

3. Modify RecordStore to compute formulas:
   - On record create/update, evaluate formula fields
   - Store computed value (for indexing/filtering)
   - Re-evaluate when dependencies change

4. Handle circular dependencies:
   - Build dependency graph
   - Topological sort for evaluation order
   - Error on cycles

**Frontend Implementation:**

1. Create `FormulaEditor.svelte`:
   - Text editor with syntax highlighting
   - Field reference autocomplete (type `{` to see fields)
   - Function autocomplete with descriptions
   - Live preview of result
   - Error display

2. Formula reference panel:
   - List of available functions
   - Usage examples
   - Field reference syntax

---

### 4.2 Rollup Fields ✅

**Backend Implementation:**

1. Extend field types:
```go
const (
    FieldTypeRollup FieldType = "rollup"
)

type RollupOptions struct {
    LinkedFieldID      string `json:"linkedFieldId"`      // The linked_record field
    RollupFieldID      string `json:"rollupFieldId"`      // Field in linked table to aggregate
    AggregationFunction string `json:"aggregationFunction"` // COUNT, SUM, AVG, MIN, MAX
}
```

2. Rollup computation:
```go
func (s *RecordStore) ComputeRollup(ctx context.Context, record *models.Record, field *models.Field) (interface{}, error) {
    opts := field.Options.(*models.RollupOptions)

    // Get linked record IDs
    linkedIDs := record.Values[opts.LinkedFieldID].([]string)

    // Fetch linked records
    linkedRecords, _ := s.GetRecordsByIDs(ctx, linkedIDs)

    // Extract values to aggregate
    values := make([]interface{}, len(linkedRecords))
    for i, r := range linkedRecords {
        values[i] = r.Values[opts.RollupFieldID]
    }

    // Apply aggregation
    return aggregate(opts.AggregationFunction, values)
}
```

3. Recompute rollups when linked records change

**Frontend Implementation:**

1. Update field configuration modal:
   - Select linked field (only show linked_record fields)
   - Select field from linked table to rollup
   - Select aggregation function
   - Preview result

---

### 4.3 Lookup Fields ✅

**Backend Implementation:**

1. Extend field types:
```go
const (
    FieldTypeLookup FieldType = "lookup"
)

type LookupOptions struct {
    LinkedFieldID  string `json:"linkedFieldId"`  // The linked_record field
    LookupFieldID  string `json:"lookupFieldId"`  // Field to pull from linked table
}
```

2. Lookup computation (similar to rollup but returns array of values)

**Frontend Implementation:**

1. Similar to rollup configuration
2. Display as comma-separated values or expandable list

---

### 4.4 File Attachments ✅

**Backend Implementation:**

1. Create migration `013_create_attachments.sql`:
```sql
CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    record_id UUID NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    field_id UUID NOT NULL REFERENCES fields(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    storage_key VARCHAR(500) NOT NULL, -- S3 key or file path
    thumbnail_key VARCHAR(500), -- For images
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_attachments_record_field ON attachments(record_id, field_id);
```

2. Extend field types:
```go
const (
    FieldTypeAttachment FieldType = "attachment"
)

type AttachmentOptions struct {
    AllowedTypes []string `json:"allowedTypes"` // ['image/*', 'application/pdf']
    MaxSizeBytes int64    `json:"maxSizeBytes"` // Default 10MB
}
```

3. Create storage interface `/backend/internal/storage/`:
```go
type Storage interface {
    Upload(ctx context.Context, key string, data io.Reader, contentType string) error
    Download(ctx context.Context, key string) (io.ReadCloser, error)
    GetSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
    Delete(ctx context.Context, key string) error
}

// Implementations
type S3Storage struct { ... }
type LocalStorage struct { ... } // For development
```

4. Create `/backend/internal/api/handlers/attachment.go`:
```go
// POST /records/:recordId/fields/:fieldId/attachments - Upload file
// GET  /attachments/:id - Get file (redirect to signed URL)
// DELETE /attachments/:id - Delete file
```

**API Endpoints:**
```
POST   /records/:recordId/fields/:fieldId/attachments  # Upload (multipart)
GET    /attachments/:id                                 # Download/redirect
DELETE /attachments/:id                                 # Delete
```

**Frontend Implementation:**

1. Create `FileUploader.svelte`:
   - Drag-and-drop zone
   - Click to browse
   - Upload progress indicator
   - Multiple file support
   - Type/size validation

2. Create `FilePreviewer.svelte`:
   - Image thumbnails
   - File type icons
   - Click to download or preview
   - Lightbox for images

3. Grid cell display:
   - Thumbnail previews
   - Count badge for multiple files
   - Quick actions (download, delete)

**Storage Configuration:**

```yaml
# docker-compose.yml addition
services:
  minio:
    image: minio/minio
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    command: server /data --console-address ":9001"
    volumes:
      - minio_data:/data
```

---

## Phase 5: Real-time & Automation (High Effort, High Impact) ⏳ IN PROGRESS

### 5.1 Real-time Collaboration ✅

**Backend Implementation:**

1. Add WebSocket support to main.go:
```go
// Using gorilla/websocket
import "github.com/gorilla/websocket"

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

// WebSocket hub
type Hub struct {
    bases      map[uuid.UUID]map[*Client]bool // baseID -> connected clients
    broadcast  chan Message
    register   chan *Client
    unregister chan *Client
}

type Client struct {
    hub    *Hub
    conn   *websocket.Conn
    send   chan Message
    baseID uuid.UUID
    userID uuid.UUID
}
```

2. Create `/backend/internal/realtime/`:
```go
package realtime

type Message struct {
    Type    string      `json:"type"` // 'record_update', 'presence', 'cursor'
    BaseID  uuid.UUID   `json:"baseId"`
    Payload interface{} `json:"payload"`
    UserID  uuid.UUID   `json:"userId"`
}

// Message types
const (
    MsgTypePresence      = "presence"       // User joined/left
    MsgTypeCursor        = "cursor"         // Cursor position
    MsgTypeRecordCreated = "record_created"
    MsgTypeRecordUpdated = "record_updated"
    MsgTypeRecordDeleted = "record_deleted"
    MsgTypeFieldUpdated  = "field_updated"
)
```

3. Modify stores to broadcast changes:
```go
func (s *RecordStore) UpdateRecord(...) (*models.Record, error) {
    record, err := s.updateRecordInternal(...)
    if err == nil && s.hub != nil {
        s.hub.broadcast <- realtime.Message{
            Type:    realtime.MsgTypeRecordUpdated,
            BaseID:  record.BaseID,
            Payload: record,
            UserID:  userID,
        }
    }
    return record, err
}
```

4. WebSocket endpoint:
```go
// GET /ws?baseId=xxx (with auth token in query or header)
func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request)
```

**Frontend Implementation:**

1. Create `realtimeStore.ts`:
```typescript
interface RealtimeState {
    connected: boolean;
    presence: Map<string, UserPresence>;
    cursors: Map<string, CursorPosition>;
}

function createRealtimeStore() {
    let ws: WebSocket | null = null;
    const { subscribe, update } = writable<RealtimeState>({...});

    return {
        subscribe,
        connect(baseId: string) { ... },
        disconnect() { ... },
        sendCursor(position: CursorPosition) { ... },
        onMessage(handler: (msg: Message) => void) { ... }
    };
}
```

2. Create `PresenceIndicator.svelte`:
   - Avatar stack of active users
   - Tooltip showing who's viewing
   - Color-coded cursors in grid

3. Update Grid.svelte:
   - Subscribe to realtime updates
   - Optimistically update UI
   - Show other users' cursors
   - Handle conflicts (last-write-wins with notification)

**Conflict Resolution Strategy:**
- Client sends update with `expectedVersion`
- Server checks version, returns conflict if mismatch
- Client prompts user: "This record was modified by X. Reload or overwrite?"

---

### 5.2 Automations ✅

**Backend Implementation:**

1. Create migration `014_create_automations.sql`:
```sql
CREATE TABLE automations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_id UUID NOT NULL REFERENCES bases(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE automation_triggers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    automation_id UUID NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
    trigger_type VARCHAR(50) NOT NULL, -- 'record_created', 'record_updated', 'field_matches'
    config JSONB NOT NULL
);

CREATE TABLE automation_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    automation_id UUID NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
    action_type VARCHAR(50) NOT NULL, -- 'send_email', 'create_record', 'update_record', 'webhook'
    config JSONB NOT NULL,
    position INT NOT NULL
);

CREATE TABLE automation_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    automation_id UUID NOT NULL REFERENCES automations(id) ON DELETE CASCADE,
    trigger_record_id UUID,
    status VARCHAR(20) NOT NULL, -- 'success', 'failed', 'running'
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    log JSONB
);

CREATE INDEX idx_automation_runs_automation ON automation_runs(automation_id, started_at DESC);
```

2. Create automation engine `/backend/internal/automation/`:
```go
package automation

type Engine struct {
    store     *store.AutomationStore
    executor  *ActionExecutor
}

// Trigger types
type Trigger interface {
    Match(event Event) bool
}

type RecordCreatedTrigger struct {
    TableID uuid.UUID
}

type FieldMatchesTrigger struct {
    TableID   uuid.UUID
    FieldID   string
    Operator  string
    Value     interface{}
}

// Action types
type Action interface {
    Execute(ctx context.Context, input ActionInput) error
}

type SendEmailAction struct {
    To      string // Can reference record fields with {{fieldName}}
    Subject string
    Body    string
}

type CreateRecordAction struct {
    TableID uuid.UUID
    Values  map[string]interface{} // Can reference trigger record
}

type WebhookAction struct {
    URL     string
    Method  string
    Headers map[string]string
    Body    string // Template with {{record.fieldName}}
}
```

3. Integration with existing stores:
```go
// In RecordStore.CreateRecord
func (s *RecordStore) CreateRecord(...) (*models.Record, error) {
    record, err := s.createRecordInternal(...)
    if err == nil && s.automationEngine != nil {
        go s.automationEngine.ProcessEvent(Event{
            Type:   EventRecordCreated,
            Record: record,
        })
    }
    return record, err
}
```

**API Endpoints:**
```
GET    /bases/:baseId/automations           # List automations
POST   /bases/:baseId/automations           # Create automation
GET    /automations/:id                     # Get automation details
PATCH  /automations/:id                     # Update automation
DELETE /automations/:id                     # Delete automation
POST   /automations/:id/toggle              # Enable/disable
GET    /automations/:id/runs                # Get run history
POST   /automations/:id/test                # Test run with sample data
```

**Frontend Implementation:**

1. Create `AutomationBuilder.svelte`:
   - Visual trigger configuration
   - Action chain builder
   - Field reference autocomplete
   - Test mode with sample data

2. Create `AutomationList.svelte`:
   - List of automations with status
   - Enable/disable toggles
   - Run history with error details

3. Create route `/bases/[baseId]/automations`:
   - Automation management page
   - Create/edit modal

---

### 5.3 Webhooks / API Keys ⏳

**Backend Implementation:**

1. Create migration `015_create_api_keys_webhooks.sql`:
```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL, -- Hashed API key
    key_prefix VARCHAR(10) NOT NULL, -- First 10 chars for identification
    scopes JSONB NOT NULL, -- ['read:bases', 'write:records', etc.]
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_id UUID NOT NULL REFERENCES bases(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    events JSONB NOT NULL, -- ['record.created', 'record.updated', etc.]
    secret VARCHAR(255), -- For signature verification
    is_active BOOLEAN DEFAULT true,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    response_status INT,
    response_body TEXT,
    delivered_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_webhooks_base ON webhooks(base_id);
CREATE INDEX idx_webhook_deliveries ON webhook_deliveries(webhook_id, delivered_at DESC);
```

2. API key authentication middleware:
```go
func (m *AuthMiddleware) APIKeyAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        if apiKey == "" {
            // Fall back to Bearer token
            m.Required(next).ServeHTTP(w, r)
            return
        }

        user, scopes, err := m.store.ValidateAPIKey(r.Context(), apiKey)
        if err != nil {
            writeError(w, http.StatusUnauthorized, "invalid_api_key", "Invalid API key")
            return
        }

        ctx := context.WithValue(r.Context(), "user", user)
        ctx = context.WithValue(ctx, "apiScopes", scopes)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

3. Webhook delivery:
```go
type WebhookDeliverer struct {
    store  *store.WebhookStore
    client *http.Client
}

func (d *WebhookDeliverer) Deliver(ctx context.Context, event Event) {
    webhooks, _ := d.store.GetWebhooksForEvent(ctx, event.BaseID, event.Type)

    for _, webhook := range webhooks {
        go d.deliverToWebhook(ctx, webhook, event)
    }
}

func (d *WebhookDeliverer) deliverToWebhook(ctx context.Context, webhook *models.Webhook, event Event) {
    payload := buildPayload(event)
    signature := computeSignature(payload, webhook.Secret)

    req, _ := http.NewRequest("POST", webhook.URL, bytes.NewReader(payload))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Vibetable-Signature", signature)

    resp, err := d.client.Do(req)
    d.store.LogDelivery(ctx, webhook.ID, event.Type, payload, resp)
}
```

**API Endpoints:**
```
# API Keys
GET    /api-keys                   # List user's API keys
POST   /api-keys                   # Create new key (returns full key once)
DELETE /api-keys/:id               # Revoke key

# Webhooks
GET    /bases/:baseId/webhooks     # List webhooks
POST   /bases/:baseId/webhooks     # Create webhook
PATCH  /webhooks/:id               # Update webhook
DELETE /webhooks/:id               # Delete webhook
GET    /webhooks/:id/deliveries    # Delivery history
POST   /webhooks/:id/test          # Send test event

# Incoming webhook (for creating records)
POST   /incoming/:baseId/:tableId  # Create record via webhook (API key auth)
```

**Frontend Implementation:**

1. Create `APIKeyManager.svelte`:
   - List of API keys (show prefix only)
   - Create new key modal (show full key once)
   - Scope selection
   - Revoke button

2. Create `WebhookConfig.svelte`:
   - Webhook list
   - Event type selection
   - URL and secret configuration
   - Delivery history with retry option

3. Create settings pages:
   - `/settings/api-keys`
   - `/bases/[baseId]/settings/webhooks`

---

## Implementation Sequence

### Recommended Order

Based on dependencies and incremental value:

```
Week 1-2:
├── CSV Import/Export (foundation for data movement)
├── Keyboard Shortcuts (improves daily UX)
└── Duplicate Base/Table (simple, useful)

Week 3-4:
├── Undo/Redo (requires action tracking infrastructure)
├── Record Color Coding (view enhancement)
└── Activity Log (uses similar infrastructure to undo)

Week 5-6:
├── Form View (new way to collect data)
├── Public Shared Views (builds on view system)
└── Comments on Records (collaboration foundation)

Week 7-8:
├── Calendar View (new visualization)
├── Gallery View (new visualization)
└── File Attachments (requires storage setup)

Week 9-10:
├── Formula Fields (complex but high value)
├── Lookup Fields (builds on linked records)
└── Rollup Fields (builds on lookup)

Week 11-12:
├── Real-time Collaboration (requires WebSocket infrastructure)
├── Automations (requires automation engine)
└── Webhooks / API Keys (enables integrations)
```

### Technical Dependencies Graph

```
                    ┌─────────────────┐
                    │ Activity Log    │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
        ┌──────────┐  ┌──────────┐  ┌──────────┐
        │ Undo/Redo│  │ Comments │  │Real-time │
        └──────────┘  └──────────┘  └────┬─────┘
                                         │
                                         ▼
                                  ┌──────────────┐
                                  │ Automations  │
                                  └──────────────┘
                                         │
                                         ▼
                                  ┌──────────────┐
                                  │  Webhooks    │
                                  └──────────────┘

        ┌──────────┐
        │ Linked   │
        │ Records  │
        └────┬─────┘
             │
     ┌───────┼───────┐
     │       │       │
     ▼       ▼       ▼
┌────────┐ ┌──────┐ ┌───────┐
│ Lookup │ │Rollup│ │Formula│
└────────┘ └──────┘ └───────┘

┌──────────────┐
│ Attachments  │──────► Storage Service (S3/MinIO)
└──────────────┘

┌──────────┐     ┌──────────────┐
│ Views    │────►│ Public Views │
└──────────┘     └──────────────┘
                        │
                        ▼
                 ┌──────────────┐
                 │  Form View   │
                 └──────────────┘
```

---

## Database Migration Summary

New migrations in order:

| # | Migration | Features |
|---|-----------|----------|
| 008 | create_actions | Undo/Redo |
| 009 | create_forms | Form View |
| 010 | add_view_sharing | Public Shared Views |
| 011 | create_comments | Comments on Records |
| 012 | create_activity | Activity Log |
| 013 | create_attachments | File Attachments |
| 014 | create_automations | Automations |
| 015 | create_api_keys_webhooks | Webhooks / API Keys |

---

## API Summary

New endpoints by feature:

### CSV Import/Export
```
POST /tables/:tableId/import/preview
POST /tables/:tableId/import
GET  /tables/:tableId/export
```

### Duplicate
```
POST /tables/:id/duplicate
POST /bases/:id/duplicate
```

### Undo/Redo
```
POST /bases/:baseId/undo
POST /bases/:baseId/redo
```

### Form View
```
GET    /tables/:tableId/forms
POST   /tables/:tableId/forms
GET    /forms/:id
PATCH  /forms/:id
DELETE /forms/:id
GET    /public/forms/:token
POST   /public/forms/:token
```

### Public Views
```
POST  /views/:id/make-public
POST  /views/:id/make-private
GET   /public/views/:token
```

### Comments
```
GET    /records/:recordId/comments
POST   /records/:recordId/comments
PATCH  /comments/:id
DELETE /comments/:id
```

### Activity Log
```
GET /bases/:baseId/activity
GET /records/:recordId/activity
```

### File Attachments
```
POST   /records/:recordId/fields/:fieldId/attachments
GET    /attachments/:id
DELETE /attachments/:id
```

### Real-time
```
GET /ws?baseId=xxx (WebSocket)
```

### Automations
```
GET    /bases/:baseId/automations
POST   /bases/:baseId/automations
GET    /automations/:id
PATCH  /automations/:id
DELETE /automations/:id
POST   /automations/:id/toggle
GET    /automations/:id/runs
```

### Webhooks / API Keys
```
GET    /api-keys
POST   /api-keys
DELETE /api-keys/:id
GET    /bases/:baseId/webhooks
POST   /bases/:baseId/webhooks
PATCH  /webhooks/:id
DELETE /webhooks/:id
POST   /incoming/:baseId/:tableId
```

---

## Testing Strategy

### Unit Tests (Go)
- Formula parser and evaluator
- Rollup/Lookup computation
- Webhook signature generation
- Automation trigger matching

### Integration Tests (Go)
- CSV import with various data types
- Undo/Redo across transactions
- Real-time message broadcasting
- Automation execution flow

### Frontend Tests (Svelte)
- Keyboard navigation
- Form builder interactions
- File upload flow
- Real-time state updates

### E2E Tests (Playwright)
- Full CSV import workflow
- Form submission by public user
- Automation creating records
- Collaborative editing scenario

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Formula injection attacks | Whitelist allowed functions, sandbox evaluation |
| Large CSV uploads blocking | Stream processing, background jobs |
| WebSocket connection storms | Connection limits per user, rate limiting |
| Automation infinite loops | Max execution depth, cooldown periods |
| Storage costs for attachments | Size limits, quota per base |
| Real-time conflicts | Last-write-wins with notification, no silent overwrites |

---

## Content-Worthy Milestones

📸 **Phase 1 Complete:** "Data portability unlocked - import your entire spreadsheet with one click"

📸 **Form View Launch:** "Turn any table into a public form - collect submissions from anyone"

📸 **Calendar View:** "Your data, visualized - see records on a calendar"

📸 **Real-time Working:** "True collaboration - see changes happen live as your team works"

📸 **Automations Launch:** "Set it and forget it - your data works while you sleep"

---

## Conclusion

This plan provides a comprehensive roadmap for expanding VibeTable from MVP to a feature-rich Airtable competitor. The phased approach ensures:

1. **Quick wins first** - Deliver value early with low-effort features
2. **Foundation building** - Activity log and undo infrastructure support later features
3. **User delight** - Visual features (calendar, gallery) provide demo appeal
4. **Power features last** - Complex features (real-time, automations) build on stable base

Each feature is designed to fit the existing architecture patterns, making implementation predictable and maintainable.
