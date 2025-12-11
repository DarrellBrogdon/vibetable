import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render } from '@testing-library/svelte';

// Create the mock store entirely inside vi.mock to avoid hoisting issues
// We'll export a reference so we can manipulate it in tests
let mockStoreValue = {
	connected: true,
	connecting: false,
	presence: new Map<string, any>([
		['user-1', {
			userId: 'user-1',
			email: 'alice@example.com',
			name: 'Alice Smith',
			joinedAt: '2024-01-01T00:00:00Z',
			updatedAt: '2024-01-01T00:00:00Z'
		}],
		['user-2', {
			userId: 'user-2',
			email: 'bob@example.com',
			name: 'Bob',
			joinedAt: '2024-01-01T00:00:00Z',
			updatedAt: '2024-01-01T00:00:00Z'
		}]
	])
};

const subscribers = new Set<(v: typeof mockStoreValue) => void>();

vi.mock('$lib/stores/realtime', () => ({
	realtime: {
		subscribe(fn: (v: any) => void) {
			subscribers.add(fn);
			fn(mockStoreValue);
			return () => subscribers.delete(fn);
		}
	}
}));

function setMockStoreValue(value: typeof mockStoreValue) {
	mockStoreValue = value;
	subscribers.forEach(fn => fn(mockStoreValue));
}

import PresenceIndicator from './PresenceIndicator.svelte';

describe('PresenceIndicator component', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		setMockStoreValue({
			connected: true,
			connecting: false,
			presence: new Map([
				['user-1', {
					userId: 'user-1',
					email: 'alice@example.com',
					name: 'Alice Smith',
					joinedAt: '2024-01-01T00:00:00Z',
					updatedAt: '2024-01-01T00:00:00Z'
				}],
				['user-2', {
					userId: 'user-2',
					email: 'bob@example.com',
					name: 'Bob',
					joinedAt: '2024-01-01T00:00:00Z',
					updatedAt: '2024-01-01T00:00:00Z'
				}]
			])
		});
	});

	describe('rendering', () => {
		it('should render presence indicator container', () => {
			render(PresenceIndicator);

			const container = document.querySelector('.presence-indicator');
			expect(container).toBeTruthy();
		});

		it('should show avatars when connected', () => {
			render(PresenceIndicator);

			const avatars = document.querySelectorAll('.avatar');
			expect(avatars.length).toBeGreaterThan(0);
		});

		it('should show connected status', () => {
			render(PresenceIndicator);

			const statusDot = document.querySelector('.connected .status-dot');
			expect(statusDot).toBeTruthy();
		});
	});

	describe('user initials', () => {
		it('should display user initials', () => {
			render(PresenceIndicator);

			const avatars = document.querySelectorAll('.avatar');
			expect(avatars.length).toBeGreaterThan(0);
			// Initials should be present in avatars
			const hasInitials = Array.from(avatars).some(a => a.textContent && a.textContent.length > 0);
			expect(hasInitials).toBe(true);
		});
	});

	describe('connection states', () => {
		it('should show connecting state', () => {
			setMockStoreValue({
				connected: false,
				connecting: true,
				presence: new Map()
			});

			render(PresenceIndicator);

			const connectingStatus = document.querySelector('.connecting');
			expect(connectingStatus).toBeTruthy();
		});

		it('should show disconnected state', () => {
			setMockStoreValue({
				connected: false,
				connecting: false,
				presence: new Map()
			});

			render(PresenceIndicator);

			const disconnectedStatus = document.querySelector('.disconnected');
			expect(disconnectedStatus).toBeTruthy();
		});
	});

	describe('overflow handling', () => {
		it('should show overflow count when more than 5 users', () => {
			setMockStoreValue({
				connected: true,
				connecting: false,
				presence: new Map([
					['user-1', { userId: 'user-1', email: 'a@test.com', name: 'A', joinedAt: '', updatedAt: '' }],
					['user-2', { userId: 'user-2', email: 'b@test.com', name: 'B', joinedAt: '', updatedAt: '' }],
					['user-3', { userId: 'user-3', email: 'c@test.com', name: 'C', joinedAt: '', updatedAt: '' }],
					['user-4', { userId: 'user-4', email: 'd@test.com', name: 'D', joinedAt: '', updatedAt: '' }],
					['user-5', { userId: 'user-5', email: 'e@test.com', name: 'E', joinedAt: '', updatedAt: '' }],
					['user-6', { userId: 'user-6', email: 'f@test.com', name: 'F', joinedAt: '', updatedAt: '' }],
					['user-7', { userId: 'user-7', email: 'g@test.com', name: 'G', joinedAt: '', updatedAt: '' }]
				])
			});

			render(PresenceIndicator);

			const overflowAvatar = document.querySelector('.avatar.overflow');
			expect(overflowAvatar).toBeTruthy();
			expect(overflowAvatar?.textContent).toContain('+');
		});
	});
});
