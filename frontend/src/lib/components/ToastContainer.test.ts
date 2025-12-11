import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup, waitFor } from '@testing-library/svelte';
import { get } from 'svelte/store';
import ToastContainer from './ToastContainer.svelte';
import { toastStore } from '$lib/stores/toast';
import { tick } from 'svelte';

describe('ToastContainer', () => {
	beforeEach(() => {
		// Clear any existing toasts
		const currentToasts = get(toastStore);
		currentToasts.forEach(t => toastStore.remove(t.id));
	});

	afterEach(() => {
		cleanup();
		// Clean up toasts after each test
		const currentToasts = get(toastStore);
		currentToasts.forEach(t => toastStore.remove(t.id));
	});

	it('should render without toasts', () => {
		const { container } = render(ToastContainer);
		expect(container.querySelector('.toast-container')).toBeTruthy();
		expect(container.querySelectorAll('.toast')).toHaveLength(0);
	});

	it('should render a success toast', async () => {
		render(ToastContainer);

		// Add toast with 0 duration (no auto-remove)
		toastStore.success('Success message', 0);

		await waitFor(() => {
			const toast = document.querySelector('.toast-success');
			expect(toast).toBeTruthy();
		});

		const toast = document.querySelector('.toast-success');
		expect(toast?.textContent).toContain('Success message');
	});

	it('should render an error toast', async () => {
		render(ToastContainer);

		toastStore.error('Error message', 0);

		await waitFor(() => {
			const toast = document.querySelector('.toast-error');
			expect(toast).toBeTruthy();
		});

		expect(document.querySelector('.toast-error')?.textContent).toContain('Error message');
	});

	it('should render an info toast', async () => {
		render(ToastContainer);

		toastStore.info('Info message', 0);

		await waitFor(() => {
			expect(document.querySelector('.toast-info')).toBeTruthy();
		});

		expect(document.querySelector('.toast-info')?.textContent).toContain('Info message');
	});

	it('should render multiple toasts', async () => {
		render(ToastContainer);

		toastStore.success('First', 0);
		toastStore.error('Second', 0);
		toastStore.info('Third', 0);

		await waitFor(() => {
			expect(document.querySelectorAll('.toast')).toHaveLength(3);
		});
	});

	it('should remove toast when close button is clicked', async () => {
		render(ToastContainer);

		toastStore.success('To be removed', 0);

		await waitFor(() => {
			expect(document.querySelectorAll('.toast')).toHaveLength(1);
		});

		const closeButton = document.querySelector('.toast-close');
		expect(closeButton).toBeTruthy();

		await fireEvent.click(closeButton!);

		// Toast should be removed from the store
		expect(get(toastStore)).toHaveLength(0);
	});

	it('should display correct icon for success toast', async () => {
		render(ToastContainer);

		toastStore.success('Check icon', 0);

		await waitFor(() => {
			expect(document.querySelector('.toast-success .toast-icon')).toBeTruthy();
		});

		const icon = document.querySelector('.toast-success .toast-icon');
		expect(icon?.textContent?.trim()).toBe('✓');
	});

	it('should display correct icon for error toast', async () => {
		render(ToastContainer);

		toastStore.error('X icon', 0);

		await waitFor(() => {
			expect(document.querySelector('.toast-error .toast-icon')).toBeTruthy();
		});

		const icon = document.querySelector('.toast-error .toast-icon');
		expect(icon?.textContent?.trim()).toBe('✗');
	});

	it('should display correct icon for info toast', async () => {
		render(ToastContainer);

		toastStore.info('Info icon', 0);

		await waitFor(() => {
			expect(document.querySelector('.toast-info .toast-icon')).toBeTruthy();
		});

		const icon = document.querySelector('.toast-info .toast-icon');
		expect(icon?.textContent?.trim()).toBe('ⓘ');
	});
});
