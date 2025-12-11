import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { auth, bases, tables, fields, records, views, csv, forms, publicForms, publicViews, comments, activity, attachments, automations, apiKeys, webhooks, ApiError } from './client';

// Mock fetch globally
const mockFetch = vi.fn();
global.fetch = mockFetch;

// Mock localStorage
const localStorageMock = (() => {
	let store: Record<string, string> = {};
	return {
		getItem: vi.fn((key: string) => store[key] || null),
		setItem: vi.fn((key: string, value: string) => { store[key] = value; }),
		removeItem: vi.fn((key: string) => { delete store[key]; }),
		clear: vi.fn(() => { store = {}; }),
	};
})();
Object.defineProperty(global, 'localStorage', { value: localStorageMock });

// Mock window for URL generation
Object.defineProperty(global, 'window', {
	value: { location: { origin: 'http://localhost:5173' } },
	writable: true,
});

// Mock import.meta.env
vi.stubGlobal('import', { meta: { env: { PUBLIC_API_URL: 'http://localhost:8080' } } });

describe('API Client', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		localStorageMock.clear();
	});

	describe('request helper', () => {
		it('should add Authorization header when token exists', async () => {
			localStorageMock.setItem('token', 'test-token');
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ user: { id: '1', email: 'test@example.com' } }),
			});

			await auth.me();

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/auth/me',
				expect.objectContaining({
					headers: expect.objectContaining({
						'Authorization': 'Bearer test-token',
					}),
				})
			);
		});

		it('should not add Authorization header when no token', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ user: { id: '1', email: 'test@example.com' } }),
			});

			await auth.me();

			const callArgs = mockFetch.mock.calls[0];
			expect(callArgs[1].headers['Authorization']).toBeUndefined();
		});

		it('should throw ApiError on non-ok response', async () => {
			localStorageMock.setItem('token', 'test-token');
			mockFetch.mockResolvedValueOnce({
				ok: false,
				status: 401,
				json: () => Promise.resolve({ error: 'unauthorized', message: 'Invalid token' }),
			});

			await expect(auth.me()).rejects.toThrow(ApiError);
		});

		it('should include error details in ApiError', async () => {
			localStorageMock.setItem('token', 'test-token');
			mockFetch.mockResolvedValueOnce({
				ok: false,
				status: 401,
				json: () => Promise.resolve({ error: 'unauthorized', message: 'Invalid token' }),
			});

			try {
				await auth.me();
				expect.fail('Should have thrown');
			} catch (e) {
				expect(e).toBeInstanceOf(ApiError);
				expect((e as ApiError).status).toBe(401);
				expect((e as ApiError).code).toBe('unauthorized');
				expect((e as ApiError).message).toBe('Invalid token');
			}
		});
	});

	describe('auth', () => {
		it('should login with email and password', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ user: { id: '1', email: 'test@example.com' }, token: 'session-token' }),
			});

			const result = await auth.login('test@example.com', 'password123');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/auth/login',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ email: 'test@example.com', password: 'password123' }),
				})
			);
			expect(result.user.email).toBe('test@example.com');
			expect(result.token).toBe('session-token');
		});

		it('should request forgot password', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Password reset email sent' }),
			});

			const result = await auth.forgotPassword('test@example.com');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/auth/forgot-password',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ email: 'test@example.com' }),
				})
			);
			expect(result.message).toBe('Password reset email sent');
		});

		it('should reset password', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Password reset successful' }),
			});

			const result = await auth.resetPassword('reset-token', 'newpassword123');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/auth/reset-password',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ token: 'reset-token', password: 'newpassword123' }),
				})
			);
			expect(result.message).toBe('Password reset successful');
		});

		it('should get current user', async () => {
			localStorageMock.setItem('token', 'test-token');
			const mockUser = { id: '1', email: 'test@example.com' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ user: mockUser }),
			});

			const result = await auth.me();

			expect(result.user).toEqual(mockUser);
		});

		it('should update user name', async () => {
			localStorageMock.setItem('token', 'test-token');
			const mockUser = { id: '1', email: 'test@example.com', name: 'New Name' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ user: mockUser }),
			});

			const result = await auth.updateMe('New Name');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/auth/me',
				expect.objectContaining({
					method: 'PATCH',
					body: JSON.stringify({ name: 'New Name' }),
				})
			);
			expect(result.user.name).toBe('New Name');
		});

		it('should logout', async () => {
			localStorageMock.setItem('token', 'test-token');
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Logged out' }),
			});

			const result = await auth.logout();

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/auth/logout',
				expect.objectContaining({
					method: 'POST',
				})
			);
			expect(result.message).toBe('Logged out');
		});
	});

	describe('bases', () => {
		beforeEach(() => {
			localStorageMock.setItem('token', 'test-token');
		});

		it('should list bases', async () => {
			const mockBases = [{ id: '1', name: 'Base 1' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ bases: mockBases }),
			});

			const result = await bases.list();

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/bases',
				expect.anything()
			);
			expect(result.bases).toEqual(mockBases);
		});

		it('should create base', async () => {
			const mockBase = { id: '1', name: 'New Base' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockBase),
			});

			const result = await bases.create('New Base');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/bases',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ name: 'New Base' }),
				})
			);
			expect(result).toEqual(mockBase);
		});

		it('should get base', async () => {
			const mockBase = { id: '1', name: 'Base 1' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockBase),
			});

			const result = await bases.get('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/bases/1',
				expect.anything()
			);
			expect(result).toEqual(mockBase);
		});

		it('should update base', async () => {
			const mockBase = { id: '1', name: 'Updated Name' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockBase),
			});

			const result = await bases.update('1', 'Updated Name');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/bases/1',
				expect.objectContaining({
					method: 'PATCH',
					body: JSON.stringify({ name: 'Updated Name' }),
				})
			);
			expect(result).toEqual(mockBase);
		});

		it('should delete base', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Base deleted' }),
			});

			const result = await bases.delete('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/bases/1',
				expect.objectContaining({
					method: 'DELETE',
				})
			);
			expect(result.message).toBe('Base deleted');
		});

		it('should duplicate base without records', async () => {
			const mockBase = { id: '2', name: 'Base 1 (copy)' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockBase),
			});

			const result = await bases.duplicate('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/bases/1/duplicate',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ include_records: false }),
				})
			);
			expect(result).toEqual(mockBase);
		});

		it('should duplicate base with records', async () => {
			const mockBase = { id: '2', name: 'Base 1 (copy)' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockBase),
			});

			const result = await bases.duplicate('1', true);

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/bases/1/duplicate',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ include_records: true }),
				})
			);
		});

		describe('collaborators', () => {
			it('should list collaborators', async () => {
				const mockCollabs = [{ id: '1', user_id: '2', role: 'editor' }];
				mockFetch.mockResolvedValueOnce({
					ok: true,
					json: () => Promise.resolve({ collaborators: mockCollabs }),
				});

				const result = await bases.listCollaborators('base-1');

				expect(mockFetch).toHaveBeenCalledWith(
					'http://localhost:8080/api/v1/bases/base-1/collaborators',
					expect.anything()
				);
				expect(result.collaborators).toEqual(mockCollabs);
			});

			it('should add collaborator', async () => {
				const mockCollab = { id: '1', user_id: '2', role: 'editor' };
				mockFetch.mockResolvedValueOnce({
					ok: true,
					json: () => Promise.resolve(mockCollab),
				});

				const result = await bases.addCollaborator('base-1', 'user@example.com', 'editor');

				expect(mockFetch).toHaveBeenCalledWith(
					'http://localhost:8080/api/v1/bases/base-1/collaborators',
					expect.objectContaining({
						method: 'POST',
						body: JSON.stringify({ email: 'user@example.com', role: 'editor' }),
					})
				);
				expect(result).toEqual(mockCollab);
			});

			it('should update collaborator role', async () => {
				const mockCollab = { id: '1', user_id: '2', role: 'viewer' };
				mockFetch.mockResolvedValueOnce({
					ok: true,
					json: () => Promise.resolve(mockCollab),
				});

				const result = await bases.updateCollaboratorRole('base-1', 'user-2', 'viewer');

				expect(mockFetch).toHaveBeenCalledWith(
					'http://localhost:8080/api/v1/bases/base-1/collaborators/user-2',
					expect.objectContaining({
						method: 'PATCH',
						body: JSON.stringify({ role: 'viewer' }),
					})
				);
				expect(result).toEqual(mockCollab);
			});

			it('should remove collaborator', async () => {
				mockFetch.mockResolvedValueOnce({
					ok: true,
					json: () => Promise.resolve({ message: 'Collaborator removed' }),
				});

				const result = await bases.removeCollaborator('base-1', 'user-2');

				expect(mockFetch).toHaveBeenCalledWith(
					'http://localhost:8080/api/v1/bases/base-1/collaborators/user-2',
					expect.objectContaining({
						method: 'DELETE',
					})
				);
				expect(result.message).toBe('Collaborator removed');
			});
		});
	});

	describe('tables', () => {
		beforeEach(() => {
			localStorageMock.setItem('token', 'test-token');
		});

		it('should list tables', async () => {
			const mockTables = [{ id: '1', name: 'Table 1' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ tables: mockTables }),
			});

			const result = await tables.list('base-1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/bases/base-1/tables',
				expect.anything()
			);
			expect(result.tables).toEqual(mockTables);
		});

		it('should create table', async () => {
			const mockTable = { id: '1', name: 'New Table' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockTable),
			});

			const result = await tables.create('base-1', 'New Table');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/bases/base-1/tables',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ name: 'New Table' }),
				})
			);
			expect(result).toEqual(mockTable);
		});

		it('should get table', async () => {
			const mockTable = { id: '1', name: 'Table 1' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockTable),
			});

			const result = await tables.get('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/1',
				expect.anything()
			);
			expect(result).toEqual(mockTable);
		});

		it('should update table', async () => {
			const mockTable = { id: '1', name: 'Updated Table' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockTable),
			});

			const result = await tables.update('1', 'Updated Table');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/1',
				expect.objectContaining({
					method: 'PATCH',
					body: JSON.stringify({ name: 'Updated Table' }),
				})
			);
			expect(result).toEqual(mockTable);
		});

		it('should delete table', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Table deleted' }),
			});

			const result = await tables.delete('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/1',
				expect.objectContaining({
					method: 'DELETE',
				})
			);
			expect(result.message).toBe('Table deleted');
		});

		it('should duplicate table', async () => {
			const mockTable = { id: '2', name: 'Table 1 (copy)' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockTable),
			});

			const result = await tables.duplicate('1', true);

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/1/duplicate',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ include_records: true }),
				})
			);
			expect(result).toEqual(mockTable);
		});
	});

	describe('csv', () => {
		beforeEach(() => {
			localStorageMock.setItem('token', 'test-token');
		});

		it('should preview CSV', async () => {
			const mockPreview = { columns: ['name', 'email'], rows: [{ name: 'John', email: 'john@example.com' }], total: 1 };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockPreview),
			});

			const result = await csv.preview('table-1', 'name,email\nJohn,john@example.com');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/csv/preview',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ data: 'name,email\nJohn,john@example.com' }),
				})
			);
			expect(result).toEqual(mockPreview);
		});

		it('should import CSV', async () => {
			const mockResult = { imported: 10, skipped: 2, errors: 0 };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockResult),
			});

			const result = await csv.import('table-1', 'csv-data', { name: 'field-1' });

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/csv/import',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ data: 'csv-data', mappings: { name: 'field-1' } }),
				})
			);
			expect(result).toEqual(mockResult);
		});

		it('should generate export URL', () => {
			localStorageMock.setItem('token', 'test-token');
			const url = csv.exportUrl('table-1');
			expect(url).toBe('http://localhost:8080/api/v1/tables/table-1/csv/export?token=test-token');
		});
	});

	describe('fields', () => {
		beforeEach(() => {
			localStorageMock.setItem('token', 'test-token');
		});

		it('should list fields', async () => {
			const mockFields = [{ id: '1', name: 'Field 1', field_type: 'text' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ fields: mockFields }),
			});

			const result = await fields.list('table-1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/fields',
				expect.anything()
			);
			expect(result.fields).toEqual(mockFields);
		});

		it('should create field with options', async () => {
			const mockField = { id: '1', name: 'Status', field_type: 'single_select', options: { options: [] } };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockField),
			});

			const options = { options: [{ id: '1', name: 'Active', color: 'green' }] };
			const result = await fields.create('table-1', 'Status', 'single_select', options);

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/fields',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ name: 'Status', field_type: 'single_select', options }),
				})
			);
			expect(result).toEqual(mockField);
		});

		it('should create field without options', async () => {
			const mockField = { id: '1', name: 'Title', field_type: 'text' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockField),
			});

			const result = await fields.create('table-1', 'Title', 'text');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/fields',
				expect.objectContaining({
					method: 'POST',
				})
			);
			expect(result).toEqual(mockField);
		});

		it('should get field', async () => {
			const mockField = { id: '1', name: 'Field 1' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockField),
			});

			const result = await fields.get('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/fields/1',
				expect.anything()
			);
			expect(result).toEqual(mockField);
		});

		it('should update field', async () => {
			const mockField = { id: '1', name: 'Updated Field', options: {} };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockField),
			});

			const result = await fields.update('1', { name: 'Updated Field' });

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/fields/1',
				expect.objectContaining({
					method: 'PATCH',
					body: JSON.stringify({ name: 'Updated Field' }),
				})
			);
			expect(result).toEqual(mockField);
		});

		it('should delete field', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Field deleted' }),
			});

			const result = await fields.delete('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/fields/1',
				expect.objectContaining({
					method: 'DELETE',
				})
			);
			expect(result.message).toBe('Field deleted');
		});

		it('should reorder fields', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Fields reordered' }),
			});

			const result = await fields.reorder('table-1', ['field-3', 'field-1', 'field-2']);

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/fields/reorder',
				expect.objectContaining({
					method: 'PUT',
					body: JSON.stringify({ field_ids: ['field-3', 'field-1', 'field-2'] }),
				})
			);
			expect(result.message).toBe('Fields reordered');
		});
	});

	describe('records', () => {
		beforeEach(() => {
			localStorageMock.setItem('token', 'test-token');
		});

		it('should list records', async () => {
			const mockRecords = [{ id: '1', values: { field1: 'value1' } }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ records: mockRecords }),
			});

			const result = await records.list('table-1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/records',
				expect.anything()
			);
			expect(result.records).toEqual(mockRecords);
		});

		it('should create record', async () => {
			const mockRecord = { id: '1', values: { field1: 'value1' } };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockRecord),
			});

			const result = await records.create('table-1', { field1: 'value1' });

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/records',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ values: { field1: 'value1' } }),
				})
			);
			expect(result).toEqual(mockRecord);
		});

		it('should get record', async () => {
			const mockRecord = { id: '1', values: {} };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockRecord),
			});

			const result = await records.get('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/records/1',
				expect.anything()
			);
			expect(result).toEqual(mockRecord);
		});

		it('should update record', async () => {
			const mockRecord = { id: '1', values: { field1: 'updated' } };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockRecord),
			});

			const result = await records.update('1', { field1: 'updated' });

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/records/1',
				expect.objectContaining({
					method: 'PATCH',
					body: JSON.stringify({ values: { field1: 'updated' } }),
				})
			);
			expect(result).toEqual(mockRecord);
		});

		it('should delete record', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Record deleted' }),
			});

			const result = await records.delete('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/records/1',
				expect.objectContaining({
					method: 'DELETE',
				})
			);
			expect(result.message).toBe('Record deleted');
		});

		it('should update record color', async () => {
			const mockRecord = { id: '1', values: {}, color: 'red' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockRecord),
			});

			const result = await records.updateColor('1', 'red');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/records/1/color',
				expect.objectContaining({
					method: 'PATCH',
					body: JSON.stringify({ color: 'red' }),
				})
			);
			expect(result).toEqual(mockRecord);
		});
	});

	describe('views', () => {
		beforeEach(() => {
			localStorageMock.setItem('token', 'test-token');
		});

		it('should list views', async () => {
			const mockViews = [{ id: '1', name: 'Grid View', type: 'grid' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ views: mockViews }),
			});

			const result = await views.list('table-1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/views',
				expect.anything()
			);
			expect(result.views).toEqual(mockViews);
		});

		it('should create grid view', async () => {
			const mockView = { id: '1', name: 'Grid View', type: 'grid' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockView),
			});

			const result = await views.create('table-1', 'Grid View', 'grid');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/views',
				expect.objectContaining({
					method: 'POST',
				})
			);
			expect(result).toEqual(mockView);
		});

		it('should create kanban view with config', async () => {
			const config = { group_by_field_id: 'field-1' };
			const mockView = { id: '1', name: 'Kanban', type: 'kanban', config };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockView),
			});

			const result = await views.create('table-1', 'Kanban', 'kanban', config);

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/views',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ name: 'Kanban', type: 'kanban', config }),
				})
			);
			expect(result).toEqual(mockView);
		});

		it('should get view', async () => {
			const mockView = { id: '1', name: 'View 1' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockView),
			});

			const result = await views.get('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/views/1',
				expect.anything()
			);
			expect(result).toEqual(mockView);
		});

		it('should update view', async () => {
			const mockView = { id: '1', name: 'Updated View', config: { filters: [] } };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockView),
			});

			const result = await views.update('1', { name: 'Updated View', config: { filters: [] } });

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/views/1',
				expect.objectContaining({
					method: 'PATCH',
					body: JSON.stringify({ name: 'Updated View', config: { filters: [] } }),
				})
			);
			expect(result).toEqual(mockView);
		});

		it('should delete view', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'View deleted' }),
			});

			const result = await views.delete('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/views/1',
				expect.objectContaining({
					method: 'DELETE',
				})
			);
			expect(result.message).toBe('View deleted');
		});

		it('should set view public', async () => {
			const mockView = { id: '1', name: 'View 1', is_public: true };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockView),
			});

			const result = await views.setPublic('1', true);

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/views/1/public',
				expect.objectContaining({
					method: 'PATCH',
					body: JSON.stringify({ is_public: true }),
				})
			);
			expect(result).toEqual(mockView);
		});

		it('should generate public URL', () => {
			const url = views.getPublicUrl('abc123');
			expect(url).toBe('http://localhost:5173/v/abc123');
		});
	});

	describe('forms', () => {
		beforeEach(() => {
			localStorageMock.setItem('token', 'test-token');
		});

		it('should list forms', async () => {
			const mockForms = [{ id: '1', name: 'Contact Form' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ forms: mockForms }),
			});

			const result = await forms.list('table-1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/forms',
				expect.anything()
			);
			expect(result.forms).toEqual(mockForms);
		});

		it('should create form', async () => {
			const mockForm = { id: '1', name: 'New Form' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockForm),
			});

			const result = await forms.create('table-1', 'New Form');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/forms',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ name: 'New Form' }),
				})
			);
			expect(result).toEqual(mockForm);
		});

		it('should get form', async () => {
			const mockForm = { id: '1', name: 'Form 1' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockForm),
			});

			const result = await forms.get('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/forms/1',
				expect.anything()
			);
			expect(result).toEqual(mockForm);
		});

		it('should update form', async () => {
			const mockForm = { id: '1', name: 'Updated Form', is_active: true };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockForm),
			});

			const result = await forms.update('1', { name: 'Updated Form', is_active: true });

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/forms/1',
				expect.objectContaining({
					method: 'PATCH',
				})
			);
			expect(result).toEqual(mockForm);
		});

		it('should update form fields', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Fields updated' }),
			});

			const formFields = [{ field_id: 'f1', is_required: true, is_visible: true, position: 0 }];
			const result = await forms.updateFields('1', formFields);

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/forms/1/fields',
				expect.objectContaining({
					method: 'PATCH',
					body: JSON.stringify({ fields: formFields }),
				})
			);
			expect(result.message).toBe('Fields updated');
		});

		it('should delete form', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Form deleted' }),
			});

			const result = await forms.delete('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/forms/1',
				expect.objectContaining({
					method: 'DELETE',
				})
			);
			expect(result.message).toBe('Form deleted');
		});

		it('should generate public form URL', () => {
			const url = forms.getPublicUrl('token123');
			expect(url).toBe('http://localhost:5173/f/token123');
		});
	});

	describe('publicForms', () => {
		it('should get public form', async () => {
			const mockForm = { id: '1', name: 'Public Form', fields: [] };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockForm),
			});

			const result = await publicForms.get('token123');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/public/forms/token123',
				expect.anything()
			);
			expect(result).toEqual(mockForm);
		});

		it('should submit public form', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Submitted', record_id: 'rec-1' }),
			});

			const result = await publicForms.submit('token123', { field1: 'value1' });

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/public/forms/token123',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ values: { field1: 'value1' } }),
				})
			);
			expect(result.record_id).toBe('rec-1');
		});
	});

	describe('publicViews', () => {
		it('should get public view', async () => {
			const mockView = { id: '1', name: 'Public View', records: [] };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockView),
			});

			const result = await publicViews.get('token123');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/public/views/token123',
				expect.anything()
			);
			expect(result).toEqual(mockView);
		});
	});

	describe('comments', () => {
		beforeEach(() => {
			localStorageMock.setItem('token', 'test-token');
		});

		it('should list comments', async () => {
			const mockComments = [{ id: '1', content: 'Test comment' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ comments: mockComments }),
			});

			const result = await comments.list('record-1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/records/record-1/comments',
				expect.anything()
			);
			expect(result.comments).toEqual(mockComments);
		});

		it('should create comment', async () => {
			const mockComment = { id: '1', content: 'New comment' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockComment),
			});

			const result = await comments.create('record-1', 'New comment');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/records/record-1/comments',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ content: 'New comment', parent_id: undefined }),
				})
			);
			expect(result).toEqual(mockComment);
		});

		it('should create reply comment', async () => {
			const mockComment = { id: '2', content: 'Reply', parent_id: '1' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockComment),
			});

			const result = await comments.create('record-1', 'Reply', '1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/records/record-1/comments',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ content: 'Reply', parent_id: '1' }),
				})
			);
		});

		it('should get comment', async () => {
			const mockComment = { id: '1', content: 'Test' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockComment),
			});

			const result = await comments.get('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/comments/1',
				expect.anything()
			);
		});

		it('should update comment', async () => {
			const mockComment = { id: '1', content: 'Updated' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockComment),
			});

			const result = await comments.update('1', 'Updated');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/comments/1',
				expect.objectContaining({
					method: 'PATCH',
					body: JSON.stringify({ content: 'Updated' }),
				})
			);
		});

		it('should delete comment', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Deleted' }),
			});

			const result = await comments.delete('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/comments/1',
				expect.objectContaining({
					method: 'DELETE',
				})
			);
		});

		it('should resolve comment', async () => {
			const mockComment = { id: '1', resolved: true };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockComment),
			});

			const result = await comments.resolve('1', true);

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/comments/1/resolve',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ resolved: true }),
				})
			);
		});
	});

	describe('activity', () => {
		beforeEach(() => {
			localStorageMock.setItem('token', 'test-token');
		});

		it('should list activity for base', async () => {
			const mockActivities = [{ id: '1', action: 'create' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ activities: mockActivities }),
			});

			const result = await activity.listForBase('base-1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/bases/base-1/activity',
				expect.anything()
			);
			expect(result.activities).toEqual(mockActivities);
		});

		it('should list activity with filters', async () => {
			const mockActivities = [{ id: '1', action: 'update' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ activities: mockActivities }),
			});

			const result = await activity.listForBase('base-1', { userId: 'user-1', action: 'update', limit: 10 });

			expect(mockFetch).toHaveBeenCalledWith(
				expect.stringContaining('userId=user-1'),
				expect.anything()
			);
		});

		it('should list activity for record', async () => {
			const mockActivities = [{ id: '1', action: 'update' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ activities: mockActivities }),
			});

			const result = await activity.listForRecord('record-1', 5);

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/records/record-1/activity?limit=5',
				expect.anything()
			);
		});
	});

	describe('attachments', () => {
		beforeEach(() => {
			localStorageMock.setItem('token', 'test-token');
		});

		it('should list attachments', async () => {
			const mockAttachments = [{ id: '1', filename: 'test.pdf' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ attachments: mockAttachments }),
			});

			const result = await attachments.list('record-1', 'field-1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/records/record-1/fields/field-1/attachments',
				expect.anything()
			);
			expect(result.attachments).toEqual(mockAttachments);
		});

		it('should upload attachment', async () => {
			const mockAttachment = { id: '1', filename: 'test.pdf' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockAttachment),
			});

			const file = new File(['test content'], 'test.pdf', { type: 'application/pdf' });
			const result = await attachments.upload('record-1', 'field-1', file);

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/records/record-1/fields/field-1/attachments',
				expect.objectContaining({
					method: 'POST',
				})
			);
			expect(result).toEqual(mockAttachment);
		});

		it('should get attachment', async () => {
			const mockAttachment = { id: '1', filename: 'test.pdf' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockAttachment),
			});

			const result = await attachments.get('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/attachments/1',
				expect.anything()
			);
		});

		it('should delete attachment', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Deleted' }),
			});

			const result = await attachments.delete('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/attachments/1',
				expect.objectContaining({
					method: 'DELETE',
				})
			);
		});

		it('should generate download URL', () => {
			localStorageMock.setItem('token', 'test-token');
			const url = attachments.getDownloadUrl('att-1');
			expect(url).toBe('http://localhost:8080/api/v1/attachments/att-1/download?token=test-token');
		});
	});

	describe('automations', () => {
		beforeEach(() => {
			localStorageMock.setItem('token', 'test-token');
		});

		it('should list automations', async () => {
			const mockAutomations = [{ id: '1', name: 'Auto 1' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ automations: mockAutomations }),
			});

			const result = await automations.list('table-1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/automations',
				expect.anything()
			);
			expect(result.automations).toEqual(mockAutomations);
		});

		it('should create automation', async () => {
			const mockAutomation = { id: '1', name: 'New Auto' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockAutomation),
			});

			const result = await automations.create('table-1', {
				name: 'New Auto',
				triggerType: 'record_created',
				actionType: 'send_email',
			});

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/tables/table-1/automations',
				expect.objectContaining({
					method: 'POST',
				})
			);
			expect(result).toEqual(mockAutomation);
		});

		it('should get automation', async () => {
			const mockAutomation = { id: '1', name: 'Auto 1' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockAutomation),
			});

			const result = await automations.get('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/automations/1',
				expect.anything()
			);
		});

		it('should update automation', async () => {
			const mockAutomation = { id: '1', name: 'Updated' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockAutomation),
			});

			const result = await automations.update('1', { name: 'Updated' });

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/automations/1',
				expect.objectContaining({
					method: 'PATCH',
				})
			);
		});

		it('should delete automation', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Deleted' }),
			});

			const result = await automations.delete('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/automations/1',
				expect.objectContaining({
					method: 'DELETE',
				})
			);
		});

		it('should toggle automation', async () => {
			const mockAutomation = { id: '1', enabled: true };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockAutomation),
			});

			const result = await automations.toggle('1', true);

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/automations/1/toggle',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ enabled: true }),
				})
			);
		});

		it('should list automation runs', async () => {
			const mockRuns = [{ id: '1', status: 'success' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ runs: mockRuns }),
			});

			const result = await automations.listRuns('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/automations/1/runs',
				expect.anything()
			);
			expect(result.runs).toEqual(mockRuns);
		});
	});

	describe('apiKeys', () => {
		beforeEach(() => {
			localStorageMock.setItem('token', 'test-token');
		});

		it('should list API keys', async () => {
			const mockKeys = [{ id: '1', name: 'Key 1' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ api_keys: mockKeys }),
			});

			const result = await apiKeys.list();

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/api-keys',
				expect.anything()
			);
			expect(result.api_keys).toEqual(mockKeys);
		});

		it('should create API key', async () => {
			const mockKey = { id: '1', name: 'New Key', key: 'vt_xxx' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockKey),
			});

			const result = await apiKeys.create('New Key', ['read:bases'], '2024-12-31');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/api-keys',
				expect.objectContaining({
					method: 'POST',
					body: JSON.stringify({ name: 'New Key', scopes: ['read:bases'], expires_at: '2024-12-31' }),
				})
			);
			expect(result).toEqual(mockKey);
		});

		it('should get API key', async () => {
			const mockKey = { id: '1', name: 'Key 1' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockKey),
			});

			const result = await apiKeys.get('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/api-keys/1',
				expect.anything()
			);
		});

		it('should delete API key', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Deleted' }),
			});

			const result = await apiKeys.delete('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/api-keys/1',
				expect.objectContaining({
					method: 'DELETE',
				})
			);
		});
	});

	describe('webhooks', () => {
		beforeEach(() => {
			localStorageMock.setItem('token', 'test-token');
		});

		it('should list webhooks', async () => {
			const mockWebhooks = [{ id: '1', name: 'Webhook 1' }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ webhooks: mockWebhooks }),
			});

			const result = await webhooks.list('base-1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/bases/base-1/webhooks',
				expect.anything()
			);
			expect(result.webhooks).toEqual(mockWebhooks);
		});

		it('should create webhook', async () => {
			const mockWebhook = { id: '1', name: 'New Webhook', url: 'https://example.com/hook' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockWebhook),
			});

			const result = await webhooks.create('base-1', 'New Webhook', 'https://example.com/hook', ['record.created']);

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/bases/base-1/webhooks',
				expect.objectContaining({
					method: 'POST',
				})
			);
			expect(result).toEqual(mockWebhook);
		});

		it('should get webhook', async () => {
			const mockWebhook = { id: '1', name: 'Webhook 1' };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockWebhook),
			});

			const result = await webhooks.get('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/webhooks/1',
				expect.anything()
			);
		});

		it('should update webhook', async () => {
			const mockWebhook = { id: '1', name: 'Updated', is_active: false };
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve(mockWebhook),
			});

			const result = await webhooks.update('1', { name: 'Updated', is_active: false });

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/webhooks/1',
				expect.objectContaining({
					method: 'PATCH',
				})
			);
		});

		it('should delete webhook', async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ message: 'Deleted' }),
			});

			const result = await webhooks.delete('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/webhooks/1',
				expect.objectContaining({
					method: 'DELETE',
				})
			);
		});

		it('should list webhook deliveries', async () => {
			const mockDeliveries = [{ id: '1', status: 200 }];
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: () => Promise.resolve({ deliveries: mockDeliveries }),
			});

			const result = await webhooks.listDeliveries('1');

			expect(mockFetch).toHaveBeenCalledWith(
				'http://localhost:8080/api/v1/webhooks/1/deliveries',
				expect.anything()
			);
			expect(result.deliveries).toEqual(mockDeliveries);
		});
	});

	describe('ApiError', () => {
		it('should be an instance of Error', () => {
			const error = new ApiError(404, 'not_found', 'Resource not found');
			expect(error).toBeInstanceOf(Error);
			expect(error.name).toBe('ApiError');
		});

		it('should contain status, code, and message', () => {
			const error = new ApiError(401, 'unauthorized', 'Please login');
			expect(error.status).toBe(401);
			expect(error.code).toBe('unauthorized');
			expect(error.message).toBe('Please login');
		});
	});
});
