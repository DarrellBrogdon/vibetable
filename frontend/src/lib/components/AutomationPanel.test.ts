import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen, waitFor } from '@testing-library/svelte';
import AutomationPanel from './AutomationPanel.svelte';
import type { Automation, Field } from '$lib/types';

// Mock the API client
vi.mock('$lib/api/client', () => ({
	automations: {
		list: vi.fn().mockResolvedValue({
			automations: [
				{
					id: 'auto-1',
					table_id: 'table-1',
					name: 'Notify on create',
					trigger_type: 'record_created',
					action_type: 'send_webhook',
					action_config: { url: 'https://example.com/webhook' },
					enabled: true,
					run_count: 10,
					last_triggered_at: '2024-01-01T00:00:00Z',
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				},
				{
					id: 'auto-2',
					table_id: 'table-1',
					name: 'Update on change',
					trigger_type: 'record_updated',
					action_type: 'update_record',
					action_config: {},
					enabled: false,
					run_count: 5,
					created_at: '2024-01-01',
					updated_at: '2024-01-01'
				}
			]
		}),
		create: vi.fn().mockResolvedValue({
			id: 'auto-3',
			name: 'New Automation'
		}),
		toggle: vi.fn().mockResolvedValue({}),
		delete: vi.fn().mockResolvedValue({})
	}
}));

// Mock confirm
global.confirm = vi.fn(() => true);

describe('AutomationPanel component', () => {
	const mockFields: Field[] = [
		{ id: 'field-1', table_id: 'table-1', name: 'Name', field_type: 'text', options: {}, position: 0, created_at: '', updated_at: '' }
	];

	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('rendering', () => {
		it('should render panel header', async () => {
			render(AutomationPanel, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			expect(screen.getByText('Automations')).toBeTruthy();
		});

		it('should show new automation button', () => {
			render(AutomationPanel, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			expect(screen.getByText('+ New Automation')).toBeTruthy();
		});

		it('should load and display automations', async () => {
			render(AutomationPanel, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			await waitFor(() => {
				expect(screen.getByText('Notify on create')).toBeTruthy();
				expect(screen.getByText('Update on change')).toBeTruthy();
			});
		});

		it('should display trigger labels', async () => {
			render(AutomationPanel, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			await waitFor(() => {
				expect(screen.getByText('When a record is created')).toBeTruthy();
				expect(screen.getByText('When a record is updated')).toBeTruthy();
			});
		});

		it('should display action labels', async () => {
			render(AutomationPanel, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			await waitFor(() => {
				expect(screen.getByText('Send a webhook')).toBeTruthy();
				expect(screen.getByText('Update the record')).toBeTruthy();
			});
		});

		it('should display run counts', async () => {
			render(AutomationPanel, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			await waitFor(() => {
				expect(screen.getByText(/Runs: 10/)).toBeTruthy();
				expect(screen.getByText(/Runs: 5/)).toBeTruthy();
			});
		});
	});

	describe('create form', () => {
		it('should toggle create form when button clicked', async () => {
			render(AutomationPanel, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			const newButton = screen.getByText('+ New Automation');
			await fireEvent.click(newButton);

			expect(screen.getByText('Cancel')).toBeTruthy();
		});

		it('should show form fields when creating', async () => {
			render(AutomationPanel, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			const newButton = screen.getByText('+ New Automation');
			await fireEvent.click(newButton);

			expect(screen.getByLabelText('Name')).toBeTruthy();
			expect(screen.getByLabelText(/When this happens/)).toBeTruthy();
			expect(screen.getByLabelText(/Do this/)).toBeTruthy();
		});

		it('should show webhook URL field when webhook action selected', async () => {
			render(AutomationPanel, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			const newButton = screen.getByText('+ New Automation');
			await fireEvent.click(newButton);

			expect(screen.getByLabelText('Webhook URL')).toBeTruthy();
		});
	});

	describe('automation actions', () => {
		it('should have toggle switches', async () => {
			render(AutomationPanel, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			await waitFor(() => {
				const toggles = document.querySelectorAll('.toggle');
				expect(toggles.length).toBeGreaterThan(0);
			});
		});

		it('should have delete buttons', async () => {
			render(AutomationPanel, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			await waitFor(() => {
				const deleteButtons = document.querySelectorAll('.btn-icon');
				expect(deleteButtons.length).toBeGreaterThan(0);
			});
		});
	});

	describe('empty state', () => {
		it('should show empty message when no automations', async () => {
			vi.mocked(await import('$lib/api/client')).automations.list.mockResolvedValueOnce({
				automations: []
			});

			render(AutomationPanel, {
				props: {
					tableId: 'table-1',
					fields: mockFields
				}
			});

			await waitFor(() => {
				expect(screen.getByText('No automations yet.')).toBeTruthy();
			});
		});
	});
});
