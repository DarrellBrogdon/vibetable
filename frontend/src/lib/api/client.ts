import type { User, Base, Table, Field, Record, RecordColor, BaseCollaborator, View, ViewConfig, ViewType, Form, FormField, PublicForm, PublicView, Comment, Activity, Attachment, Automation, AutomationRun, TriggerType, ActionType, APIKey, APIKeyWithToken, Webhook, WebhookDelivery, WebhookEvent } from '$lib/types';

const API_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';

class ApiError extends Error {
	constructor(public status: number, public code: string, message: string) {
		super(message);
		this.name = 'ApiError';
	}
}

async function request<T>(
	endpoint: string,
	options: RequestInit = {}
): Promise<T> {
	const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;

	const headers: HeadersInit = {
		'Content-Type': 'application/json',
		...options.headers,
	};

	if (token) {
		(headers as any)['Authorization'] = `Bearer ${token}`;
	}

	const response = await fetch(`${API_URL}/api/v1${endpoint}`, {
		...options,
		headers,
	});

	const data = await response.json();

	if (!response.ok) {
		throw new ApiError(response.status, data.error || 'unknown', data.message || 'Request failed');
	}

	return data;
}

// Auth API
export const auth = {
	login: (email: string, password: string) =>
		request<{ user: User; token: string }>('/auth/login', {
			method: 'POST',
			body: JSON.stringify({ email, password }),
		}),

	forgotPassword: (email: string) =>
		request<{ message: string }>('/auth/forgot-password', {
			method: 'POST',
			body: JSON.stringify({ email }),
		}),

	resetPassword: (token: string, password: string) =>
		request<{ message: string }>('/auth/reset-password', {
			method: 'POST',
			body: JSON.stringify({ token, password }),
		}),

	me: () => request<{ user: User }>('/auth/me'),

	updateMe: (name: string) =>
		request<{ user: User }>('/auth/me', {
			method: 'PATCH',
			body: JSON.stringify({ name }),
		}),

	logout: () =>
		request<{ message: string }>('/auth/logout', {
			method: 'POST',
		}),
};

// Bases API
export const bases = {
	list: () => request<{ bases: Base[] }>('/bases'),

	create: (name: string) =>
		request<Base>('/bases', {
			method: 'POST',
			body: JSON.stringify({ name }),
		}),

	get: (id: string) => request<Base>(`/bases/${id}`),

	update: (id: string, name: string) =>
		request<Base>(`/bases/${id}`, {
			method: 'PATCH',
			body: JSON.stringify({ name }),
		}),

	delete: (id: string) =>
		request<{ message: string }>(`/bases/${id}`, {
			method: 'DELETE',
		}),

	duplicate: (id: string, includeRecords: boolean = false) =>
		request<Base>(`/bases/${id}/duplicate`, {
			method: 'POST',
			body: JSON.stringify({ include_records: includeRecords }),
		}),

	// Collaborators
	listCollaborators: (baseId: string) =>
		request<{ collaborators: BaseCollaborator[] }>(`/bases/${baseId}/collaborators`),

	addCollaborator: (baseId: string, email: string, role: string) =>
		request<BaseCollaborator>(`/bases/${baseId}/collaborators`, {
			method: 'POST',
			body: JSON.stringify({ email, role }),
		}),

	updateCollaboratorRole: (baseId: string, userId: string, role: string) =>
		request<BaseCollaborator>(`/bases/${baseId}/collaborators/${userId}`, {
			method: 'PATCH',
			body: JSON.stringify({ role }),
		}),

	removeCollaborator: (baseId: string, userId: string) =>
		request<{ message: string }>(`/bases/${baseId}/collaborators/${userId}`, {
			method: 'DELETE',
		}),
};

// Tables API
export const tables = {
	list: (baseId: string) => request<{ tables: Table[] }>(`/bases/${baseId}/tables`),

	create: (baseId: string, name: string) =>
		request<Table>(`/bases/${baseId}/tables`, {
			method: 'POST',
			body: JSON.stringify({ name }),
		}),

	get: (id: string) => request<Table>(`/tables/${id}`),

	update: (id: string, name: string) =>
		request<Table>(`/tables/${id}`, {
			method: 'PATCH',
			body: JSON.stringify({ name }),
		}),

	delete: (id: string) =>
		request<{ message: string }>(`/tables/${id}`, {
			method: 'DELETE',
		}),

	duplicate: (id: string, includeRecords: boolean = false) =>
		request<Table>(`/tables/${id}/duplicate`, {
			method: 'POST',
			body: JSON.stringify({ include_records: includeRecords }),
		}),
};

