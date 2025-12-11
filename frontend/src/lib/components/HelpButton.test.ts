import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import HelpButton from './HelpButton.svelte';

describe('HelpButton component', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('rendering', () => {
		it('should render help link', () => {
			render(HelpButton);

			const link = document.querySelector('a.help-button');
			expect(link).toBeTruthy();
		});

		it('should have appropriate title attribute', () => {
			render(HelpButton);

			const link = document.querySelector('a.help-button');
			expect(link?.getAttribute('title')).toBe('Help & Documentation');
		});

		it('should link to docs page', () => {
			render(HelpButton);

			const link = document.querySelector('a.help-button');
			expect(link?.getAttribute('href')).toBe('/docs');
		});

		it('should open in new tab', () => {
			render(HelpButton);

			const link = document.querySelector('a.help-button');
			expect(link?.getAttribute('target')).toBe('_blank');
		});

		it('should display Help text', () => {
			render(HelpButton);

			expect(screen.getByText('Help')).toBeTruthy();
		});

		it('should render SVG icon', () => {
			render(HelpButton);

			const svg = document.querySelector('svg');
			expect(svg).toBeTruthy();
		});
	});

	describe('size prop', () => {
		it('should apply small class when size is sm', () => {
			render(HelpButton, {
				props: {
					size: 'sm'
				}
			});

			const link = document.querySelector('a.help-button.small');
			expect(link).toBeTruthy();
		});

		it('should not apply small class when size is md', () => {
			render(HelpButton, {
				props: {
					size: 'md'
				}
			});

			const link = document.querySelector('a.help-button');
			expect(link?.classList.contains('small')).toBe(false);
		});
	});
});
