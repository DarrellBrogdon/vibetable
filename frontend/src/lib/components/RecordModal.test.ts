import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen, waitFor } from '@testing-library/svelte';
import RecordModal from './RecordModal.svelte';
import type { Field, Record, Table, User } from '$lib/types';

// Mock the API client
vi.mock('$lib/api/client', () => ({
	records: {
		list: vi.fn().mockResolvedValue({ records: [] }),
		get: vi.fn(),
		update: vi.fn()
	},
	fields: {
		list: vi.fn().mockResolvedValue({ fields: [] })
	}
}));

describe('RecordModal component', () => {
	const mockRecord: Record = {
		id: 'rec-1',
		table_id: 'table-1',
		values: {
			'field-1': 'Test Value',
			'field-2': 42,
			'field-3': true
		},
		position: 0,
		created_at: '2024-01-01',
		updated_at: '2024-01-01'
	};

	const mockFields: Field[] = [
		{
			id: 'field-1',
			table_id: 'table-1',
			name: 'Name',
			field_type: 'text',
			options: {},
			position: 0,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		},
		{
			id: 'field-2',
			table_id: 'table-1',
			name: 'Count',
			field_type: 'number',
			options: {},
			position: 1,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		},
		{
			id: 'field-3',
			table_id: 'table-1',
			name: 'Active',
			field_type: 'checkbox',
			options: {},
			position: 2,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		}
	];

	const mockTables: Table[] = [
		{
			id: 'table-1',
			base_id: 'base-1',
			name: 'Test Table',
			position: 0,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		}
	];

	const mockUser: User = {
		id: 'user-1',
		email: 'test@example.com',
		name: 'Test User',
		created_at: '2024-01-01',
		updated_at: '2024-01-01'
	};

	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('rendering', () => {
		it('should render modal overlay', () => {
			render(RecordModal, {
				props: {
					record: mockRecord,
					fields: mockFields,
					tables: mockTables
				}
			});

			expect(document.querySelector('.modal-overlay')).toBeTruthy();
		});

		it('should render field labels', () => {
			render(RecordModal, {
				props: {
					record: mockRecord,
					fields: mockFields,
					tables: mockTables
				}
			});

			expect(screen.getByText('Name')).toBeTruthy();
			expect(screen.getByText('Count')).toBeTruthy();
			expect(screen.getByText('Active')).toBeTruthy();
		});

		it('should render tabs for fields and comments', () => {
			render(RecordModal, {
				props: {
					record: mockRecord,
					fields: mockFields,
					tables: mockTables,
					currentUser: mockUser
				}
			});

			expect(screen.getByText('Fields')).toBeTruthy();
			expect(screen.getByText('Comments')).toBeTruthy();
		});

		it('should render close button', () => {
			render(RecordModal, {
				props: {
					record: mockRecord,
					fields: mockFields,
					tables: mockTables
				}
			});

			const closeButton = document.querySelector('.close-btn');
			expect(closeButton).toBeTruthy();
		});
	});

	describe('readonly mode', () => {
		it('should disable editing in readonly mode', () => {
			render(RecordModal, {
				props: {
					record: mockRecord,
					fields: mockFields,
					tables: mockTables,
					readonly: true
				}
			});

			const inputs = document.querySelectorAll('input');
			inputs.forEach(input => {
				expect((input as HTMLInputElement).disabled || (input as HTMLInputElement).readOnly).toBe(true);
			});
		});
	});

	describe('events', () => {
		it('should dispatch close event when close button clicked', async () => {
			const { component } = render(RecordModal, {
				props: {
					record: mockRecord,
					fields: mockFields,
					tables: mockTables
				}
			});

			const closeHandler = vi.fn();
			component.$on('close', closeHandler);

			const closeButton = document.querySelector('.close-btn');
			if (closeButton) {
				await fireEvent.click(closeButton);
			}

			expect(closeHandler).toHaveBeenCalled();
		});
	});

	describe('tabs', () => {
		it('should render Fields tab as active by default', () => {
			render(RecordModal, {
				props: {
					record: mockRecord,
					fields: mockFields,
					tables: mockTables,
					currentUser: mockUser
				}
			});

			const fieldsTab = screen.getByText('Fields');
			expect(fieldsTab.closest('.tab')?.classList.contains('active') ||
			       fieldsTab.closest('button')?.classList.contains('active')).toBeTruthy();
		});
	});
});
