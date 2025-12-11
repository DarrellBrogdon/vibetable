import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';
import { toastStore, type Toast } from './toast';

describe('toastStore', () => {
	beforeEach(() => {
		vi.useFakeTimers();
		// Clear any existing toasts
		const currentToasts = get(toastStore);
		currentToasts.forEach(t => toastStore.remove(t.id));
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	describe('success', () => {
		it('should add a success toast', () => {
			toastStore.success('Operation completed');

			const toasts = get(toastStore);
			expect(toasts).toHaveLength(1);
			expect(toasts[0].type).toBe('success');
			expect(toasts[0].message).toBe('Operation completed');
			expect(toasts[0].id).toBeDefined();
		});

		it('should auto-remove after default duration (3000ms)', () => {
			toastStore.success('Will disappear');

			expect(get(toastStore)).toHaveLength(1);

			vi.advanceTimersByTime(3000);

			expect(get(toastStore)).toHaveLength(0);
		});

		it('should auto-remove after custom duration', () => {
			toastStore.success('Custom duration', 5000);

			expect(get(toastStore)).toHaveLength(1);

			vi.advanceTimersByTime(3000);
			expect(get(toastStore)).toHaveLength(1); // Still there

			vi.advanceTimersByTime(2000);
			expect(get(toastStore)).toHaveLength(0); // Now gone
		});
	});

	describe('error', () => {
		it('should add an error toast', () => {
			toastStore.error('Something went wrong');

			const toasts = get(toastStore);
			expect(toasts).toHaveLength(1);
			expect(toasts[0].type).toBe('error');
			expect(toasts[0].message).toBe('Something went wrong');
		});

		it('should auto-remove after 5000ms by default for errors', () => {
			toastStore.error('Error message');

			expect(get(toastStore)).toHaveLength(1);

			vi.advanceTimersByTime(4999);
			expect(get(toastStore)).toHaveLength(1);

			vi.advanceTimersByTime(1);
			expect(get(toastStore)).toHaveLength(0);
		});

		it('should respect custom duration for errors', () => {
			toastStore.error('Quick error', 1000);

			vi.advanceTimersByTime(1000);
			expect(get(toastStore)).toHaveLength(0);
		});
	});

	describe('info', () => {
		it('should add an info toast', () => {
			toastStore.info('FYI');

			const toasts = get(toastStore);
			expect(toasts).toHaveLength(1);
			expect(toasts[0].type).toBe('info');
			expect(toasts[0].message).toBe('FYI');
		});

		it('should auto-remove after default duration', () => {
			toastStore.info('Information');

			vi.advanceTimersByTime(3000);
			expect(get(toastStore)).toHaveLength(0);
		});
	});

	describe('remove', () => {
		it('should manually remove a toast by id', () => {
			const id = toastStore.success('To be removed', 0); // duration 0 = no auto-remove

			expect(get(toastStore)).toHaveLength(1);

			toastStore.remove(id);

			expect(get(toastStore)).toHaveLength(0);
		});

		it('should only remove the specified toast', () => {
			toastStore.success('First', 0);
			const secondId = toastStore.success('Second', 0);
			toastStore.success('Third', 0);

			expect(get(toastStore)).toHaveLength(3);

			toastStore.remove(secondId);

			const toasts = get(toastStore);
			expect(toasts).toHaveLength(2);
			expect(toasts.map(t => t.message)).toEqual(['First', 'Third']);
		});

		it('should handle removing non-existent toast gracefully', () => {
			toastStore.success('Exists', 0);

			expect(() => toastStore.remove('non-existent-id')).not.toThrow();
			expect(get(toastStore)).toHaveLength(1);
		});
	});

	describe('multiple toasts', () => {
		it('should handle multiple toasts simultaneously', () => {
			toastStore.success('Success message', 0);
			toastStore.error('Error message', 0);
			toastStore.info('Info message', 0);

			const toasts = get(toastStore);
			expect(toasts).toHaveLength(3);
			expect(toasts.map(t => t.type)).toEqual(['success', 'error', 'info']);
		});

		it('should remove toasts in correct order based on their durations', () => {
			toastStore.success('Short', 1000);
			toastStore.error('Medium', 2000);
			toastStore.info('Long', 3000);

			expect(get(toastStore)).toHaveLength(3);

			vi.advanceTimersByTime(1000);
			expect(get(toastStore)).toHaveLength(2);

			vi.advanceTimersByTime(1000);
			expect(get(toastStore)).toHaveLength(1);

			vi.advanceTimersByTime(1000);
			expect(get(toastStore)).toHaveLength(0);
		});
	});

	describe('toast id generation', () => {
		it('should generate unique ids for each toast', () => {
			const ids = new Set<string>();

			for (let i = 0; i < 100; i++) {
				const id = toastStore.success(`Toast ${i}`, 0);
				expect(ids.has(id)).toBe(false);
				ids.add(id);
			}
		});
	});
});
