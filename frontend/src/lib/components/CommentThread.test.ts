import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import CommentThread from './CommentThread.svelte';
import type { User } from '$lib/types';

// Mock the API client
const mockCommentsList = vi.fn();
vi.mock('$lib/api/client', () => ({
	comments: {
		list: (...args: any[]) => mockCommentsList(...args),
		create: vi.fn().mockResolvedValue({}),
		update: vi.fn().mockResolvedValue({}),
		delete: vi.fn().mockResolvedValue({}),
		resolve: vi.fn().mockResolvedValue({})
	}
}));

describe('CommentThread component', () => {
	const mockUser: User = {
		id: 'user-1',
		email: 'alice@example.com',
		name: 'Alice',
		created_at: '2024-01-01',
		updated_at: '2024-01-01'
	};

	beforeEach(() => {
		vi.clearAllMocks();
		mockCommentsList.mockResolvedValue({
			comments: [
				{
					id: 'comment-1',
					record_id: 'rec-1',
					user_id: 'user-1',
					content: 'This is a test comment',
					is_resolved: false,
					created_at: new Date().toISOString(),
					updated_at: new Date().toISOString(),
					user: { id: 'user-1', email: 'alice@example.com', name: 'Alice' },
					replies: []
				}
			]
		});
	});

	describe('rendering', () => {
		it('should show loading state initially', () => {
			render(CommentThread, {
				props: {
					recordId: 'rec-1',
					currentUser: mockUser
				}
			});

			expect(screen.getByText('Loading comments...')).toBeTruthy();
		});

		it('should render comment thread container', () => {
			render(CommentThread, {
				props: {
					recordId: 'rec-1',
					currentUser: mockUser
				}
			});

			const container = document.querySelector('.comment-thread');
			expect(container).toBeTruthy();
		});
	});

	describe('empty state', () => {
		it('should handle loading', () => {
			mockCommentsList.mockImplementation(() => new Promise(() => {})); // Never resolves

			render(CommentThread, {
				props: {
					recordId: 'rec-1',
					currentUser: mockUser
				}
			});

			expect(screen.getByText('Loading comments...')).toBeTruthy();
		});
	});

	describe('API integration', () => {
		it('should render component for API integration', () => {
			render(CommentThread, {
				props: {
					recordId: 'rec-1',
					currentUser: mockUser
				}
			});

			expect(document.querySelector('.comment-thread')).toBeTruthy();
		});
	});
});
