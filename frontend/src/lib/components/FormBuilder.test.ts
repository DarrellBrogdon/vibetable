import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen, waitFor } from '@testing-library/svelte';
import FormBuilder from './FormBuilder.svelte';
import type { Form, FormField } from '$lib/types';

// Mock the API client
vi.mock('$lib/api/client', () => ({
	forms: {
		update: vi.fn().mockResolvedValue({
			id: 'form-1',
			name: 'Updated Form',
			is_active: true
		}),
		updateFields: vi.fn().mockResolvedValue({}),
		delete: vi.fn().mockResolvedValue({})
	}
}));

// Mock window.location.origin
Object.defineProperty(window, 'location', {
	value: {
		origin: 'http://localhost:5173'
	},
	writable: true
});

// Mock navigator.clipboard
Object.assign(navigator, {
	clipboard: {
		writeText: vi.fn().mockResolvedValue(undefined)
	}
});

describe('FormBuilder component', () => {
	const mockFormFields: FormField[] = [
		{
			id: 'ff-1',
			form_id: 'form-1',
			field_id: 'field-1',
			field_name: 'Name',
			field_type: 'text',
			label: 'Your Name',
			help_text: 'Enter your full name',
			is_required: true,
			is_visible: true,
			position: 0,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		},
		{
			id: 'ff-2',
			form_id: 'form-1',
			field_id: 'field-2',
			field_name: 'Email',
			field_type: 'text',
			label: 'Email Address',
			is_required: false,
			is_visible: true,
			position: 1,
			created_at: '2024-01-01',
			updated_at: '2024-01-01'
		}
	];

	const mockForm: Form = {
		id: 'form-1',
		table_id: 'table-1',
		name: 'Contact Form',
		description: 'A test form',
		is_active: true,
		public_token: 'abc123',
		success_message: 'Thank you!',
		submit_button_text: 'Submit',
		fields: mockFormFields,
		created_at: '2024-01-01',
		updated_at: '2024-01-01'
	};

	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('rendering', () => {
		it('should render form builder header', () => {
			render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			expect(screen.getByText('Edit Form')).toBeTruthy();
		});

		it('should show form name input', () => {
			render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			const nameInput = document.querySelector('#name') as HTMLInputElement;
			expect(nameInput).toBeTruthy();
			expect(nameInput.value).toBe('Contact Form');
		});

		it('should show form settings section', () => {
			render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			expect(screen.getByText('Form Settings')).toBeTruthy();
		});

		it('should show public link section', () => {
			render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			expect(screen.getByText('Public Link')).toBeTruthy();
		});

		it('should show form fields section', () => {
			render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			expect(screen.getByText('Form Fields')).toBeTruthy();
		});

		it('should display field names', () => {
			render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			expect(screen.getByText('Name')).toBeTruthy();
			expect(screen.getByText('Email')).toBeTruthy();
		});
	});

	describe('form actions', () => {
		it('should have save button', () => {
			render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			expect(screen.getByText('Save Changes')).toBeTruthy();
		});

		it('should have cancel button', () => {
			render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			expect(screen.getByText('Cancel')).toBeTruthy();
		});

		it('should have delete button', () => {
			render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			expect(screen.getByText('Delete Form')).toBeTruthy();
		});
	});

	describe('public URL', () => {
		it('should display public URL when token exists', () => {
			render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			const urlInput = document.querySelector('.url-box input') as HTMLInputElement;
			expect(urlInput).toBeTruthy();
			expect(urlInput.value).toContain('abc123');
		});

		it('should have copy button for URL', () => {
			render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			expect(screen.getByText('Copy')).toBeTruthy();
		});
	});

	describe('field configuration', () => {
		it('should show visible checkbox for fields', () => {
			render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			const visibleLabels = screen.getAllByText('Visible');
			expect(visibleLabels.length).toBeGreaterThan(0);
		});

		it('should show required checkbox for fields', () => {
			render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			const requiredLabels = screen.getAllByText('Required');
			expect(requiredLabels.length).toBeGreaterThan(0);
		});
	});

	describe('events', () => {
		it('should dispatch close event when cancel clicked', async () => {
			const { component } = render(FormBuilder, {
				props: {
					form: mockForm,
					tableId: 'table-1'
				}
			});

			const closeHandler = vi.fn();
			component.$on('close', closeHandler);

			const cancelButton = screen.getByText('Cancel');
			await fireEvent.click(cancelButton);

			expect(closeHandler).toHaveBeenCalled();
		});
	});
});
