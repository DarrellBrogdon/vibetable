import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen, waitFor } from '@testing-library/svelte';
import APIKeysPanel from './APIKeysPanel.svelte';
import type { APIKey } from '$lib/types';

// Mock the API client
vi.mock('$lib/api/client', () => ({
	apiKeys: {
		list: vi.fn().mockResolvedValue({
			api_keys: [
				{
					id: 'key-1',
					user_id: 'user-1',
					name: 'Production Key',
					key_prefix: 'vt_prod_abc',
					last_used_at: '2024-01-01T00:00:00Z',
					created_at: '2024-01-01T00:00:00Z'
				},
				{
					id: 'key-2',
					user_id: 'user-1',
					name: 'Development Key',
					key_prefix: 'vt_dev_xyz',
					created_at: '2024-01-01T00:00:00Z'
				}
			]
		}),
		create: vi.fn().mockResolvedValue({
			id: 'key-3',
			name: 'New Key',
			token: 'vt_new_key_full_token_here'
		}),
		delete: vi.fn().mockResolvedValue({})
	}
}));

// Mock confirm
global.confirm = vi.fn(() => true);

// Mock navigator.clipboard
Object.assign(navigator, {
	clipboard: {
		writeText: vi.fn().mockResolvedValue(undefined)
	}
});

describe('APIKeysPanel component', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('rendering', () => {
		it('should render panel header', async () => {
			render(APIKeysPanel);

			expect(screen.getByText('API Keys')).toBeTruthy();
		});

		it('should show new API key button', () => {
			render(APIKeysPanel);

			expect(screen.getByText('+ New API Key')).toBeTruthy();
		});

		it('should load and display API keys', async () => {
			render(APIKeysPanel);

			await waitFor(() => {
				expect(screen.getByText('Production Key')).toBeTruthy();
				expect(screen.getByText('Development Key')).toBeTruthy();
			});
		});

		it('should display key prefixes', async () => {
			render(APIKeysPanel);

			await waitFor(() => {
				expect(screen.getByText('vt_prod_abc...')).toBeTruthy();
				expect(screen.getByText('vt_dev_xyz...')).toBeTruthy();
			});
		});

		it('should display creation dates', async () => {
			render(APIKeysPanel);

			await waitFor(() => {
				const dateElements = screen.getAllByText(/Created/);
				expect(dateElements.length).toBeGreaterThan(0);
			});
		});

		it('should display last used date when available', async () => {
			render(APIKeysPanel);

			await waitFor(() => {
				expect(screen.getByText(/Last used/)).toBeTruthy();
			});
		});
	});

	describe('create form', () => {
		it('should toggle create form when button clicked', async () => {
			render(APIKeysPanel);

			const newButton = screen.getByText('+ New API Key');
			await fireEvent.click(newButton);

			expect(screen.getByText('Cancel')).toBeTruthy();
		});

		it('should show name input when creating', async () => {
			render(APIKeysPanel);

			const newButton = screen.getByText('+ New API Key');
			await fireEvent.click(newButton);

			expect(screen.getByLabelText('Key Name')).toBeTruthy();
		});

		it('should show create button', async () => {
			render(APIKeysPanel);

			const newButton = screen.getByText('+ New API Key');
			await fireEvent.click(newButton);

			expect(screen.getByText('Create API Key')).toBeTruthy();
		});
	});

	describe('key actions', () => {
		it('should have delete buttons', async () => {
			render(APIKeysPanel);

			await waitFor(() => {
				const deleteButtons = document.querySelectorAll('.btn-icon');
				expect(deleteButtons.length).toBeGreaterThan(0);
			});
		});
	});

	describe('empty state', () => {
		it('should show empty message when no keys', async () => {
			vi.mocked(await import('$lib/api/client')).apiKeys.list.mockResolvedValueOnce({
				api_keys: []
			});

			render(APIKeysPanel);

			await waitFor(() => {
				expect(screen.getByText('No API keys yet.')).toBeTruthy();
			});
		});

		it('should show hint in empty state', async () => {
			vi.mocked(await import('$lib/api/client')).apiKeys.list.mockResolvedValueOnce({
				api_keys: []
			});

			render(APIKeysPanel);

			await waitFor(() => {
				expect(screen.getByText(/Create an API key to access/)).toBeTruthy();
			});
		});
	});
});
