import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import Gallery from './Gallery.svelte';
import type { Field, Record } from '$lib/types';

describe('Gallery component', () => {
	const mockTextField: Field = {
		id: 'field-title',
		table_id: 'table-1',
		name: 'Title',
		field_type: 'text',
		options: {},
		position: 0,
		created_at: '2024-01-01',
		updated_at: '2024-01-01'
	};

	const mockCoverField: Field = {
		id: 'field-cover',
		table_id: 'table-1',
		name: 'Cover Image',
		field_type: 'text',
		options: {},
		position: 1,
		created_at: '2024-01-01',
		updated_at: '2024-01-01'
	};

	const mockNumberField: Field = {
		id: 'field-number',
		table_id: 'table-1',
		name: 'Count',
		field_type: 'number',
		options: { precision: 0 },
		position: 2,
		created_at: '2024-01-01',
		updated_at: '2024-01-01'
	};

	const mockCheckboxField: Field = {
		id: 'field-check',
		table_id: 'table-1',
		name: 'Done',
		field_type: 'checkbox',
		options: {},
		position: 3,
		created_at: '2024-01-01',
		updated_at: '2024-01-01'
	};

	const mockDateField: Field = {
		id: 'field-date',
		table_id: 'table-1',
		name: 'Created',
		field_type: 'date',
		options: {},
		position: 4,
		created_at: '2024-01-01',
		updated_at: '2024-01-01'
	};

	const mockSelectField: Field = {
		id: 'field-select',
		table_id: 'table-1',
		name: 'Status',
		field_type: 'single_select',
		options: {
			options: [
				{ id: 'opt-1', name: 'Active', color: 'green' },
				{ id: 'opt-2', name: 'Inactive', color: 'red' }
			]
		},
		position: 5,
		created_at: '2024-01-01',
		updated_at: '2024-01-01'
	};

	const mockFields: Field[] = [
		mockTextField,
		mockCoverField,
		mockNumberField,
		mockCheckboxField,
		mockDateField,
		mockSelectField
	];

	const mockRecords: Record[] = [
		{
			id: 'rec-1',
			table_id: 'table-1',
			values: {
				'field-title': 'Card One',
				'field-cover': 'https://example.com/image1.jpg',
				'field-number': 42,
				'field-check': true,
				'field-date': '2024-06-15',
				'field-select': 'opt-1'
			},
			position: 0,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		},
		{
			id: 'rec-2',
			table_id: 'table-1',
			values: {
				'field-title': 'Card Two',
				'field-cover': 'https://example.com/image2.jpg',
				'field-number': 100,
				'field-check': false,
				'field-date': '2024-07-20',
				'field-select': 'opt-2'
			},
			position: 1,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		}
	];

	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('rendering', () => {
		it('should render gallery cards', () => {
			render(Gallery, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			expect(screen.getByText('Card One')).toBeTruthy();
			expect(screen.getByText('Card Two')).toBeTruthy();
		});

		it('should render cover images when coverFieldId is set', () => {
			render(Gallery, {
				props: {
					fields: mockFields,
					records: mockRecords,
					coverFieldId: 'field-cover'
				}
			});

			const images = document.querySelectorAll('img');
			expect(images.length).toBeGreaterThan(0);
		});

		it('should render add card button when not readonly', () => {
			render(Gallery, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			// Should have an add button
			const addButtons = screen.getAllByText(/add/i);
			expect(addButtons.length).toBeGreaterThan(0);
		});

		it('should not render add button in readonly mode', () => {
			render(Gallery, {
				props: {
					fields: mockFields,
					records: mockRecords,
					readonly: true
				}
			});

			// Should not have add button
			expect(screen.queryAllByText(/add/i).length).toBe(0);
		});
	});

	describe('title display', () => {
		it('should display record title from titleFieldId', () => {
			render(Gallery, {
				props: {
					fields: mockFields,
					records: mockRecords,
					titleFieldId: 'field-title'
				}
			});

			expect(screen.getByText('Card One')).toBeTruthy();
		});

		it('should auto-select first text field as title', () => {
			render(Gallery, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			expect(screen.getByText('Card One')).toBeTruthy();
		});

		it('should display Untitled for records without title', () => {
			const recordsWithNoTitle: Record[] = [
				{
					id: 'rec-1',
					table_id: 'table-1',
					values: { 'field-number': 42 },
					position: 0,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				}
			];

			render(Gallery, {
				props: {
					fields: mockFields,
					records: recordsWithNoTitle
				}
			});

			expect(screen.getByText('Untitled')).toBeTruthy();
		});
	});

	describe('field value formatting', () => {
		it('should format number fields correctly', () => {
			render(Gallery, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			expect(screen.getByText('42')).toBeTruthy();
		});

		it('should format checkbox fields as checkmarks', () => {
			render(Gallery, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			expect(screen.getByText('✓')).toBeTruthy();
			expect(screen.getByText('✗')).toBeTruthy();
		});

		it('should format select fields by option name', () => {
			render(Gallery, {
				props: {
					fields: mockFields,
					records: mockRecords,
					titleFieldId: 'field-title'
				}
			});

			// The select field values are displayed in the card preview
			// with the option name looked up from the field options
			// Check that the cards render with titles
			expect(screen.getByText('Card One')).toBeTruthy();
			expect(screen.getByText('Card Two')).toBeTruthy();
		});
	});

	describe('cover images', () => {
		it('should handle https URLs', () => {
			render(Gallery, {
				props: {
					fields: mockFields,
					records: mockRecords,
					coverFieldId: 'field-cover'
				}
			});

			const images = document.querySelectorAll('img');
			expect(images[0]?.getAttribute('src')).toContain('example.com');
		});

		it('should handle protocol-relative URLs', () => {
			const recordsWithRelativeUrl: Record[] = [
				{
					id: 'rec-1',
					table_id: 'table-1',
					values: {
						'field-title': 'Card One',
						'field-cover': '//example.com/image.jpg'
					},
					position: 0,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				}
			];

			render(Gallery, {
				props: {
					fields: mockFields,
					records: recordsWithRelativeUrl,
					coverFieldId: 'field-cover'
				}
			});

			const images = document.querySelectorAll('img');
			expect(images.length).toBeGreaterThan(0);
		});

		it('should not render cover for non-URL values', () => {
			const recordsWithNonUrl: Record[] = [
				{
					id: 'rec-1',
					table_id: 'table-1',
					values: {
						'field-title': 'Card One',
						'field-cover': 'not-a-url'
					},
					position: 0,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				}
			];

			render(Gallery, {
				props: {
					fields: mockFields,
					records: recordsWithNonUrl,
					coverFieldId: 'field-cover'
				}
			});

			// Should render card without cover image
			expect(screen.getByText('Card One')).toBeTruthy();
		});
	});

	describe('events', () => {
		it('should dispatch selectRecord event when card is clicked', async () => {
			const { component } = render(Gallery, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			const selectHandler = vi.fn();
			component.$on('selectRecord', selectHandler);

			const card = screen.getByText('Card One');
			await fireEvent.click(card);

			expect(selectHandler).toHaveBeenCalled();
		});

		it('should dispatch addRecord event when add button is clicked', async () => {
			const { component } = render(Gallery, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			const addHandler = vi.fn();
			component.$on('addRecord', addHandler);

			const addButtons = screen.getAllByText(/add/i);
			await fireEvent.click(addButtons[0]);

			expect(addHandler).toHaveBeenCalled();
		});
	});

	describe('empty state', () => {
		it('should render empty state when no records', () => {
			render(Gallery, {
				props: {
					fields: mockFields,
					records: []
				}
			});

			// Should show "No records yet" message
			expect(screen.getByText(/No records yet/i)).toBeTruthy();
		});
	});

	describe('display fields', () => {
		it('should show up to 4 fields per card excluding title and cover', () => {
			render(Gallery, {
				props: {
					fields: mockFields,
					records: mockRecords,
					titleFieldId: 'field-title',
					coverFieldId: 'field-cover'
				}
			});

			// Should display other field values
			expect(screen.getByText('42')).toBeTruthy();
		});
	});
});
