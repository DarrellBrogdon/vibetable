import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen, waitFor } from '@testing-library/svelte';
import WebhooksPanel from './WebhooksPanel.svelte';

// Mock the API client
const mockWebhooksList = vi.fn();
const mockWebhooksCreate = vi.fn();
const mockWebhooksUpdate = vi.fn();
const mockWebhooksDelete = vi.fn();
const mockWebhooksListDeliveries = vi.fn();

vi.mock('$lib/api/client', () => ({
	webhooks: {
		list: (...args: any[]) => mockWebhooksList(...args),
		create: (...args: any[]) => mockWebhooksCreate(...args),
		update: (...args: any[]) => mockWebhooksUpdate(...args),
		delete: (...args: any[]) => mockWebhooksDelete(...args),
		listDeliveries: (...args: any[]) => mockWebhooksListDeliveries(...args)
	}
}));

// Mock confirm
global.confirm = vi.fn(() => true);

describe('WebhooksPanel component', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		mockWebhooksList.mockResolvedValue({
			webhooks: [
				{
					id: 'webhook-1',
					base_id: 'base-1',
					name: 'Slack Notification',
					url: 'https://hooks.slack.com/services/xxx',
					events: ['record.created', 'record.updated'],
					is_active: true,
					created_at: '2024-01-01T00:00:00Z'
				},
				{
					id: 'webhook-2',
					base_id: 'base-1',
					name: 'Zapier Integration',
					url: 'https://hooks.zapier.com/xxx',
					events: ['record.deleted'],
					is_active: false,
					created_at: '2024-01-01T00:00:00Z'
				}
			]
		});
		mockWebhooksCreate.mockResolvedValue({
			id: 'webhook-3',
			name: 'New Webhook'
		});
		mockWebhooksUpdate.mockResolvedValue({});
		mockWebhooksDelete.mockResolvedValue({});
		mockWebhooksListDeliveries.mockResolvedValue({
			deliveries: [
				{
					id: 'delivery-1',
					webhook_id: 'webhook-1',
					event_type: 'record.created',
					response_status: 200,
					duration_ms: 150,
					delivered_at: '2024-01-01T00:00:00Z'
				}
			]
		});
	});

	describe('rendering', () => {
		it('should render panel header', () => {
			render(WebhooksPanel, {
				props: {
					baseId: 'base-1'
				}
			});

			expect(screen.getByText('Webhooks')).toBeTruthy();
		});

		it('should show new webhook button', () => {
			render(WebhooksPanel, {
				props: {
					baseId: 'base-1'
				}
			});

			expect(screen.getByText('+ New Webhook')).toBeTruthy();
		});

		it('should show loading state initially', () => {
			render(WebhooksPanel, {
				props: {
					baseId: 'base-1'
				}
			});

			expect(screen.getByText('Loading webhooks...')).toBeTruthy();
		});
	});

	describe('create form', () => {
		it('should toggle create form when button clicked', async () => {
			render(WebhooksPanel, {
				props: {
					baseId: 'base-1'
				}
			});

			const newButton = screen.getByText('+ New Webhook');
			await fireEvent.click(newButton);

			expect(screen.getByText('Cancel')).toBeTruthy();
		});

		it('should show form inputs when creating', async () => {
			render(WebhooksPanel, {
				props: {
					baseId: 'base-1'
				}
			});

			const newButton = screen.getByText('+ New Webhook');
			await fireEvent.click(newButton);

			// Check for input fields by their id
			expect(document.querySelector('#name')).toBeTruthy();
			expect(document.querySelector('#url')).toBeTruthy();
		});

		it('should show event checkboxes', async () => {
			render(WebhooksPanel, {
				props: {
					baseId: 'base-1'
				}
			});

			const newButton = screen.getByText('+ New Webhook');
			await fireEvent.click(newButton);

			expect(screen.getByText('Record Created')).toBeTruthy();
			expect(screen.getByText('Record Updated')).toBeTruthy();
			expect(screen.getByText('Record Deleted')).toBeTruthy();
		});

		it('should show secret field', async () => {
			render(WebhooksPanel, {
				props: {
					baseId: 'base-1'
				}
			});

			const newButton = screen.getByText('+ New Webhook');
			await fireEvent.click(newButton);

			expect(document.querySelector('#secret')).toBeTruthy();
		});
	});

	describe('API integration', () => {
		it('should call list API on mount', () => {
			render(WebhooksPanel, {
				props: {
					baseId: 'base-1'
				}
			});

			expect(mockWebhooksList).toHaveBeenCalledWith('base-1');
		});
	});

	describe('empty state', () => {
		it('should show empty message when no webhooks', async () => {
			mockWebhooksList.mockResolvedValueOnce({
				webhooks: []
			});

			render(WebhooksPanel, {
				props: {
					baseId: 'base-1'
				}
			});

			await waitFor(() => {
				expect(screen.getByText('No webhooks yet.')).toBeTruthy();
			});
		});
	});
});
