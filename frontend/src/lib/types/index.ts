export interface User {
	id: string;
	email: string;
	name?: string;
	created_at: string;
	updated_at: string;
}

export type CollaboratorRole = 'owner' | 'editor' | 'viewer';

export interface Base {
	id: string;
	name: string;
	created_by: string;
	created_at: string;
	updated_at: string;
	role?: CollaboratorRole;
}

export interface BaseCollaborator {
	id: string;
	base_id: string;
	user_id: string;
	role: CollaboratorRole;
	created_at: string;
	user?: User;
}

export interface Table {
	id: string;
	base_id: string;
	name: string;
	position: number;
	created_at: string;
	updated_at: string;
}

export type FieldType = 'text' | 'number' | 'checkbox' | 'date' | 'single_select' | 'multi_select' | 'linked_record' | 'formula' | 'rollup' | 'lookup' | 'attachment';

export interface SelectOption {
	id: string;
	name: string;
	color?: string;
}

export interface FieldOptions {
	// Number field options
	precision?: number;
	format?: string;

	// Date field options
	include_time?: boolean;

	// Select field options
	options?: SelectOption[];

	// Linked record options
	linked_table_id?: string;

	// Formula field options
	expression?: string;
	result_type?: 'text' | 'number' | 'date' | 'boolean';

	// Rollup field options
	rollup_linked_field_id?: string;
	rollup_field_id?: string;
	aggregation_function?: 'COUNT' | 'COUNTA' | 'SUM' | 'AVG' | 'MIN' | 'MAX';

	// Lookup field options
	lookup_linked_field_id?: string;
	lookup_field_id?: string;

	// Attachment field options
	allowed_types?: string[];
	max_size_bytes?: number;
}

// Helper to check if a field type is computed (read-only)
export function isComputedField(fieldType: FieldType): boolean {
	return fieldType === 'formula' || fieldType === 'rollup' || fieldType === 'lookup';
}

export interface Field {
	id: string;
	table_id: string;
	name: string;
	field_type: FieldType;
	options: FieldOptions;
	position: number;
	created_at: string;
	updated_at: string;
}

export type RecordColor = 'red' | 'orange' | 'yellow' | 'green' | 'blue' | 'purple' | 'pink' | 'gray';

export interface Record {
	id: string;
	table_id: string;
	values: { [fieldId: string]: any };
	position: number;
	color?: RecordColor | null;
	created_at: string;
	updated_at: string;
}

export type ViewType = 'grid' | 'kanban' | 'calendar' | 'gallery';

export interface ViewFilter {
	field_id: string;
	operator: string;
	value: string;
}

export interface ViewSort {
	field_id: string;
	direction: 'asc' | 'desc';
}

export interface ViewConfig {
	filters?: ViewFilter[];
	sorts?: ViewSort[];
	group_by_field_id?: string;
	date_field_id?: string;
	title_field_id?: string;
	cover_field_id?: string;
	visible_fields?: string[];
}

export interface View {
	id: string;
	table_id: string;
	name: string;
	type: ViewType;
	config: ViewConfig;
	position: number;
	public_token?: string;
	is_public: boolean;
	created_at: string;
	updated_at: string;
}

export interface PublicView {
	view: View;
	table: Table;
	fields: Field[];
	records: Record[];
}

// Form types
export interface FormField {
	id: string;
	form_id: string;
	field_id: string;
	label?: string;
	help_text?: string;
	is_required: boolean;
	is_visible: boolean;
	position: number;
	field_name?: string;
	field_type?: FieldType;
	field_options?: FieldOptions;
}

export interface Form {
	id: string;
	table_id: string;
	name: string;
	description?: string;
	public_token?: string;
	is_active: boolean;
	success_message: string;
	redirect_url?: string;
	submit_button_text: string;
	created_by: string;
	created_at: string;
	updated_at: string;
	fields?: FormField[];
}

export interface PublicFormField {
	field_id: string;
	label: string;
	help_text?: string;
	is_required: boolean;
	field_type: FieldType;
	field_options?: FieldOptions;
	position: number;
}

export interface PublicForm {
	id: string;
	name: string;
	description?: string;
	success_message: string;
	redirect_url?: string;
	submit_button_text: string;
	fields: PublicFormField[];
}

// Comment types
export interface Comment {
	id: string;
	record_id: string;
	user_id: string;
	content: string;
	parent_id?: string;
	is_resolved: boolean;
	created_at: string;
	updated_at: string;
	user?: User;
	replies?: Comment[];
}

// Attachment types
export interface Attachment {
	id: string;
	record_id: string;
	field_id: string;
	filename: string;
	content_type: string;
	size_bytes: number;
	url: string;
	thumbnail_url?: string;
	width?: number;
	height?: number;
	created_by: string;
	created_at: string;
}

export interface AttachmentSummary {
	id: string;
	filename: string;
	content_type: string;
	size_bytes: number;
	url: string;
	thumbnail_url?: string;
	width?: number;
	height?: number;
}

// Automation types
export type TriggerType = 'record_created' | 'record_updated' | 'record_deleted' | 'field_value_changed' | 'scheduled';
export type ActionType = 'send_email' | 'update_record' | 'create_record' | 'send_webhook';
export type RunStatus = 'pending' | 'running' | 'success' | 'failed';

export interface Automation {
	id: string;
	base_id: string;
	table_id: string;
	name: string;
	description?: string;
	enabled: boolean;
	trigger_type: TriggerType;
	trigger_config: { [key: string]: any };
	action_type: ActionType;
	action_config: { [key: string]: any };
	created_by: string;
	last_triggered_at?: string;
	run_count: number;
	created_at: string;
	updated_at: string;
}

export interface AutomationRun {
	id: string;
	automation_id: string;
	status: RunStatus;
	trigger_record_id?: string;
	trigger_data?: { [key: string]: any };
	result?: { [key: string]: any };
	error?: string;
	started_at: string;
	completed_at?: string;
}

// API Key types
export interface APIKey {
	id: string;
	user_id: string;
	name: string;
	key_prefix: string;
	scopes: string[];
	last_used_at?: string;
	expires_at?: string;
	created_at: string;
}

export interface APIKeyWithToken extends APIKey {
	token: string;
}

// Webhook types
export type WebhookEvent = 'record.created' | 'record.updated' | 'record.deleted';

export interface Webhook {
	id: string;
	base_id: string;
	name: string;
	url: string;
	events: WebhookEvent[];
	secret?: string;
	is_active: boolean;
	created_by: string;
	created_at: string;
	updated_at: string;
}

export interface WebhookDelivery {
	id: string;
	webhook_id: string;
	event_type: string;
	payload: string;
	response_status?: number;
	response_body?: string;
	error?: string;
	duration_ms?: number;
	delivered_at: string;
}

// Activity types
export type ActivityAction = 'create' | 'update' | 'delete';
export type ActivityEntityType = 'record' | 'field' | 'table' | 'view' | 'base';

export interface ActivityChange {
	field_id?: string;
	field_name?: string;
	old_value?: any;
	new_value?: any;
}

export interface Activity {
	id: string;
	base_id: string;
	table_id?: string;
	record_id?: string;
	user_id: string;
	action: ActivityAction;
	entity_type: ActivityEntityType;
	entity_name?: string;
	changes?: ActivityChange[];
	created_at: string;
	user?: User;
}
