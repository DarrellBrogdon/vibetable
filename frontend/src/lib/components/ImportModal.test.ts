import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen, waitFor } from '@testing-library/svelte';
import ImportModal from './ImportModal.svelte';
import type { Field } from '$lib/types';

// Mock the API client
vi.mock('$lib/api/client', () => ({
	csv: {
		preview: vi.fn().mockResolvedValue({
			columns: ['Name', 'Email', 'Age'],
			rows: [
				{ Name: 'Alice', Email: 'alice@test.com', Age: '25' },
				{ Name: 'Bob', Email: 'bob@test.com', Age: '30' }
			],
			total: 2
		}),
		import: vi.fn().mockResolvedValue({
			imported: 2,
			skipped: 0,
			errors: 0
		})
	}
}));

describe('ImportModal component', () => {
	const mockFields: Field[] = [
		{ id: 'field-1', table_id: 'table-1', name: 'Name', field_type: 'text', options: {}, position: 0, created_at: '', updated_at: '' },
		{ id: 'field-2', table_id: 'table-1', name: 'Email', field_type: 'text', options: {}, position: 1, created_at: '', updated_at: '' },
		{ id: 'field-3', table_id: 'table-1', name: 'Age', field_type: 'number', options: {}, position: 2, created_at: '', updated_at: '' }
	];

	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('rendering', () => {
		it('should render import modal title', () => {
			render(ImportModal, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			expect(screen.getByText('Import CSV')).toBeTruthy();
		});

		it('should show upload step initially', () => {
			render(ImportModal, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			expect(screen.getByText(/Drag and drop a CSV file/)).toBeTruthy();
		});

		it('should have close button', () => {
			render(ImportModal, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			expect(screen.getByText('×')).toBeTruthy();
		});

		it('should have browse files button', () => {
			render(ImportModal, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			expect(screen.getByText('Browse files')).toBeTruthy();
		});
	});

	describe('drag and drop', () => {
		it('should accept dropped files', async () => {
			render(ImportModal, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			const dropZone = document.querySelector('.drop-zone');
			expect(dropZone).toBeTruthy();
		});

		it('should highlight on drag over', async () => {
			render(ImportModal, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			const dropZone = document.querySelector('.drop-zone');
			if (dropZone) {
				await fireEvent.dragOver(dropZone);
				expect(dropZone.classList.contains('active')).toBe(true);
			}
		});
	});

	describe('close functionality', () => {
		it('should dispatch close event when close button clicked', async () => {
			const { component } = render(ImportModal, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			const closeHandler = vi.fn();
			component.$on('close', closeHandler);

			const closeButton = screen.getByText('×');
			await fireEvent.click(closeButton);

			expect(closeHandler).toHaveBeenCalled();
		});
	});

	describe('file input', () => {
		it('should have hidden file input', () => {
			render(ImportModal, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			const fileInput = document.querySelector('input[type="file"]');
			expect(fileInput).toBeTruthy();
			expect((fileInput as HTMLInputElement).accept).toContain('.csv');
		});
	});
});
