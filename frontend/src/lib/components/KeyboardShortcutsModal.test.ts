import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import KeyboardShortcutsModal from './KeyboardShortcutsModal.svelte';

describe('KeyboardShortcutsModal component', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('rendering', () => {
		it('should render the modal with title', () => {
			render(KeyboardShortcutsModal);

			expect(screen.getByText('Keyboard Shortcuts')).toBeTruthy();
		});

		it('should render navigation shortcuts section', () => {
			render(KeyboardShortcutsModal);

			expect(screen.getByText('Navigation')).toBeTruthy();
			expect(screen.getByText('Move between cells')).toBeTruthy();
			expect(screen.getByText('Move to next cell')).toBeTruthy();
		});

		it('should render editing shortcuts section', () => {
			render(KeyboardShortcutsModal);

			expect(screen.getByText('Editing')).toBeTruthy();
			expect(screen.getByText('Copy cell value')).toBeTruthy();
			expect(screen.getByText('Paste cell value')).toBeTruthy();
			expect(screen.getByText('Undo last change')).toBeTruthy();
		});

		it('should render general shortcuts section', () => {
			render(KeyboardShortcutsModal);

			expect(screen.getByText('General')).toBeTruthy();
			expect(screen.getByText('Search / Quick find')).toBeTruthy();
		});

		it('should render close button', () => {
			render(KeyboardShortcutsModal);

			expect(screen.getByText('×')).toBeTruthy();
		});

		it('should render Esc hint in footer', () => {
			render(KeyboardShortcutsModal);

			// The text is split by the kbd element, so check for the footer containing Esc
			const footer = document.querySelector('.modal-footer');
			expect(footer).toBeTruthy();
			expect(footer?.textContent).toContain('Esc');
			expect(footer?.textContent).toContain('close');
		});
	});

	describe('mouse interaction', () => {
		it('should dispatch close event when close button clicked', async () => {
			const { component } = render(KeyboardShortcutsModal);

			const closeHandler = vi.fn();
			component.$on('close', closeHandler);

			const closeButton = screen.getByText('×');
			await fireEvent.click(closeButton);

			expect(closeHandler).toHaveBeenCalled();
		});

		it('should dispatch close event when overlay clicked', async () => {
			const { component } = render(KeyboardShortcutsModal);

			const closeHandler = vi.fn();
			component.$on('close', closeHandler);

			const overlay = document.querySelector('.modal-overlay');
			if (overlay) {
				await fireEvent.click(overlay);
				expect(closeHandler).toHaveBeenCalled();
			}
		});
	});

	describe('platform detection', () => {
		it('should show appropriate modifier key', () => {
			render(KeyboardShortcutsModal);

			// Should show either Ctrl or ⌘ depending on platform
			const text = document.body.textContent || '';
			expect(text.includes('Ctrl') || text.includes('⌘')).toBe(true);
		});
	});
});
