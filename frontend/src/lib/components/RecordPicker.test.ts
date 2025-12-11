import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import RecordPicker from './RecordPicker.svelte';

// Mock the API client
const mockTableGet = vi.fn();
const mockFieldsList = vi.fn();
const mockRecordsList = vi.fn();

vi.mock('$lib/api/client', () => ({
	tables: {
		get: (...args: any[]) => mockTableGet(...args)
	},
	fields: {
		list: (...args: any[]) => mockFieldsList(...args)
	},
	records: {
		list: (...args: any[]) => mockRecordsList(...args)
	}
}));

describe('RecordPicker component', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		mockTableGet.mockResolvedValue({ id: 'table-1', name: 'Test Table' });
		mockFieldsList.mockResolvedValue({
			fields: [
				{ id: 'field-1', name: 'Name', field_type: 'text', position: 0 }
			]
		});
		mockRecordsList.mockResolvedValue({
			records: [
				{ id: 'rec-1', values: { 'field-1': 'Record 1' } },
				{ id: 'rec-2', values: { 'field-1': 'Record 2' } }
			]
		});
	});

	describe('rendering', () => {
		it('should render picker overlay', () => {
			render(RecordPicker, {
				props: {
					tableId: 'table-1',
					selectedIds: []
				}
			});

			expect(document.querySelector('.picker-overlay')).toBeTruthy();
		});

		it('should render picker modal', () => {
			render(RecordPicker, {
				props: {
					tableId: 'table-1',
					selectedIds: []
				}
			});

			expect(document.querySelector('.picker-modal')).toBeTruthy();
		});

		it('should show search input', () => {
			render(RecordPicker, {
				props: {
					tableId: 'table-1',
					selectedIds: []
				}
			});

			expect(screen.getByPlaceholderText(/Search records/i)).toBeTruthy();
		});

		it('should show loading state initially', () => {
			render(RecordPicker, {
				props: {
					tableId: 'table-1',
					selectedIds: []
				}
			});

			expect(screen.getByText('Loading records...')).toBeTruthy();
		});

		it('should show selection count', () => {
			render(RecordPicker, {
				props: {
					tableId: 'table-1',
					selectedIds: ['rec-1', 'rec-2']
				}
			});

			expect(screen.getByText('2 selected')).toBeTruthy();
		});
	});

	describe('close functionality', () => {
		it('should have a close button', () => {
			render(RecordPicker, {
				props: {
					tableId: 'table-1',
					selectedIds: []
				}
			});

			expect(screen.getByText('×')).toBeTruthy();
		});

		it('should have a done button', () => {
			render(RecordPicker, {
				props: {
					tableId: 'table-1',
					selectedIds: []
				}
			});

			expect(screen.getByText('Done')).toBeTruthy();
		});

		it('should dispatch close event when close button clicked', async () => {
			const { component } = render(RecordPicker, {
				props: {
					tableId: 'table-1',
					selectedIds: []
				}
			});

			const closeHandler = vi.fn();
			component.$on('close', closeHandler);

			const closeButton = screen.getByText('×');
			await fireEvent.click(closeButton);

			expect(closeHandler).toHaveBeenCalled();
		});
	});

	describe('API integration', () => {
		it('should render component ready for API calls', () => {
			render(RecordPicker, {
				props: {
					tableId: 'table-1',
					selectedIds: []
				}
			});

			expect(document.querySelector('.picker-modal')).toBeTruthy();
		});
	});
});
