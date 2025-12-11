import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen, waitFor } from '@testing-library/svelte';
import ShareModal from './ShareModal.svelte';

// Mock the API client
const mockListCollaborators = vi.fn();
const mockAddCollaborator = vi.fn();
const mockRemoveCollaborator = vi.fn();

vi.mock('$lib/api/client', () => ({
	bases: {
		listCollaborators: (...args: any[]) => mockListCollaborators(...args),
		addCollaborator: (...args: any[]) => mockAddCollaborator(...args),
		removeCollaborator: (...args: any[]) => mockRemoveCollaborator(...args)
	}
}));

// Mock confirm
global.confirm = vi.fn(() => true);

describe('ShareModal component', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		mockListCollaborators.mockResolvedValue({
			collaborators: [
				{
					id: 'collab-1',
					base_id: 'base-1',
					user_id: 'user-1',
					role: 'owner',
					created_at: '2024-01-01',
					user: { id: 'user-1', email: 'owner@example.com', name: 'Owner' }
				},
				{
					id: 'collab-2',
					base_id: 'base-1',
					user_id: 'user-2',
					role: 'editor',
					created_at: '2024-01-01',
					user: { id: 'user-2', email: 'editor@example.com', name: 'Editor' }
				}
			]
		});
		mockAddCollaborator.mockResolvedValue({
			id: 'collab-3',
			base_id: 'base-1',
			user_id: 'user-3',
			role: 'viewer',
			created_at: '2024-01-01',
			user: { id: 'user-3', email: 'new@example.com' }
		});
		mockRemoveCollaborator.mockResolvedValue({ message: 'Removed' });
	});

	describe('rendering', () => {
		it('should render share modal title', async () => {
			render(ShareModal, {
				props: {
					baseId: 'base-1',
					isOwner: true
				}
			});

			expect(screen.getByText('Share this base')).toBeTruthy();
		});

		it('should render close button', () => {
			render(ShareModal, {
				props: {
					baseId: 'base-1',
					isOwner: true
				}
			});

			expect(screen.getByText('×')).toBeTruthy();
		});

		it('should show invite form when owner', () => {
			render(ShareModal, {
				props: {
					baseId: 'base-1',
					isOwner: true
				}
			});

			expect(screen.getByPlaceholderText(/email/i)).toBeTruthy();
		});

		it('should show role selector in invite form', () => {
			render(ShareModal, {
				props: {
					baseId: 'base-1',
					isOwner: true
				}
			});

			expect(screen.getByText('Editor')).toBeTruthy();
		});

		it('should show invite button', () => {
			render(ShareModal, {
				props: {
					baseId: 'base-1',
					isOwner: true
				}
			});

			expect(screen.getByText('Invite')).toBeTruthy();
		});

		it('should show people with access section', () => {
			render(ShareModal, {
				props: {
					baseId: 'base-1',
					isOwner: true
				}
			});

			expect(screen.getByText('People with access')).toBeTruthy();
		});
	});

	describe('invite functionality', () => {
		it('should allow typing email address', async () => {
			render(ShareModal, {
				props: {
					baseId: 'base-1',
					isOwner: true
				}
			});

			const emailInput = screen.getByPlaceholderText(/email/i) as HTMLInputElement;
			await fireEvent.input(emailInput, { target: { value: 'new@example.com' } });

			expect(emailInput.value).toBe('new@example.com');
		});
	});

	describe('close functionality', () => {
		it('should dispatch close event when close button clicked', async () => {
			const { component } = render(ShareModal, {
				props: {
					baseId: 'base-1',
					isOwner: true
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
			render(ShareModal, {
				props: {
					baseId: 'base-1',
					isOwner: true
				}
			});

			expect(document.querySelector('.modal-overlay')).toBeTruthy();
		});
	});
});
