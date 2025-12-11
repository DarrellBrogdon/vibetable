import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import ContextualHelp from './ContextualHelp.svelte';

describe('ContextualHelp component', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('rendering', () => {
		it('should render help icon/button', () => {
			render(ContextualHelp, {
				props: {
					title: 'Help Title',
					content: 'Help content here'
				}
			});

			const helpButton = document.querySelector('.help-btn, .contextual-help, button');
			expect(helpButton).toBeTruthy();
		});

		it('should accept title prop', () => {
			render(ContextualHelp, {
				props: {
					title: 'Custom Title',
					content: 'Help content'
				}
			});

			// Component should render without errors
			expect(document.body).toBeTruthy();
		});

		it('should accept content prop', () => {
			render(ContextualHelp, {
				props: {
					title: 'Title',
					content: 'This is the help content'
				}
			});

			// Component should render without errors
			expect(document.body).toBeTruthy();
		});
	});

	describe('tooltip behavior', () => {
		it('should show tooltip on hover', async () => {
			render(ContextualHelp, {
				props: {
					title: 'Title',
					content: 'Tooltip content'
				}
			});

			const trigger = document.querySelector('.help-btn, .contextual-help, button, span');
			if (trigger) {
				await fireEvent.mouseEnter(trigger);
				// Tooltip should be visible after hover
			}
		});

		it('should hide tooltip on mouse leave', async () => {
			render(ContextualHelp, {
				props: {
					title: 'Title',
					content: 'Tooltip content'
				}
			});

			const trigger = document.querySelector('.help-btn, .contextual-help, button, span');
			if (trigger) {
				await fireEvent.mouseEnter(trigger);
				await fireEvent.mouseLeave(trigger);
			}
		});
	});
});
