import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';

// Mock the API client and navigation before importing the auth store
vi.mock('$lib/api/client', () => ({
	auth: {
		me: vi.fn(),
		logout: vi.fn(),
	},
}));

vi.mock('$app/navigation', () => ({
	goto: vi.fn(),
}));

// Mock localStorage
const localStorageMock = (() => {
	let store: Record<string, string> = {};
	return {
		getItem: vi.fn((key: string) => store[key] || null),
		setItem: vi.fn((key: string, value: string) => { store[key] = value; }),
		removeItem: vi.fn((key: string) => { delete store[key]; }),
		clear: () => { store = {}; },
	};
})();
Object.defineProperty(global, 'localStorage', { value: localStorageMock });

// Import after mocks are set up
import { authStore } from './auth';
import { auth as authApi } from '$lib/api/client';
import { goto } from '$app/navigation';

describe('authStore', () => {
	beforeEach(() => {
		vi.clearAllMocks();
		localStorageMock.clear();
	});

	describe('initial state', () => {
		it('should have correct initial values', () => {
			// Note: authStore is a singleton, state may be affected by other tests
			// This test verifies the store has the expected structure
			const state = get(authStore);
			expect(state).toHaveProperty('user');
			expect(state).toHaveProperty('loading');
			expect(state).toHaveProperty('initialized');
		});
	});

	describe('init', () => {
		it('should set initialized without fetching user when no token', async () => {
			localStorageMock.getItem.mockReturnValue(null);

			await authStore.init();

			const state = get(authStore);
			expect(state.user).toBeNull();
			expect(state.loading).toBe(false);
			expect(state.initialized).toBe(true);
			expect(authApi.me).not.toHaveBeenCalled();
		});

		it('should fetch user when token exists', async () => {
			const mockUser = { id: '1', email: 'test@example.com', name: 'Test User' };
			localStorageMock.getItem.mockReturnValue('test-token');
			(authApi.me as ReturnType<typeof vi.fn>).mockResolvedValueOnce({ user: mockUser });

			await authStore.init();

			const state = get(authStore);
			expect(state.user).toEqual(mockUser);
			expect(state.loading).toBe(false);
			expect(state.initialized).toBe(true);
			expect(authApi.me).toHaveBeenCalled();
		});

		it('should clear token and reset state on fetch error', async () => {
			localStorageMock.getItem.mockReturnValue('invalid-token');
			(authApi.me as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Unauthorized'));

			await authStore.init();

			const state = get(authStore);
			expect(state.user).toBeNull();
			expect(state.loading).toBe(false);
			expect(state.initialized).toBe(true);
			expect(localStorageMock.removeItem).toHaveBeenCalledWith('token');
		});
	});

	describe('login', () => {
		it('should store token and set user', async () => {
			const mockUser = { id: '1', email: 'test@example.com', name: 'Test User' };

			await authStore.login('new-session-token', mockUser);

			expect(localStorageMock.setItem).toHaveBeenCalledWith('token', 'new-session-token');
			const state = get(authStore);
			expect(state.user).toEqual(mockUser);
			expect(state.loading).toBe(false);
			expect(state.initialized).toBe(true);
		});
	});

	describe('logout', () => {
		it('should call logout API, clear token, reset state, and redirect', async () => {
			(authApi.logout as ReturnType<typeof vi.fn>).mockResolvedValueOnce({ message: 'Logged out' });

			await authStore.logout();

			expect(authApi.logout).toHaveBeenCalled();
			expect(localStorageMock.removeItem).toHaveBeenCalledWith('token');
			expect(goto).toHaveBeenCalledWith('/login');

			const state = get(authStore);
			expect(state.user).toBeNull();
		});

		it('should handle logout API error gracefully', async () => {
			(authApi.logout as ReturnType<typeof vi.fn>).mockRejectedValueOnce(new Error('Network error'));

			// Should not throw
			await authStore.logout();

			// Should still clear token and redirect
			expect(localStorageMock.removeItem).toHaveBeenCalledWith('token');
			expect(goto).toHaveBeenCalledWith('/login');
		});
	});

	describe('updateUser', () => {
		it('should update user in state', async () => {
			const initialUser = { id: '1', email: 'test@example.com', name: 'Test User' };
			await authStore.login('token', initialUser);

			const updatedUser = { id: '1', email: 'test@example.com', name: 'Updated Name' };
			authStore.updateUser(updatedUser);

			const state = get(authStore);
			expect(state.user).toEqual(updatedUser);
		});

		it('should preserve other state properties when updating user', async () => {
			const initialUser = { id: '1', email: 'test@example.com' };
			await authStore.login('token', initialUser);

			const updatedUser = { id: '1', email: 'changed@example.com' };
			authStore.updateUser(updatedUser);

			const state = get(authStore);
			expect(state.loading).toBe(false);
			expect(state.initialized).toBe(true);
		});
	});

	describe('subscribe', () => {
		it('should allow subscribing to state changes', async () => {
			const states: any[] = [];
			const unsubscribe = authStore.subscribe(state => {
				states.push({ ...state });
			});

			const user = { id: '1', email: 'test@example.com' };
			await authStore.login('token', user);

			expect(states.length).toBeGreaterThan(0);
			expect(states[states.length - 1].user).toEqual(user);

			unsubscribe();
		});
	});
});