// CSV API
export interface CSVPreviewResponse {
	columns: string[];
	rows: { [key: string]: string }[];
	total: number;
}

export interface CSVImportResponse {
	imported: number;
	skipped: number;
	errors: number;
}

export const csv = {
	preview: (tableId: string, data: string) =>
		request<CSVPreviewResponse>(`/tables/${tableId}/csv/preview`, {
			method: 'POST',
			body: JSON.stringify({ data }),
		}),

	import: (tableId: string, data: string, mappings: { [column: string]: string }) =>
		request<CSVImportResponse>(`/tables/${tableId}/csv/import`, {
			method: 'POST',
			body: JSON.stringify({ data, mappings }),
		}),

	exportUrl: (tableId: string) => {
		const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;
		const API_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';
		return `${API_URL}/api/v1/tables/${tableId}/csv/export?token=${token}`;
	},
};

// Fields API
export const fields = {
	list: (tableId: string) => request<{ fields: Field[] }>(`/tables/${tableId}/fields`),

	create: (tableId: string, name: string, field_type: string, options?: any) =>
		request<Field>(`/tables/${tableId}/fields`, {
			method: 'POST',
			body: JSON.stringify({ name, field_type, options }),
		}),

	get: (id: string) => request<Field>(`/fields/${id}`),

	update: (id: string, data: { name?: string; options?: any }) =>
		request<Field>(`/fields/${id}`, {
			method: 'PATCH',
			body: JSON.stringify(data),
		}),

	delete: (id: string) =>
		request<{ message: string }>(`/fields/${id}`, {
			method: 'DELETE',
		}),

	reorder: (tableId: string, fieldIds: string[]) =>
		request<{ message: string }>(`/tables/${tableId}/fields/reorder`, {
			method: 'PUT',
			body: JSON.stringify({ field_ids: fieldIds }),
		}),
};

// Records API
export const records = {
	list: (tableId: string) => request<{ records: Record[] }>(`/tables/${tableId}/records`),

	create: (tableId: string, values: { [fieldId: string]: any }) =>
		request<Record>(`/tables/${tableId}/records`, {
			method: 'POST',
			body: JSON.stringify({ values }),
		}),

	get: (id: string) => request<Record>(`/records/${id}`),

	update: (id: string, values: { [fieldId: string]: any }) =>
		request<Record>(`/records/${id}`, {
			method: 'PATCH',
			body: JSON.stringify({ values }),
		}),

	delete: (id: string) =>
		request<{ message: string }>(`/records/${id}`, {
			method: 'DELETE',
		}),

	updateColor: (id: string, color: RecordColor | null) =>
		request<Record>(`/records/${id}/color`, {
			method: 'PATCH',
			body: JSON.stringify({ color }),
		}),
};

// Views API
export const views = {
	list: (tableId: string) => request<{ views: View[] }>(`/tables/${tableId}/views`),

	create: (tableId: string, name: string, type: ViewType, config?: ViewConfig) =>
		request<View>(`/tables/${tableId}/views`, {
			method: 'POST',
			body: JSON.stringify({ name, type, config }),
		}),

	get: (id: string) => request<View>(`/views/${id}`),

	update: (id: string, data: { name?: string; config?: ViewConfig }) =>
		request<View>(`/views/${id}`, {
			method: 'PATCH',
			body: JSON.stringify(data),
		}),

	delete: (id: string) =>
		request<{ message: string }>(`/views/${id}`, {
			method: 'DELETE',
		}),

	setPublic: (id: string, isPublic: boolean) =>
		request<View>(`/views/${id}/public`, {
			method: 'PATCH',
			body: JSON.stringify({ is_public: isPublic }),
		}),

	getPublicUrl: (token: string) => {
		const baseUrl = typeof window !== 'undefined' ? window.location.origin : '';
		return `${baseUrl}/v/${token}`;
	},
};

