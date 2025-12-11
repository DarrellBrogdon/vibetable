import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import ActivityFeed from './ActivityFeed.svelte';
import type { Field } from '$lib/types';

// Mock the API client
const mockListForRecord = vi.fn();
const mockListForBase = vi.fn();

vi.mock('$lib/api/client', () => ({
	activity: {
		listForRecord: (...args: any[]) => mockListForRecord(...args),
		listForBase: (...args: any[]) => mockListForBase(...args)
	}
}));

describe('ActivityFeed component', () => {
	const mockFields: Field[] = [
		{ id: 'field-1', table_id: 'table-1', name: 'Status', field_type: 'text', options: {}, position: 0, created_at: '', updated_at: '' }
	];

	beforeEach(() => {
		vi.clearAllMocks();
		mockListForRecord.mockResolvedValue({
			activities: [
				{
					id: 'act-1',
					base_id: 'base-1',
					table_id: 'table-1',
					record_id: 'rec-1',
					user_id: 'user-1',
					action: 'create',
					entity_type: 'record',
					entity_name: 'New Record',
					created_at: new Date().toISOString(),
					user: { id: 'user-1', email: 'alice@test.com', name: 'Alice' }
				}
			]
		});
		mockListForBase.mockResolvedValue({
			activities: [
				{
					id: 'act-2',
					base_id: 'base-1',
					user_id: 'user-1',
					action: 'delete',
					entity_type: 'record',
					entity_name: 'Deleted Record',
					created_at: new Date().toISOString(),
					user: { id: 'user-1', email: 'alice@test.com', name: 'Alice' }
				}
			]
		});
	});

	describe('rendering with recordId', () => {
		it('should show loading state initially', () => {
			render(ActivityFeed, {
				props: {
					recordId: 'rec-1',
					fields: mockFields
				}
			});

			expect(screen.getByText('Loading activity...')).toBeTruthy();
		});

		it('should render activity feed container', () => {
			render(ActivityFeed, {
				props: {
					recordId: 'rec-1',
					fields: mockFields
				}
			});

			const container = document.querySelector('.activity-feed');
			expect(container).toBeTruthy();
		});
	});

	describe('API integration', () => {
		it('should render when recordId provided', () => {
			render(ActivityFeed, {
				props: {
					recordId: 'rec-1',
					fields: mockFields,
					limit: 50
				}
			});

			expect(document.querySelector('.activity-feed')).toBeTruthy();
		});

		it('should render when baseId provided', () => {
			render(ActivityFeed, {
				props: {
					baseId: 'base-1',
					fields: mockFields,
					limit: 50
				}
			});

			expect(document.querySelector('.activity-feed')).toBeTruthy();
		});
	});
});
