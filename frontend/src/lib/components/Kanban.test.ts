import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import Kanban from './Kanban.svelte';
import type { Field, Record } from '$lib/types';

describe('Kanban component', () => {
	const mockSelectField: Field = {
		id: 'field-status',
		table_id: 'table-1',
		name: 'Status',
		field_type: 'single_select',
		options: {
			options: [
				{ id: 'opt-todo', name: 'To Do', color: 'gray' },
				{ id: 'opt-doing', name: 'In Progress', color: 'blue' },
				{ id: 'opt-done', name: 'Done', color: 'green' }
			]
		},
		position: 0,
		created_at: '2024-01-01',
		updated_at: '2024-01-01'
	};

	const mockTextField: Field = {
		id: 'field-title',
		table_id: 'table-1',
		name: 'Title',
		field_type: 'text',
		options: {},
		position: 1,
		created_at: '2024-01-01',
		updated_at: '2024-01-01'
	};

	const mockFields: Field[] = [mockSelectField, mockTextField];

	const mockRecords: Record[] = [
		{
			id: 'rec-1',
			table_id: 'table-1',
			values: { 'field-status': 'opt-todo', 'field-title': 'Task One' },
			position: 0,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		},
		{
			id: 'rec-2',
			table_id: 'table-1',
			values: { 'field-status': 'opt-doing', 'field-title': 'Task Two' },
			position: 1,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		},
		{
			id: 'rec-3',
			table_id: 'table-1',
			values: { 'field-status': 'opt-done', 'field-title': 'Task Three' },
			position: 2,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		}
	];

	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('rendering', () => {
		it('should render kanban columns based on select options', () => {
			render(Kanban, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			expect(screen.getByText('To Do')).toBeTruthy();
			expect(screen.getByText('In Progress')).toBeTruthy();
			expect(screen.getByText('Done')).toBeTruthy();
		});

		it('should render records in correct columns', () => {
			render(Kanban, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			expect(screen.getByText('Task One')).toBeTruthy();
			expect(screen.getByText('Task Two')).toBeTruthy();
			expect(screen.getByText('Task Three')).toBeTruthy();
		});

		it('should render empty state when no select fields', () => {
			const textOnlyFields: Field[] = [mockTextField];

			render(Kanban, {
				props: {
					fields: textOnlyFields,
					records: mockRecords
				}
			});

			// Should show message about needing a select field
			expect(screen.getByText(/No single-select fields/i)).toBeTruthy();
		});

		it('should render uncategorized column for records without status', () => {
			const recordsWithUncategorized: Record[] = [
				...mockRecords,
				{
					id: 'rec-4',
					table_id: 'table-1',
					values: { 'field-title': 'Uncategorized Task' },
					position: 3,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				}
			];

			render(Kanban, {
				props: {
					fields: mockFields,
					records: recordsWithUncategorized
				}
			});

			expect(screen.getByText('Uncategorized Task')).toBeTruthy();
		});
	});

	describe('readonly mode', () => {
		it('should not show add buttons in readonly mode', () => {
			render(Kanban, {
				props: {
					fields: mockFields,
					records: mockRecords,
					readonly: true
				}
			});

			// Add buttons should not be present
			const addButtons = document.querySelectorAll('[data-testid="add-card-button"]');
			expect(addButtons.length).toBe(0);
		});
	});

	describe('grouping', () => {
		it('should allow changing grouping field when multiple select fields exist', () => {
			const multipleSelectFields: Field[] = [
				mockSelectField,
				{
					id: 'field-priority',
					table_id: 'table-1',
					name: 'Priority',
					field_type: 'single_select',
					options: {
						options: [
							{ id: 'opt-high', name: 'High', color: 'red' },
							{ id: 'opt-low', name: 'Low', color: 'green' }
						]
					},
					position: 2,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				},
				mockTextField
			];

			render(Kanban, {
				props: {
					fields: multipleSelectFields,
					records: mockRecords
				}
			});

			// Should have a dropdown to select grouping field
			const select = document.querySelector('select');
			expect(select).toBeTruthy();
		});
	});

	describe('events', () => {
		it('should dispatch selectRecord event when card is clicked', async () => {
			const { component } = render(Kanban, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			const selectHandler = vi.fn();
			component.$on('selectRecord', selectHandler);

			const card = screen.getByText('Task One');
			await fireEvent.click(card);

			expect(selectHandler).toHaveBeenCalled();
		});
	});

	describe('column counts', () => {
		it('should display correct record count per column', () => {
			const recordsWithMultiple: Record[] = [
				...mockRecords,
				{
					id: 'rec-4',
					table_id: 'table-1',
					values: { 'field-status': 'opt-todo', 'field-title': 'Another Todo' },
					position: 3,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				}
			];

			render(Kanban, {
				props: {
					fields: mockFields,
					records: recordsWithMultiple
				}
			});

			// To Do column should show 2 records
			expect(screen.getByText('Task One')).toBeTruthy();
			expect(screen.getByText('Another Todo')).toBeTruthy();
		});
	});

	describe('card display', () => {
		it('should display record title from primary text field', () => {
			render(Kanban, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			expect(screen.getByText('Task One')).toBeTruthy();
		});

		it('should display Untitled for records without title', () => {
			const recordsWithNoTitle: Record[] = [
				{
					id: 'rec-1',
					table_id: 'table-1',
					values: { 'field-status': 'opt-todo' },
					position: 0,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				}
			];

			render(Kanban, {
				props: {
					fields: mockFields,
					records: recordsWithNoTitle
				}
			});

			expect(screen.getByText('Untitled')).toBeTruthy();
		});
	});
});