// Forms API
export interface FormFieldUpdate {
	field_id: string;
	label?: string;
	help_text?: string;
	is_required: boolean;
	is_visible: boolean;
	position: number;
}

export const forms = {
	list: (tableId: string) => request<{ forms: Form[] }>(`/tables/${tableId}/forms`),

	create: (tableId: string, name: string) =>
		request<Form>(`/tables/${tableId}/forms`, {
			method: 'POST',
			body: JSON.stringify({ name }),
		}),

	get: (id: string) => request<Form>(`/forms/${id}`),

	update: (id: string, data: {
		name?: string;
		description?: string;
		is_active?: boolean;
		success_message?: string;
		redirect_url?: string;
		submit_button_text?: string;
	}) =>
		request<Form>(`/forms/${id}`, {
			method: 'PATCH',
			body: JSON.stringify(data),
		}),

	updateFields: (id: string, fields: FormFieldUpdate[]) =>
		request<{ message: string }>(`/forms/${id}/fields`, {
			method: 'PATCH',
			body: JSON.stringify({ fields }),
		}),

	delete: (id: string) =>
		request<{ message: string }>(`/forms/${id}`, {
			method: 'DELETE',
		}),

	// Public form URLs
	getPublicUrl: (token: string) => {
		const baseUrl = typeof window !== 'undefined' ? window.location.origin : '';
		return `${baseUrl}/f/${token}`;
	},
};

// Public Forms API (no auth required)
async function publicRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
	const headers: HeadersInit = {
		'Content-Type': 'application/json',
		...options.headers,
	};

	const response = await fetch(`${API_URL}/api/v1${endpoint}`, {
		...options,
		headers,
	});

	const data = await response.json();

	if (!response.ok) {
		throw new ApiError(response.status, data.error || 'unknown', data.message || 'Request failed');
	}

	return data;
}

export const publicForms = {
	get: (token: string) => publicRequest<PublicForm>(`/public/forms/${token}`),

	submit: (token: string, values: { [fieldId: string]: any }) =>
		publicRequest<{ message: string; record_id: string }>(`/public/forms/${token}`, {
			method: 'POST',
			body: JSON.stringify({ values }),
		}),
};

// Public Views API (no auth required)
export const publicViews = {
	get: (token: string) => publicRequest<PublicView>(`/public/views/${token}`),
};

// Comments API
export const comments = {
	list: (recordId: string) =>
		request<{ comments: Comment[] }>(`/records/${recordId}/comments`),

	create: (recordId: string, content: string, parentId?: string) =>
		request<Comment>(`/records/${recordId}/comments`, {
			method: 'POST',
			body: JSON.stringify({ content, parent_id: parentId }),
		}),

	get: (id: string) => request<Comment>(`/comments/${id}`),

	update: (id: string, content: string) =>
		request<Comment>(`/comments/${id}`, {
			method: 'PATCH',
			body: JSON.stringify({ content }),
		}),

	delete: (id: string) =>
		request<{ message: string }>(`/comments/${id}`, {
			method: 'DELETE',
		}),

	resolve: (id: string, resolved: boolean) =>
		request<Comment>(`/comments/${id}/resolve`, {
			method: 'POST',
			body: JSON.stringify({ resolved }),
		}),
};

// Activity API
export interface ActivityFilters {
	userId?: string;
	action?: string;
	entityType?: string;
	tableId?: string;
	limit?: number;
	offset?: number;
}

export const activity = {
	listForBase: (baseId: string, filters?: ActivityFilters) => {
		const params = new URLSearchParams();
		if (filters?.userId) params.set('userId', filters.userId);
		if (filters?.action) params.set('action', filters.action);
		if (filters?.entityType) params.set('entityType', filters.entityType);
		if (filters?.tableId) params.set('tableId', filters.tableId);
		if (filters?.limit) params.set('limit', filters.limit.toString());
		if (filters?.offset) params.set('offset', filters.offset.toString());
		const queryString = params.toString();
		return request<{ activities: Activity[] }>(`/bases/${baseId}/activity${queryString ? `?${queryString}` : ''}`);
	},

	listForRecord: (recordId: string, limit?: number) =>
		request<{ activities: Activity[] }>(`/records/${recordId}/activity${limit ? `?limit=${limit}` : ''}`),
};

