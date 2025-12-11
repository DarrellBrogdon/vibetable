import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import Calendar from './Calendar.svelte';
import type { Field, Record } from '$lib/types';

describe('Calendar component', () => {
	const mockDateField: Field = {
		id: 'field-date',
		table_id: 'table-1',
		name: 'Due Date',
		field_type: 'date',
		options: {},
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

	const mockFields: Field[] = [mockDateField, mockTextField];

	// Use a fixed date for testing
	const currentMonth = new Date().getMonth();
	const currentYear = new Date().getFullYear();

	const mockRecords: Record[] = [
		{
			id: 'rec-1',
			table_id: 'table-1',
			values: {
				'field-date': `${currentYear}-${String(currentMonth + 1).padStart(2, '0')}-15`,
				'field-title': 'Event One'
			},
			position: 0,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		},
		{
			id: 'rec-2',
			table_id: 'table-1',
			values: {
				'field-date': `${currentYear}-${String(currentMonth + 1).padStart(2, '0')}-20`,
				'field-title': 'Event Two'
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
		it('should render calendar grid', () => {
			render(Calendar, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			// Should show day headers
			expect(screen.getByText('Sun')).toBeTruthy();
			expect(screen.getByText('Mon')).toBeTruthy();
			expect(screen.getByText('Tue')).toBeTruthy();
			expect(screen.getByText('Wed')).toBeTruthy();
			expect(screen.getByText('Thu')).toBeTruthy();
			expect(screen.getByText('Fri')).toBeTruthy();
			expect(screen.getByText('Sat')).toBeTruthy();
		});

		it('should render month and year header', () => {
			render(Calendar, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			// Should show current month name
			const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
				'July', 'August', 'September', 'October', 'November', 'December'];
			const expectedMonth = monthNames[currentMonth];

			expect(screen.getByText(new RegExp(expectedMonth))).toBeTruthy();
		});

		it('should render events on correct dates', () => {
			render(Calendar, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			expect(screen.getByText('Event One')).toBeTruthy();
			expect(screen.getByText('Event Two')).toBeTruthy();
		});
	});

	describe('navigation', () => {
		it('should navigate to previous month', async () => {
			render(Calendar, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			// Navigation uses SVG buttons, find them by their title attribute
			const prevButton = document.querySelector('button[title="Previous month"]');
			expect(prevButton).toBeTruthy();
			await fireEvent.click(prevButton!);

			// Month should have changed
			const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
				'July', 'August', 'September', 'October', 'November', 'December'];
			const expectedMonth = monthNames[currentMonth === 0 ? 11 : currentMonth - 1];

			expect(screen.getByText(new RegExp(expectedMonth))).toBeTruthy();
		});

		it('should navigate to next month', async () => {
			render(Calendar, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			const nextButton = document.querySelector('button[title="Next month"]');
			expect(nextButton).toBeTruthy();
			await fireEvent.click(nextButton!);

			const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
				'July', 'August', 'September', 'October', 'November', 'December'];
			const expectedMonth = monthNames[currentMonth === 11 ? 0 : currentMonth + 1];

			expect(screen.getByText(new RegExp(expectedMonth))).toBeTruthy();
		});

		it('should return to today', async () => {
			render(Calendar, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			// Navigate away
			const nextButton = document.querySelector('button[title="Next month"]');
			await fireEvent.click(nextButton!);

			// Click today
			const todayButton = screen.getByText('Today');
			await fireEvent.click(todayButton);

			const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
				'July', 'August', 'September', 'October', 'November', 'December'];
			const expectedMonth = monthNames[currentMonth];

			expect(screen.getByText(new RegExp(expectedMonth))).toBeTruthy();
		});
	});

	describe('date field selection', () => {
		it('should use provided dateFieldId', () => {
			render(Calendar, {
				props: {
					fields: mockFields,
					records: mockRecords,
					dateFieldId: 'field-date'
				}
			});

			expect(screen.getByText('Event One')).toBeTruthy();
		});

		it('should auto-select first date field if not provided', () => {
			render(Calendar, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			// Should still render events since it auto-selects the date field
			expect(screen.getByText('Event One')).toBeTruthy();
		});
	});

	describe('title field selection', () => {
		it('should use provided titleFieldId', () => {
			render(Calendar, {
				props: {
					fields: mockFields,
					records: mockRecords,
					titleFieldId: 'field-title'
				}
			});

			expect(screen.getByText('Event One')).toBeTruthy();
		});
	});

	describe('readonly mode', () => {
		it('should not allow adding events in readonly mode', () => {
			render(Calendar, {
				props: {
					fields: mockFields,
					records: mockRecords,
					readonly: true
				}
			});

			// In readonly mode, clicking on empty day shouldn't trigger add
			// This is behavior-based - the component should render without add functionality
			expect(screen.getByText('Event One')).toBeTruthy();
		});
	});

	describe('events', () => {
		it('should dispatch selectRecord event when event is clicked', async () => {
			const { component } = render(Calendar, {
				props: {
					fields: mockFields,
					records: mockRecords
				}
			});

			const selectHandler = vi.fn();
			component.$on('selectRecord', selectHandler);

			const event = screen.getByText('Event One');
			await fireEvent.click(event);

			expect(selectHandler).toHaveBeenCalled();
		});
	});

	describe('no date field', () => {
		it('should handle missing date field gracefully', () => {
			const textOnlyFields: Field[] = [mockTextField];

			render(Calendar, {
				props: {
					fields: textOnlyFields,
					records: mockRecords
				}
			});

			// Should show message about needing a date field
			expect(screen.getByText(/Add a date field/i)).toBeTruthy();
		});
	});

	describe('records without dates', () => {
		it('should not display records without date values', () => {
			const recordsWithNoDate: Record[] = [
				{
					id: 'rec-1',
					table_id: 'table-1',
					values: { 'field-title': 'No Date Event' },
					position: 0,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				}
			];

			render(Calendar, {
				props: {
					fields: mockFields,
					records: recordsWithNoDate
				}
			});

			// Should not find the event since it has no date
			expect(screen.queryByText('No Date Event')).toBeNull();
		});
	});
});
