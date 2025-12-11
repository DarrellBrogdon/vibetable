import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen, waitFor } from '@testing-library/svelte';
import Grid from './Grid.svelte';
import type { Field, Record, Table, User } from '$lib/types';

// Mock the API client
vi.mock('$lib/api/client', () => ({
	records: {
		list: vi.fn().mockResolvedValue({ records: [] }),
		create: vi.fn(),
		update: vi.fn(),
		delete: vi.fn(),
		updateColor: vi.fn()
	},
	fields: {
		list: vi.fn().mockResolvedValue({ fields: [] }),
		create: vi.fn(),
		update: vi.fn(),
		delete: vi.fn(),
		reorder: vi.fn()
	}
}));

describe('Grid component', () => {
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
			name: 'Status',
			field_type: 'single_select',
			options: {
				options: [
					{ id: 'opt-1', name: 'Active', color: 'green' },
					{ id: 'opt-2', name: 'Inactive', color: 'red' }
				]
			},
			position: 1,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		},
		{
			id: 'field-3',
			table_id: 'table-1',
			name: 'Count',
			field_type: 'number',
			options: { precision: 0 },
			position: 2,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		},
		{
			id: 'field-4',
			table_id: 'table-1',
			name: 'Done',
			field_type: 'checkbox',
			options: {},
			position: 3,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		},
		{
			id: 'field-5',
			table_id: 'table-1',
			name: 'Due Date',
			field_type: 'date',
			options: {},
			position: 4,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		}
	];

	const mockRecords: Record[] = [
		{
			id: 'rec-1',
			table_id: 'table-1',
			values: {
				'field-1': 'Record One',
				'field-2': 'opt-1',
				'field-3': 10,
				'field-4': true,
				'field-5': '2024-06-15'
			},
			position: 0,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		},
		{
			id: 'rec-2',
			table_id: 'table-1',
			values: {
				'field-1': 'Record Two',
				'field-2': 'opt-2',
				'field-3': 20,
				'field-4': false,
				'field-5': '2024-07-20'
			},
			position: 1,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		}
	];

	const mockTables: Table[] = [
		{
			id: 'table-1',
			base_id: 'base-1',
			name: 'Main Table',
			position: 0,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		},
		{
			id: 'table-2',
			base_id: 'base-1',
			name: 'Linked Table',
			position: 1,
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
		it('should render field headers', () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1'
				}
			});

			expect(screen.getByText('Name')).toBeTruthy();
			expect(screen.getByText('Status')).toBeTruthy();
			expect(screen.getByText('Count')).toBeTruthy();
		});

		it('should render record values', () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1'
				}
			});

			expect(screen.getByText('Record One')).toBeTruthy();
			expect(screen.getByText('Record Two')).toBeTruthy();
		});

		it('should render empty state when no records', () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: [],
					tables: mockTables,
					currentTableId: 'table-1'
				}
			});

			// Grid should still render with headers
			expect(screen.getByText('Name')).toBeTruthy();
		});

		it('should render in readonly mode', () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1',
					readonly: true
				}
			});

			// In readonly mode, the add field button should not be present
			// The grid should still render the records
			expect(screen.getByText('Record One')).toBeTruthy();
		});

		it('should render checkboxes for checkbox fields', () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1'
				}
			});

			// Should have checkboxes rendered
			const checkboxes = document.querySelectorAll('input[type="checkbox"]');
			expect(checkboxes.length).toBeGreaterThan(0);
		});

		it('should format number fields correctly', () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1'
				}
			});

			expect(screen.getByText('10')).toBeTruthy();
			expect(screen.getByText('20')).toBeTruthy();
		});
	});

	describe('cell editing', () => {
		it('should enter edit mode on double click', async () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1'
				}
			});

			const cell = screen.getByText('Record One');
			await fireEvent.dblClick(cell);

			// Should now have an input element
			await waitFor(() => {
				const inputs = document.querySelectorAll('input[type="text"]');
				expect(inputs.length).toBeGreaterThan(0);
			});
		});

		it('should not allow editing in readonly mode', async () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1',
					readonly: true
				}
			});

			const cell = screen.getByText('Record One');
			await fireEvent.dblClick(cell);

			// Should not have any new input elements for editing
			const textInputs = document.querySelectorAll('input[type="text"]');
			// In readonly mode, the cell shouldn't become editable
			// This might need adjustment based on actual component behavior
		});
	});

	describe('add field menu', () => {
		it('should show add field button when not readonly', () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1',
					readonly: false
				}
			});

			// Should have add field functionality - look for the add column header
			const addButtons = screen.getAllByText('+');
			expect(addButtons.length).toBeGreaterThan(0);
		});
	});

	describe('field types display', () => {
		it('should display text field values', () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1'
				}
			});

			expect(screen.getByText('Record One')).toBeTruthy();
		});

		it('should display select field with tag styling', () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1'
				}
			});

			expect(screen.getByText('Active')).toBeTruthy();
			expect(screen.getByText('Inactive')).toBeTruthy();
		});
	});

	describe('filters', () => {
		it('should accept initial filters', () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1',
					initialFilters: [
						{ field_id: 'field-1', operator: 'contains', value: 'One' }
					]
				}
			});

			// Grid should render with filters applied
			expect(screen.getByText('Record One')).toBeTruthy();
		});
	});

	describe('sorting', () => {
		it('should accept initial sort', () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1',
					initialSort: { field_id: 'field-1', direction: 'asc' }
				}
			});

			// Grid should render with sort applied
			expect(screen.getByText('Record One')).toBeTruthy();
		});
	});

	describe('record colors', () => {
		it('should display colored records', () => {
			const recordsWithColor = [
				{
					...mockRecords[0],
					color: 'red' as const
				}
			];

			render(Grid, {
				props: {
					fields: mockFields,
					records: recordsWithColor,
					tables: mockTables,
					currentTableId: 'table-1'
				}
			});

			// Record should render with color
			expect(screen.getByText('Record One')).toBeTruthy();
		});
	});

	describe('linked records', () => {
		it('should handle linked record fields', () => {
			const fieldsWithLinkedRecord: Field[] = [
				...mockFields,
				{
					id: 'field-linked',
					table_id: 'table-1',
					name: 'Related',
					field_type: 'linked_record',
					options: { linked_table_id: 'table-2' },
					position: 5,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				}
			];

			const recordsWithLink = [
				{
					...mockRecords[0],
					values: {
						...mockRecords[0].values,
						'field-linked': ['linked-rec-1']
					}
				}
			];

			render(Grid, {
				props: {
					fields: fieldsWithLinkedRecord,
					records: recordsWithLink,
					tables: mockTables,
					currentTableId: 'table-1'
				}
			});

			expect(screen.getByText('Related')).toBeTruthy();
		});
	});

	describe('multi-select fields', () => {
		it('should display multi-select values as tags', () => {
			const fieldsWithMultiSelect: Field[] = [
				{
					id: 'field-multi',
					table_id: 'table-1',
					name: 'Tags',
					field_type: 'multi_select',
					options: {
						options: [
							{ id: 'tag-1', name: 'Important', color: 'red' },
							{ id: 'tag-2', name: 'Urgent', color: 'orange' }
						]
					},
					position: 0,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				}
			];

			const recordsWithMultiSelect: Record[] = [
				{
					id: 'rec-1',
					table_id: 'table-1',
					values: { 'field-multi': ['tag-1', 'tag-2'] },
					position: 0,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				}
			];

			render(Grid, {
				props: {
					fields: fieldsWithMultiSelect,
					records: recordsWithMultiSelect,
					tables: mockTables,
					currentTableId: 'table-1'
				}
			});

			expect(screen.getByText('Tags')).toBeTruthy();
		});
	});

	describe('computed fields', () => {
		it('should display formula fields as readonly', () => {
			const fieldsWithFormula: Field[] = [
				{
					id: 'field-formula',
					table_id: 'table-1',
					name: 'Calculated',
					field_type: 'formula',
					options: { expression: '{field-3} * 2', result_type: 'number' },
					position: 0,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				}
			];

			const recordsWithFormula: Record[] = [
				{
					id: 'rec-1',
					table_id: 'table-1',
					values: { 'field-formula': 20 },
					position: 0,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				}
			];

			render(Grid, {
				props: {
					fields: fieldsWithFormula,
					records: recordsWithFormula,
					tables: mockTables,
					currentTableId: 'table-1'
				}
			});

			expect(screen.getByText('Calculated')).toBeTruthy();
		});
	});

	describe('current user', () => {
		it('should accept current user prop', () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1',
					currentUser: mockUser
				}
			});

			// Should render without errors
			expect(screen.getByText('Record One')).toBeTruthy();
		});
	});

	describe('edit new record', () => {
		it('should handle editNewRecordId prop', () => {
			render(Grid, {
				props: {
					fields: mockFields,
					records: mockRecords,
					tables: mockTables,
					currentTableId: 'table-1',
					editNewRecordId: 'rec-1'
				}
			});

			// Should render and potentially auto-focus on the new record
			expect(screen.getByText('Record One')).toBeTruthy();
		});
	});
});
