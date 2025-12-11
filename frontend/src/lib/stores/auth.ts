import { writable } from 'svelte/store';
import type { User } from '$lib/types';
import { auth as authApi } from '$lib/api/client';
import { goto } from '$app/navigation';

interface AuthState {
	user: User | null;
	loading: boolean;
	initialized: boolean;
}

function createAuthStore() {
	const { subscribe, set, update } = writable<AuthState>({
		user: null,
		loading: true,
		initialized: false,
	});

	return {
		subscribe,

		async init() {
			const token = localStorage.getItem('token');
			if (!token) {
				set({ user: null, loading: false, initialized: true });
				return;
			}

			try {
				const { user } = await authApi.me();
				set({ user, loading: false, initialized: true });
			} catch (e) {
				localStorage.removeItem('token');
				set({ user: null, loading: false, initialized: true });
			}
		},

		async login(token: string, user: User) {
			localStorage.setItem('token', token);
			set({ user, loading: false, initialized: true });
		},

		async logout() {
			try {
				await authApi.logout();
			} catch (e) {
				// Ignore errors on logout
			}
			localStorage.removeItem('token');
			set({ user: null, loading: false, initialized: true });
			goto('/login');
		},

		updateUser(user: User) {
			update(state => ({ ...state, user }));
		},
	};
}

export const authStore = createAuthStore();