// Attachments API
export const attachments = {
	list: (recordId: string, fieldId: string) =>
		request<{ attachments: Attachment[] }>(`/records/${recordId}/fields/${fieldId}/attachments`),

	upload: async (recordId: string, fieldId: string, file: File): Promise<Attachment> => {
		const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;

		const formData = new FormData();
		formData.append('file', file);

		const headers: HeadersInit = {};
		if (token) {
			headers['Authorization'] = `Bearer ${token}`;
		}

		const response = await fetch(`${API_URL}/api/v1/records/${recordId}/fields/${fieldId}/attachments`, {
			method: 'POST',
			headers,
			body: formData,
		});

		const data = await response.json();

		if (!response.ok) {
			throw new ApiError(response.status, data.error || 'unknown', data.message || 'Upload failed');
		}

		return data;
	},

	get: (id: string) => request<Attachment>(`/attachments/${id}`),

	delete: (id: string) =>
		request<{ message: string }>(`/attachments/${id}`, {
			method: 'DELETE',
		}),

	getDownloadUrl: (id: string) => {
		const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;
		return `${API_URL}/api/v1/attachments/${id}/download?token=${token}`;
	},
};

// Automations API
export const automations = {
	list: (tableId: string) =>
		request<{ automations: Automation[] }>(`/tables/${tableId}/automations`),

	create: (tableId: string, data: {
		name: string;
		description?: string;
		enabled?: boolean;
		triggerType: TriggerType;
		triggerConfig?: { [key: string]: any };
		actionType: ActionType;
		actionConfig?: { [key: string]: any };
	}) =>
		request<Automation>(`/tables/${tableId}/automations`, {
			method: 'POST',
			body: JSON.stringify(data),
		}),

	get: (id: string) => request<Automation>(`/automations/${id}`),

	update: (id: string, data: {
		name?: string;
		description?: string;
		enabled?: boolean;
		triggerType?: TriggerType;
		triggerConfig?: { [key: string]: any };
		actionType?: ActionType;
		actionConfig?: { [key: string]: any };
	}) =>
		request<Automation>(`/automations/${id}`, {
			method: 'PATCH',
			body: JSON.stringify(data),
		}),

	delete: (id: string) =>
		request<{ message: string }>(`/automations/${id}`, {
			method: 'DELETE',
		}),

	toggle: (id: string, enabled: boolean) =>
		request<Automation>(`/automations/${id}/toggle`, {
			method: 'POST',
			body: JSON.stringify({ enabled }),
		}),

	listRuns: (id: string) =>
		request<{ runs: AutomationRun[] }>(`/automations/${id}/runs`),
};

// API Keys API
export const apiKeys = {
	list: () => request<{ api_keys: APIKey[] }>('/api-keys'),

	create: (name: string, scopes?: string[], expiresAt?: string) =>
		request<APIKeyWithToken>('/api-keys', {
			method: 'POST',
			body: JSON.stringify({ name, scopes, expires_at: expiresAt }),
		}),

	get: (id: string) => request<APIKey>(`/api-keys/${id}`),

	delete: (id: string) =>
		request<{ message: string }>(`/api-keys/${id}`, {
			method: 'DELETE',
		}),
};

// Webhooks API
export const webhooks = {
	list: (baseId: string) => request<{ webhooks: Webhook[] }>(`/bases/${baseId}/webhooks`),

	create: (baseId: string, name: string, url: string, events?: WebhookEvent[], secret?: string) =>
		request<Webhook>(`/bases/${baseId}/webhooks`, {
			method: 'POST',
			body: JSON.stringify({ name, url, events, secret }),
		}),

	get: (id: string) => request<Webhook>(`/webhooks/${id}`),

	update: (id: string, data: {
		name?: string;
		url?: string;
		events?: WebhookEvent[];
		secret?: string;
		is_active?: boolean;
	}) =>
		request<Webhook>(`/webhooks/${id}`, {
			method: 'PATCH',
			body: JSON.stringify(data),
		}),

	delete: (id: string) =>
		request<{ message: string }>(`/webhooks/${id}`, {
			method: 'DELETE',
		}),

	listDeliveries: (id: string) =>
		request<{ deliveries: WebhookDelivery[] }>(`/webhooks/${id}/deliveries`),
};

export { ApiError };
