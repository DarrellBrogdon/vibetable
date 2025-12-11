import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/svelte';
import ViewConfigPanel from './ViewConfigPanel.svelte';
import type { Field, ViewConfig, ViewType } from '$lib/types';

describe('ViewConfigPanel component', () => {
	const mockFields: Field[] = [
		{ id: 'field-1', table_id: 'table-1', name: 'Title', field_type: 'text', options: {}, position: 0, created_at: '', updated_at: '' },
		{ id: 'field-2', table_id: 'table-1', name: 'Due Date', field_type: 'date', options: {}, position: 1, created_at: '', updated_at: '' },
		{ id: 'field-3', table_id: 'table-1', name: 'Status', field_type: 'single_select', options: {}, position: 2, created_at: '', updated_at: '' },
		{ id: 'field-4', table_id: 'table-1', name: 'Image URL', field_type: 'text', options: {}, position: 3, created_at: '', updated_at: '' }
	];

	const mockConfig: ViewConfig = {
		date_field_id: '',
		title_field_id: '',
		cover_field_id: '',
		group_by_field_id: ''
	};

	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe('calendar view config', () => {
		it('should render calendar config panel', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'calendar' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText('Configure calendar View')).toBeTruthy();
		});

		it('should show date field selector for calendar', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'calendar' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText('Date Field')).toBeTruthy();
		});

		it('should show title field selector for calendar', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'calendar' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText('Title Field')).toBeTruthy();
		});

		it('should list date fields in dropdown', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'calendar' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText('Due Date')).toBeTruthy();
		});
	});

	describe('gallery view config', () => {
		it('should render gallery config panel', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'gallery' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText('Configure gallery View')).toBeTruthy();
		});

		it('should show cover image field selector', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'gallery' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText('Cover Image Field')).toBeTruthy();
		});
	});

	describe('kanban view config', () => {
		it('should render kanban config panel', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'kanban' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText('Configure kanban View')).toBeTruthy();
		});

		it('should show group by field selector', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'kanban' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText('Group By Field')).toBeTruthy();
		});

		it('should list single select fields for kanban grouping', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'kanban' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText('Status')).toBeTruthy();
		});
	});

	describe('grid view config', () => {
		it('should render grid config panel', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'grid' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText('Configure grid View')).toBeTruthy();
		});

		it('should show no additional options message for grid', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'grid' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText(/no additional configuration/i)).toBeTruthy();
		});
	});

	describe('actions', () => {
		it('should have save button', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'calendar' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText('Save')).toBeTruthy();
		});

		it('should have cancel button', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'calendar' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText('Cancel')).toBeTruthy();
		});

		it('should have close button', () => {
			render(ViewConfigPanel, {
				props: {
					viewType: 'calendar' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			expect(screen.getByText('×')).toBeTruthy();
		});

		it('should dispatch save event when save clicked', async () => {
			const { component } = render(ViewConfigPanel, {
				props: {
					viewType: 'calendar' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			const saveHandler = vi.fn();
			component.$on('save', saveHandler);

			const saveButton = screen.getByText('Save');
			await fireEvent.click(saveButton);

			expect(saveHandler).toHaveBeenCalled();
		});

		it('should dispatch close event when cancel clicked', async () => {
			const { component } = render(ViewConfigPanel, {
				props: {
					viewType: 'calendar' as ViewType,
					config: mockConfig,
					fields: mockFields
				}
			});

			const closeHandler = vi.fn();
			component.$on('close', closeHandler);

			const cancelButton = screen.getByText('Cancel');
			await fireEvent.click(cancelButton);

			expect(closeHandler).toHaveBeenCalled();
		});
	});
});
