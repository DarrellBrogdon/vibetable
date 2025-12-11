<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { bases as basesApi, tables as tablesApi, fields as fieldsApi, records as recordsApi, views as viewsApi, csv, forms as formsApi } from '$lib/api/client';
	import { authStore } from '$lib/stores/auth';
	import { toastStore } from '$lib/stores/toast';
	import { actionHistory, type Action } from '$lib/stores/actionHistory';
	import { realtime, MessageTypes, type RealtimeMessage } from '$lib/stores/realtime';
	import type { Base, Table, Field, Record, View, ViewConfig, ViewFilter, ViewSort, Form } from '$lib/types';
	import Grid from '$lib/components/Grid.svelte';
	import Kanban from '$lib/components/Kanban.svelte';
	import Calendar from '$lib/components/Calendar.svelte';
	import Gallery from '$lib/components/Gallery.svelte';
	import ShareModal from '$lib/components/ShareModal.svelte';
	import ImportModal from '$lib/components/ImportModal.svelte';
	import FormBuilder from '$lib/components/FormBuilder.svelte';
	import ViewConfigPanel from '$lib/components/ViewConfigPanel.svelte';
	import ActivityFeed from '$lib/components/ActivityFeed.svelte';
	import PresenceIndicator from '$lib/components/PresenceIndicator.svelte';
	import WebhooksPanel from '$lib/components/WebhooksPanel.svelte';
	import AutomationPanel from '$lib/components/AutomationPanel.svelte';
	import HelpButton from '$lib/components/HelpButton.svelte';

	let base: Base | null = null;
	let tables: Table[] = [];
	let activeTable: Table | null = null;
	let fields: Field[] = [];
	let records: Record[] = [];

	let loading = true;
	let loadingTable = false;

	let showNewTableModal = false;
	let newTableName = '';
	let creatingTable = false;

	let showShareModal = false;
	let showImportModal = false;

	// Table context menu
	let tableContextMenu: { tableId: string; x: number; y: number } | null = null;
	let editingTableId: string | null = null;
	let editingTableName = '';

	// Views
	let tableViews: View[] = [];
	let activeView: View | null = null;
	let showNewViewMenu = false;
	let newViewName = '';
	let newViewType: 'grid' | 'kanban' | 'calendar' | 'gallery' = 'grid';

	// Forms
	let tableForms: Form[] = [];
	let showFormsMenu = false;
	let selectedForm: Form | null = null;
	let creatingForm = false;

	// View sharing
	let shareViewPopover: string | null = null;
	let togglingViewPublic = false;

	// View configuration
	let configuringView: View | null = null;

	// Activity panel
	let showActivityPanel = false;

	// Webhooks panel (base-level)
	let showWebhooksPanel = false;

	// Automations panel (table-level)
	let showAutomationsPanel = false;

	// Computed: current view type for display
	$: currentViewType = activeView?.type || 'grid';

	// Computed: current view config
	$: currentFilters = (activeView?.config?.filters || []) as ViewFilter[];
	$: currentSort = (activeView?.config?.sorts?.[0] || null) as ViewSort | null;

	$: baseId = $page.params.baseId;

	// Realtime message handler cleanup
	let unsubscribeRealtime: (() => void) | null = null;

	onMount(async () => {
		await loadBase();
		document.addEventListener('keydown', handleGlobalKeydown);
		document.addEventListener('click', handleGlobalClick);

		// Connect to realtime hub
		realtime.connect(baseId);

		// Subscribe to realtime updates
		unsubscribeRealtime = realtime.onMessage(handleRealtimeMessage);
	});

	onDestroy(() => {
		document.removeEventListener('keydown', handleGlobalKeydown);
		document.removeEventListener('click', handleGlobalClick);
		// Clear history for this table when leaving
		if (activeTable) {
			actionHistory.clearForTable(activeTable.id);
		}
		// Disconnect from realtime hub
		realtime.disconnect();
		if (unsubscribeRealtime) {
			unsubscribeRealtime();
		}
	});

	function handleRealtimeMessage(message: RealtimeMessage) {
		// Ignore messages from current user (we already have the data)
		if (message.userId === $authStore.user?.id) return;

		// Only handle messages for the current table
		if (message.tableId && message.tableId !== activeTable?.id) return;

		switch (message.type) {
			case MessageTypes.RECORD_CREATED:
				if (message.payload && activeTable) {
					records = [...records, message.payload as Record];
					toastStore.info('A new record was added');
				}
				break;

			case MessageTypes.RECORD_UPDATED:
				if (message.payload && message.recordId) {
					records = records.map(r =>
						r.id === message.recordId ? (message.payload as Record) : r
					);
				}
				break;

			case MessageTypes.RECORD_DELETED:
				if (message.recordId) {
					records = records.filter(r => r.id !== message.recordId);
					toastStore.info('A record was deleted');
				}
				break;

			case MessageTypes.FIELD_CREATED:
				if (message.payload && activeTable) {
					fields = [...fields, message.payload as Field];
					toastStore.info('A new field was added');
				}
				break;

			case MessageTypes.FIELD_UPDATED:
				if (message.payload && message.fieldId) {
					fields = fields.map(f =>
						f.id === message.fieldId ? (message.payload as Field) : f
					);
				}
				break;

			case MessageTypes.FIELD_DELETED:
				if (message.fieldId) {
					fields = fields.filter(f => f.id !== message.fieldId);
					toastStore.info('A field was deleted');
				}
				break;

			case MessageTypes.TABLE_CREATED:
				if (message.payload) {
					tables = [...tables, message.payload as Table];
					toastStore.info('A new table was created');
				}
				break;

			case MessageTypes.TABLE_UPDATED:
				if (message.payload && message.tableId) {
					tables = tables.map(t =>
						t.id === message.tableId ? (message.payload as Table) : t
					);
					if (activeTable?.id === message.tableId) {
						activeTable = message.payload as Table;
					}
				}
				break;

			case MessageTypes.TABLE_DELETED:
				if (message.tableId) {
					tables = tables.filter(t => t.id !== message.tableId);
					if (activeTable?.id === message.tableId) {
						activeTable = tables[0] || null;
						if (activeTable) {
							selectTable(activeTable);
						}
					}
					toastStore.info('A table was deleted');
				}
				break;
		}
	}

	function handleGlobalClick(e: MouseEvent) {
		// Close forms menu when clicking outside
		if (showFormsMenu) {
			const target = e.target as HTMLElement;
			if (!target.closest('.forms-menu-wrapper')) {
				showFormsMenu = false;
			}
		}
		// Close share popover when clicking outside
		if (shareViewPopover) {
			const target = e.target as HTMLElement;
			if (!target.closest('.share-popover') && !target.closest('.share-view-btn')) {
				shareViewPopover = null;
			}
		}
	}

	function handleGlobalKeydown(e: KeyboardEvent) {
		const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
		const cmdOrCtrl = isMac ? e.metaKey : e.ctrlKey;

		// Don't handle if inside an input or textarea
		if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
			return;
		}

		// Undo: Cmd/Ctrl + Z
		if (cmdOrCtrl && e.key === 'z' && !e.shiftKey) {
			e.preventDefault();
			handleUndo();
			return;
		}

		// Redo: Cmd/Ctrl + Shift + Z or Cmd/Ctrl + Y
		if ((cmdOrCtrl && e.key === 'z' && e.shiftKey) || (cmdOrCtrl && e.key === 'y')) {
			e.preventDefault();
			handleRedo();
			return;
		}
	}

	async function handleUndo() {
		const action = actionHistory.undo();
		if (!action || !activeTable || action.tableId !== activeTable.id) return;

		try {
			switch (action.type) {
				case 'record_update':
					if (action.recordId && action.previousData) {
						await updateRecord(action.recordId, action.previousData, true);
						toastStore.info('Undo: Cell value restored');
					}
					break;

				case 'record_delete':
					// Re-create the deleted record
					if (action.previousData?.values) {
						const newRecord = await recordsApi.create(activeTable.id, action.previousData.values);
						records = [...records, newRecord];
						toastStore.info('Undo: Record restored');
					}
					break;

				case 'record_create':
					// Delete the created record
					if (action.recordId) {
						await recordsApi.delete(action.recordId);
						records = records.filter(r => r.id !== action.recordId);
						toastStore.info('Undo: Record creation reversed');
					}
					break;
			}
		} catch (e) {
			console.error('Undo failed:', e);
			toastStore.error('Undo failed');
		}
	}

	async function handleRedo() {
		const action = actionHistory.redo();
		if (!action || !activeTable || action.tableId !== activeTable.id) return;

		try {
			switch (action.type) {
				case 'record_update':
					if (action.recordId && action.newData) {
						await updateRecord(action.recordId, action.newData, true);
						toastStore.info('Redo: Cell value updated');
					}
					break;

				case 'record_delete':
					// Delete the record again
					if (action.recordId) {
						const record = records.find(r => r.id === action.recordId);
						if (record) {
							await recordsApi.delete(action.recordId);
							records = records.filter(r => r.id !== action.recordId);
							toastStore.info('Redo: Record deleted');
						}
					}
					break;

				case 'record_create':
					// Recreate the record
					if (action.newData?.values !== undefined) {
						const newRecord = await recordsApi.create(activeTable.id, action.newData.values);
						records = [...records, newRecord];
						toastStore.info('Redo: Record created');
					}
					break;
			}
		} catch (e) {
			console.error('Redo failed:', e);
			toastStore.error('Redo failed');
		}
	}

	async function loadBase() {
		try {
			const [baseResult, tablesResult] = await Promise.all([
				basesApi.get(baseId),
				tablesApi.list(baseId)
			]);
			base = baseResult;
			tables = tablesResult.tables;

			// Select first table by default
			if (tables.length > 0) {
				await selectTable(tables[0]);
			}
		} catch (e) {
			console.error('Failed to load base:', e);
		} finally {
			loading = false;
		}
	}

	async function selectTable(table: Table) {
		if (activeTable?.id === table.id) return;

		loadingTable = true;
		activeTable = table;

		try {
			const [fieldsResult, recordsResult, viewsResult, formsResult] = await Promise.all([
				fieldsApi.list(table.id),
				recordsApi.list(table.id),
				viewsApi.list(table.id),
				formsApi.list(table.id)
			]);
			fields = fieldsResult.fields;
			records = recordsResult.records;
			tableViews = viewsResult.views;
			tableForms = formsResult.forms;

			// Select first view or create default
			if (tableViews.length > 0) {
				activeView = tableViews[0];
			} else {
				// Create a default grid view
				try {
					const defaultView = await viewsApi.create(table.id, 'Grid View', 'grid', {});
					tableViews = [defaultView];
					activeView = defaultView;
				} catch (e) {
					console.error('Failed to create default view:', e);
					activeView = null;
				}
			}
		} catch (e) {
			console.error('Failed to load table data:', e);
		} finally {
			loadingTable = false;
		}
	}

	async function createTable() {
		if (!newTableName.trim()) return;

		creatingTable = true;
		try {
			const table = await tablesApi.create(baseId, newTableName.trim());
			tables = [...tables, table];
			showNewTableModal = false;
			newTableName = '';
			await selectTable(table);
			toastStore.success(`Table "${table.name}" created`);
		} catch (e) {
			console.error('Failed to create table:', e);
			toastStore.error('Failed to create table');
		} finally {
			creatingTable = false;
		}
	}

	function showTableContextMenu(e: MouseEvent, tableId: string) {
		e.preventDefault();
		tableContextMenu = { tableId, x: e.clientX, y: e.clientY };
	}

	function closeTableContextMenu() {
		tableContextMenu = null;
	}

	function startEditTable(tableId: string) {
		const table = tables.find(t => t.id === tableId);
		if (table) {
			editingTableId = tableId;
			editingTableName = table.name;
		}
		closeTableContextMenu();
	}

	async function saveTableName() {
		if (!editingTableId || !editingTableName.trim()) {
			editingTableId = null;
			return;
		}

		try {
			const updated = await tablesApi.update(editingTableId, editingTableName.trim());
			tables = tables.map(t => t.id === updated.id ? updated : t);
			if (activeTable?.id === updated.id) {
				activeTable = updated;
			}
			toastStore.success('Table renamed');
		} catch (e) {
			console.error('Failed to rename table:', e);
			toastStore.error('Failed to rename table');
		} finally {
			editingTableId = null;
			editingTableName = '';
		}
	}

	async function deleteTable(tableId: string) {
		if (!confirm('Are you sure you want to delete this table? This action cannot be undone.')) {
			closeTableContextMenu();
			return;
		}

		const tableName = tables.find(t => t.id === tableId)?.name;
		try {
			await tablesApi.delete(tableId);
			tables = tables.filter(t => t.id !== tableId);
			if (activeTable?.id === tableId) {
				// Select another table or clear
				if (tables.length > 0) {
					await selectTable(tables[0]);
				} else {
					activeTable = null;
					fields = [];
					records = [];
					tableViews = [];
					activeView = null;
				}
			}
			toastStore.success(`Table "${tableName}" deleted`);
		} catch (e) {
			console.error('Failed to delete table:', e);
			toastStore.error('Failed to delete table');
		} finally {
			closeTableContextMenu();
		}
	}

	async function duplicateTable(tableId: string, includeRecords: boolean = true) {
		closeTableContextMenu();
		try {
			const newTable = await tablesApi.duplicate(tableId, includeRecords);
			tables = [...tables, newTable];
			await selectTable(newTable);
			toastStore.success(`Table duplicated as "${newTable.name}"`);
		} catch (e) {
			console.error('Failed to duplicate table:', e);
			toastStore.error('Failed to duplicate table');
		}
	}

	function exportTable() {
		if (!activeTable) return;
		const url = csv.exportUrl(activeTable.id);
		window.open(url, '_blank');
	}

	function handleImportClose(e: CustomEvent<{ imported: number }>) {
		showImportModal = false;
		if (e.detail.imported > 0 && activeTable) {
			// Reload records
			recordsApi.list(activeTable.id).then(result => {
				records = result.records;
			});
		}
	}

	async function addField(name: string, fieldType: string, options?: any) {
		if (!activeTable) return;
		try {
			const field = await fieldsApi.create(activeTable.id, name, fieldType, options);
			fields = [...fields, field];
			toastStore.success(`Field "${field.name}" created`);
		} catch (e) {
			console.error('Failed to create field:', e);
			toastStore.error('Failed to create field');
		}
	}

	async function updateField(fieldId: string, name: string) {
		try {
			const updated = await fieldsApi.update(fieldId, { name });
			fields = fields.map(f => f.id === fieldId ? updated : f);
			toastStore.success('Field renamed');
		} catch (e) {
			console.error('Failed to update field:', e);
			toastStore.error('Failed to rename field');
		}
	}

	async function updateFieldOptions(fieldId: string, options: any) {
		try {
			const updated = await fieldsApi.update(fieldId, { options });
			fields = fields.map(f => f.id === fieldId ? updated : f);
			toastStore.success('Options updated');
		} catch (e) {
			console.error('Failed to update field options:', e);
			toastStore.error('Failed to update options');
		}
	}

	async function deleteField(fieldId: string) {
		const fieldName = fields.find(f => f.id === fieldId)?.name;
		try {
			await fieldsApi.delete(fieldId);
			fields = fields.filter(f => f.id !== fieldId);
			toastStore.success(`Field "${fieldName}" deleted`);
		} catch (e) {
			console.error('Failed to delete field:', e);
			toastStore.error('Failed to delete field');
		}
	}

	async function reorderFields(fieldIds: string[]) {
		if (!activeTable) return;
		try {
			await fieldsApi.reorder(activeTable.id, fieldIds);
			// Reorder local fields array to match
			const fieldMap = new Map(fields.map(f => [f.id, f]));
			fields = fieldIds.map(id => fieldMap.get(id)!).filter(Boolean);
			toastStore.success('Fields reordered');
		} catch (e) {
			console.error('Failed to reorder fields:', e);
			toastStore.error('Failed to reorder fields');
		}
	}

	function reorderRecords(recordIds: string[]) {
		// Reorder local records array to match the new order
		// This is client-side only for now - could add API persistence later
		const recordMap = new Map(records.map(r => [r.id, r]));
		const reorderedRecords = recordIds.map(id => recordMap.get(id)!).filter(Boolean);
		// Keep records that weren't in the reorder list at the end
		const remainingRecords = records.filter(r => !recordIds.includes(r.id));
		records = [...reorderedRecords, ...remainingRecords];
	}

	async function addRecord(skipHistory = false) {
		if (!activeTable) return;
		try {
			const record = await recordsApi.create(activeTable.id, {});
			records = [...records, record];

			// Track the action for undo/redo
			if (!skipHistory) {
				actionHistory.push({
					type: 'record_create',
					tableId: activeTable.id,
					timestamp: Date.now(),
					recordId: record.id,
					newData: { values: {} }
				});
			}

			return record;
		} catch (e) {
			console.error('Failed to create record:', e);
		}
	}

	// ID of a record that should be opened in edit mode when grid is displayed
	let editNewRecordId: string | null = null;

	async function addRecordWithValue(fieldId: string, value: any) {
		if (!activeTable) return;
		try {
			const record = await recordsApi.create(activeTable.id, { [fieldId]: value });
			records = [...records, record];

			// Switch to grid view and put the new record in edit mode
			const gridView = tableViews.find(v => v.type === 'grid');
			if (gridView) {
				editNewRecordId = record.id;
				activeView = gridView;
				toastStore.info('New card added - edit it in the grid');
			}
		} catch (e) {
			console.error('Failed to create record:', e);
		}
	}

	async function renameSelectOption(fieldId: string, optionId: string, newName: string) {
		const field = fields.find(f => f.id === fieldId);
		if (!field || !field.options?.options) return;

		// Update the option name in the field options
		const updatedOptions = (field.options.options as { id: string; name: string; color: string }[]).map(opt => {
			if (opt.id === optionId) {
				return { ...opt, name: newName };
			}
			return opt;
		});

		try {
			const updatedField = await fieldsApi.update(fieldId, {
				options: { ...field.options, options: updatedOptions }
			});
			// Update the field in local state
			fields = fields.map(f => f.id === fieldId ? updatedField : f);
			toastStore.success('Column renamed');
		} catch (e) {
			console.error('Failed to rename option:', e);
			toastStore.error('Failed to rename column');
		}
	}

	async function addRecordWithDate(dateStr: string) {
		if (!activeTable || !activeView) return;

		// Get the date field ID from the view config or auto-detect
		const dateFieldId = activeView.config?.date_field_id ||
			fields.find(f => f.field_type === 'date')?.id;

		if (!dateFieldId) {
			toastStore.error('No date field available');
			return;
		}

		try {
			const record = await recordsApi.create(activeTable.id, { [dateFieldId]: dateStr });
			records = [...records, record];
			toastStore.success('Record created');
		} catch (e) {
			console.error('Failed to create record:', e);
			toastStore.error('Failed to create record');
		}
	}

	function handleSelectRecord(recordId: string) {
		// Find the record and show its details
		const record = records.find(r => r.id === recordId);
		if (!record) return;

		// Get a title for the record
		const titleField = fields.find(f => f.field_type === 'text');
		const title = titleField ? record.values[titleField.id] || 'Untitled' : 'Untitled';

		// Switch to grid view and highlight the record
		// For now, find the grid view and switch to it
		const gridView = tableViews.find(v => v.type === 'grid');
		if (gridView && activeView?.id !== gridView.id) {
			activeView = gridView;
			toastStore.info(`Viewing record: ${title}`);
		} else {
			toastStore.info(`Selected: ${title}`);
		}
	}

	async function updateRecord(recordId: string, values: { [key: string]: any }, skipHistory = false) {
		if (!activeTable) return;

		// Get the previous values for undo
		const record = records.find(r => r.id === recordId);
		const previousValues: { [key: string]: any } = {};
		if (record) {
			for (const fieldId of Object.keys(values)) {
				previousValues[fieldId] = record.values[fieldId];
			}
		}

		try {
			const updated = await recordsApi.update(recordId, values);
			records = records.map(r => r.id === recordId ? updated : r);

			// Track the action for undo/redo (unless we're undoing/redoing)
			if (!skipHistory) {
				actionHistory.push({
					type: 'record_update',
					tableId: activeTable.id,
					timestamp: Date.now(),
					recordId,
					previousData: previousValues,
					newData: values
				});
			}
		} catch (e) {
			console.error('Failed to update record:', e);
		}
	}

	async function deleteRecord(recordId: string, skipHistory = false) {
		if (!activeTable) return;

		// Get the record data for undo
		const record = records.find(r => r.id === recordId);
		if (!record) return;

		try {
			await recordsApi.delete(recordId);
			records = records.filter(r => r.id !== recordId);

			// Track the action for undo/redo
			if (!skipHistory) {
				actionHistory.push({
					type: 'record_delete',
					tableId: activeTable.id,
					timestamp: Date.now(),
					recordId,
					previousData: { values: record.values }
				});
			}
		} catch (e) {
			console.error('Failed to delete record:', e);
		}
	}

	async function updateRecordColor(recordId: string, color: string | null) {
		try {
			const updated = await recordsApi.updateColor(recordId, color as any);
			records = records.map(r => r.id === recordId ? updated : r);
		} catch (e) {
			console.error('Failed to update record color:', e);
			toastStore.error('Failed to update color');
		}
	}

	function canEdit() {
		return base?.role === 'owner' || base?.role === 'editor';
	}

	// View management
	async function createView() {
		if (!activeTable || !newViewName.trim()) return;

		try {
			const view = await viewsApi.create(activeTable.id, newViewName.trim(), newViewType, {});
			tableViews = [...tableViews, view];
			activeView = view;
			showNewViewMenu = false;
			newViewName = '';
			newViewType = 'grid';
			toastStore.success(`View "${view.name}" created`);
		} catch (e) {
			console.error('Failed to create view:', e);
			toastStore.error('Failed to create view');
		}
	}

	function selectView(view: View) {
		activeView = view;
	}

	async function handleViewChange(event: CustomEvent<{ filters: ViewFilter[], sorts: ViewSort[] }>) {
		if (!activeView || !canEdit()) return;

		const { filters, sorts } = event.detail;
		const newConfig: ViewConfig = {
			...activeView.config,
			filters,
			sorts
		};

		try {
			const updated = await viewsApi.update(activeView.id, { config: newConfig });
			activeView = updated;
			tableViews = tableViews.map(v => v.id === updated.id ? updated : v);
			// Silent save - no toast needed for view config changes
		} catch (e) {
			console.error('Failed to save view config:', e);
			toastStore.error('Failed to save view settings');
		}
	}

	async function deleteView(viewId: string) {
		if (tableViews.length <= 1) return; // Keep at least one view

		const viewName = tableViews.find(v => v.id === viewId)?.name;
		try {
			await viewsApi.delete(viewId);
			tableViews = tableViews.filter(v => v.id !== viewId);
			if (activeView?.id === viewId) {
				activeView = tableViews[0] || null;
			}
			toastStore.success(`View "${viewName}" deleted`);
		} catch (e) {
			console.error('Failed to delete view:', e);
			toastStore.error('Failed to delete view');
		}
	}

	// Form management
	async function createForm() {
		if (!activeTable) return;

		creatingForm = true;
		try {
			const form = await formsApi.create(activeTable.id, 'New Form');
			tableForms = [...tableForms, form];
			selectedForm = form;
			showFormsMenu = false;
			toastStore.success('Form created');
		} catch (e) {
			console.error('Failed to create form:', e);
			toastStore.error('Failed to create form');
		} finally {
			creatingForm = false;
		}
	}

	function editForm(form: Form) {
		selectedForm = form;
		showFormsMenu = false;
	}

	function handleFormUpdated(event: CustomEvent<Form>) {
		const updated = event.detail;
		tableForms = tableForms.map(f => f.id === updated.id ? updated : f);
		selectedForm = null;
		toastStore.success('Form saved');
	}

	function handleFormDeleted() {
		if (selectedForm) {
			tableForms = tableForms.filter(f => f.id !== selectedForm.id);
		}
		selectedForm = null;
		toastStore.success('Form deleted');
	}

	function copyFormUrl(form: Form) {
		if (!form.public_token) return;
		const url = `${window.location.origin}/f/${form.public_token}`;
		navigator.clipboard.writeText(url);
		toastStore.success('Form URL copied to clipboard');
	}

	// View sharing functions
	async function toggleViewPublic(view: View) {
		if (togglingViewPublic) return;

		togglingViewPublic = true;
		try {
			const updated = await viewsApi.setPublic(view.id, !view.is_public);
			tableViews = tableViews.map(v => v.id === updated.id ? updated : v);
			if (activeView?.id === updated.id) {
				activeView = updated;
			}
			if (updated.is_public) {
				toastStore.success('View is now public. Anyone with the link can view it.');
			} else {
				toastStore.success('View is no longer public.');
			}
		} catch (e: any) {
			console.error('Failed to toggle view public:', e);
			if (e.code === 'forbidden') {
				toastStore.error('Only base owners can share views publicly.');
			} else {
				toastStore.error('Failed to update view sharing.');
			}
		} finally {
			togglingViewPublic = false;
		}
	}

	function copyViewUrl(view: View) {
		if (!view.public_token) return;
		const url = `${window.location.origin}/v/${view.public_token}`;
		navigator.clipboard.writeText(url);
		toastStore.success('View URL copied to clipboard');
	}

	function toggleSharePopover(viewId: string) {
		if (shareViewPopover === viewId) {
			shareViewPopover = null;
		} else {
			shareViewPopover = viewId;
		}
	}

	// View configuration
	function openViewConfig(view: View) {
		configuringView = view;
		shareViewPopover = null; // Close share popover if open
	}

	async function saveViewConfig(e: CustomEvent<{ config: ViewConfig }>) {
		if (!configuringView) return;

		try {
			const mergedConfig = { ...configuringView.config, ...e.detail.config };
			const updated = await viewsApi.update(configuringView.id, { config: mergedConfig });
			tableViews = tableViews.map(v => v.id === updated.id ? updated : v);
			if (activeView?.id === updated.id) {
				activeView = updated;
			}
			toastStore.success('View configuration saved');
			configuringView = null;
		} catch (err) {
			console.error('Failed to save view config:', err);
			toastStore.error('Failed to save view configuration');
		}
	}
</script>

<div class="base-view">
	<header class="header">
		<div class="header-left">
			<a href="/bases" class="back-btn">←</a>
			<span class="logo-icon">📊</span>
			{#if base}
				<h1>{base.name}</h1>
			{/if}
		</div>
		<div class="header-right">
			<PresenceIndicator />
			{#if base?.role === 'owner'}
				<button class="settings-btn" on:click={() => showWebhooksPanel = true}>
					Webhooks
				</button>
				<button class="share-btn" on:click={() => showShareModal = true}>
					Share
				</button>
			{/if}
			<HelpButton />
			<span class="user-email">{$authStore.user?.email}</span>
			<button class="logout-btn" on:click={() => authStore.logout()}>
				Logout
			</button>
		</div>
	</header>

	{#if loading}
		<div class="loading">Loading...</div>
	{:else}
		<div class="content">
			<aside class="sidebar">
				<div class="sidebar-section">
					<div class="sidebar-header">
						<span>Tables</span>
						{#if canEdit()}
							<button class="add-btn" on:click={() => showNewTableModal = true}>+</button>
						{/if}
					</div>
					<nav class="tables-nav">
						{#each tables as table}
							{#if editingTableId === table.id}
								<input
									type="text"
									class="table-name-input"
									bind:value={editingTableName}
									on:blur={saveTableName}
									on:keydown={(e) => {
										if (e.key === 'Enter') saveTableName();
										if (e.key === 'Escape') { editingTableId = null; }
									}}
									autofocus
								/>
							{:else}
								<button
									class="table-item"
									class:active={activeTable?.id === table.id}
									on:click={() => selectTable(table)}
									on:contextmenu={(e) => canEdit() && showTableContextMenu(e, table.id)}
									on:dblclick={() => canEdit() && startEditTable(table.id)}
								>
									{table.name}
								</button>
							{/if}
						{/each}
						{#if tables.length === 0}
							<p class="no-tables">No tables yet</p>
						{/if}
					</nav>
				</div>

				<div class="sidebar-section activity-section">
					<button class="sidebar-header clickable" on:click={() => showActivityPanel = !showActivityPanel}>
						<span>Activity</span>
						<span class="expand-icon">{showActivityPanel ? '▼' : '▶'}</span>
					</button>
					{#if showActivityPanel && base}
						<div class="activity-panel">
							<ActivityFeed baseId={base.id} {fields} limit={20} />
						</div>
					{/if}
				</div>
			</aside>

			<main class="main">
				{#if !activeTable}
					<div class="empty-state">
						<p>Select a table or create a new one</p>
					</div>
				{:else if loadingTable}
					<div class="loading">Loading table...</div>
				{:else}
					<div class="view-header">
						<div class="view-tabs">
							{#each tableViews as view}
								<div class="view-tab-wrapper">
									<button
										class="view-tab"
										class:active={activeView?.id === view.id}
										on:click={() => selectView(view)}
									>
										<span class="view-icon">{view.type === 'grid' ? '⊞' : view.type === 'kanban' ? '▦' : view.type === 'calendar' ? '📅' : '🖼'}</span>
										{view.name}
										{#if view.is_public}
											<span class="public-indicator" title="Publicly shared">🔗</span>
										{/if}
										{#if canEdit() && view.type !== 'grid'}
											<button
												class="config-view-btn"
												on:click|stopPropagation={() => openViewConfig(view)}
												title="Configure view"
											>
												<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
													<circle cx="12" cy="12" r="3"/>
													<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
												</svg>
											</button>
										{/if}
										{#if base?.role === 'owner'}
											<button
												class="share-view-btn"
												on:click|stopPropagation={() => toggleSharePopover(view.id)}
												title="Share view"
											>
												<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
													<path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8"/>
													<polyline points="16 6 12 2 8 6"/>
													<line x1="12" y1="2" x2="12" y2="15"/>
												</svg>
											</button>
										{/if}
										{#if canEdit() && tableViews.length > 1}
											<button
												class="delete-view-btn"
												on:click|stopPropagation={() => deleteView(view.id)}
												title="Delete view"
											>
												×
											</button>
										{/if}
									</button>
									{#if shareViewPopover === view.id}
										<div class="share-popover" on:click|stopPropagation>
											<div class="share-popover-header">
												<h4>Share View</h4>
												<button class="close-popover" on:click={() => shareViewPopover = null}>×</button>
											</div>
											<div class="share-option">
												<div class="share-option-info">
													<span class="share-option-title">Public link</span>
													<span class="share-option-desc">Anyone with the link can view (read-only)</span>
												</div>
												<label class="toggle">
													<input
														type="checkbox"
														checked={view.is_public}
														disabled={togglingViewPublic}
														on:change={() => toggleViewPublic(view)}
													/>
													<span class="toggle-slider"></span>
												</label>
											</div>
											{#if view.is_public && view.public_token}
												<div class="share-link-section">
													<input
														type="text"
														readonly
														value={`${window.location.origin}/v/${view.public_token}`}
														class="share-link-input"
													/>
													<button class="copy-link-btn" on:click={() => copyViewUrl(view)}>
														Copy
													</button>
												</div>
											{/if}
										</div>
									{/if}
								</div>
							{/each}
							{#if canEdit()}
								<div class="add-view-wrapper">
									<button class="add-view-btn" on:click={() => showNewViewMenu = !showNewViewMenu}>
										+ Add view
									</button>
									{#if showNewViewMenu}
										<div class="add-view-menu" on:click|stopPropagation>
											<input
												type="text"
												placeholder="View name"
												bind:value={newViewName}
												autofocus
											/>
											<div class="view-type-options">
												<label>
													<input type="radio" bind:group={newViewType} value="grid" />
													Grid
												</label>
												<label>
													<input type="radio" bind:group={newViewType} value="kanban" />
													Kanban
												</label>
												<label>
													<input type="radio" bind:group={newViewType} value="calendar" />
													Calendar
												</label>
												<label>
													<input type="radio" bind:group={newViewType} value="gallery" />
													Gallery
												</label>
											</div>
											<div class="menu-actions">
												<button class="cancel-btn" on:click={() => showNewViewMenu = false}>Cancel</button>
												<button class="create-btn" on:click={createView} disabled={!newViewName.trim()}>
													Create
												</button>
											</div>
										</div>
									{/if}
								</div>
							{/if}
						</div>
						<div class="view-actions">
							{#if canEdit()}
								<div class="forms-menu-wrapper">
									<button class="action-btn" on:click={() => showFormsMenu = !showFormsMenu} title="Public Forms">
										Forms {tableForms.length > 0 ? `(${tableForms.length})` : ''}
									</button>
									{#if showFormsMenu}
										<div class="forms-dropdown" on:click|stopPropagation>
											{#if tableForms.length === 0}
												<p class="no-forms">No forms yet</p>
											{:else}
												{#each tableForms as form}
													<div class="form-item">
														<span class="form-name">{form.name}</span>
														<span class="form-status" class:active={form.is_active} class:inactive={!form.is_active}>
															{form.is_active ? 'Active' : 'Inactive'}
														</span>
														<button class="form-action-btn" on:click={() => copyFormUrl(form)} title="Copy URL">
															<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																<rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
																<path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>
															</svg>
														</button>
														<button class="form-action-btn" on:click={() => editForm(form)} title="Edit">
															<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																<path d="M17 3a2.83 2.83 0 114 4L7.5 20.5 2 22l1.5-5.5L17 3z"/>
															</svg>
														</button>
													</div>
												{/each}
											{/if}
											<button class="create-form-btn" on:click={createForm} disabled={creatingForm}>
												{creatingForm ? 'Creating...' : '+ Create Form'}
											</button>
										</div>
									{/if}
								</div>
								<button class="action-btn" on:click={() => showAutomationsPanel = true} title="Automations">
									Automations
								</button>
								<button class="action-btn" on:click={() => showImportModal = true} title="Import CSV">
									Import
								</button>
							{/if}
							<button class="action-btn" on:click={exportTable} title="Export CSV">
								Export
							</button>
						</div>
					</div>

					{#if currentViewType === 'grid'}
						<Grid
							{fields}
							{records}
							{tables}
							currentTableId={activeTable?.id || ''}
							readonly={!canEdit()}
							initialFilters={currentFilters}
							initialSort={currentSort}
							{editNewRecordId}
							currentUser={$authStore.user}
							on:addField={(e) => addField(e.detail.name, e.detail.fieldType, e.detail.options)}
							on:updateField={(e) => updateField(e.detail.id, e.detail.name)}
							on:updateFieldOptions={(e) => updateFieldOptions(e.detail.id, e.detail.options)}
							on:deleteField={(e) => deleteField(e.detail.id)}
							on:reorderFields={(e) => reorderFields(e.detail.fieldIds)}
							on:addRecord={() => addRecord()}
							on:updateRecord={(e) => updateRecord(e.detail.id, e.detail.values)}
							on:updateRecordColor={(e) => updateRecordColor(e.detail.id, e.detail.color)}
							on:deleteRecord={(e) => deleteRecord(e.detail.id)}
							on:viewChange={handleViewChange}
							on:editNewRecordHandled={() => editNewRecordId = null}
						/>
					{:else if currentViewType === 'kanban'}
						<Kanban
							{fields}
							{records}
							readonly={!canEdit()}
							on:updateRecord={(e) => updateRecord(e.detail.id, e.detail.values)}
							on:addRecordWithValue={(e) => addRecordWithValue(e.detail.fieldId, e.detail.value)}
							on:renameOption={(e) => renameSelectOption(e.detail.fieldId, e.detail.optionId, e.detail.newName)}
							on:selectRecord={(e) => handleSelectRecord(e.detail.id)}
							on:reorderRecords={(e) => reorderRecords(e.detail.recordIds)}
						/>
					{:else if currentViewType === 'calendar'}
						<Calendar
							{fields}
							{records}
							dateFieldId={activeView?.config?.date_field_id || ''}
							titleFieldId={activeView?.config?.title_field_id || ''}
							readonly={!canEdit()}
							on:addRecord={(e) => addRecordWithDate(e.detail.date)}
							on:selectRecord={(e) => handleSelectRecord(e.detail.id)}
						/>
					{:else if currentViewType === 'gallery'}
						<Gallery
							{fields}
							{records}
							titleFieldId={activeView?.config?.title_field_id || ''}
							coverFieldId={activeView?.config?.cover_field_id || ''}
							readonly={!canEdit()}
							on:addRecord={() => addRecord()}
							on:selectRecord={(e) => handleSelectRecord(e.detail.id)}
						/>
					{/if}
				{/if}
			</main>
		</div>
	{/if}
</div>

{#if showNewTableModal}
	<div class="modal-overlay" on:click={() => showNewTableModal = false}>
		<div class="modal" on:click|stopPropagation>
			<h3>Create new table</h3>
			<form on:submit|preventDefault={createTable}>
				<input
					type="text"
					placeholder="Table name"
					bind:value={newTableName}
					disabled={creatingTable}
					autofocus
				/>
				<div class="modal-actions">
					<button type="button" class="secondary-btn" on:click={() => showNewTableModal = false}>
						Cancel
					</button>
					<button type="submit" class="primary-btn" disabled={creatingTable || !newTableName.trim()}>
						{creatingTable ? 'Creating...' : 'Create'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

{#if showShareModal && base}
	<ShareModal
		baseId={base.id}
		isOwner={base.role === 'owner'}
		on:close={() => showShareModal = false}
	/>
{/if}

{#if tableContextMenu}
	<div class="context-menu-overlay" on:click={closeTableContextMenu}>
		<div
			class="context-menu"
			style="left: {tableContextMenu.x}px; top: {tableContextMenu.y}px;"
			on:click|stopPropagation
		>
			<button class="context-menu-item" on:click={() => startEditTable(tableContextMenu.tableId)}>
				Rename
			</button>
			<button class="context-menu-item" on:click={() => duplicateTable(tableContextMenu.tableId, true)}>
				Duplicate
			</button>
			<button class="context-menu-item danger" on:click={() => deleteTable(tableContextMenu.tableId)}>
				Delete
			</button>
		</div>
	</div>
{/if}

{#if showImportModal && activeTable}
	<ImportModal
		tableId={activeTable.id}
		{fields}
		on:close={handleImportClose}
	/>
{/if}

{#if selectedForm && activeTable}
	<FormBuilder
		form={selectedForm}
		tableId={activeTable.id}
		on:close={() => selectedForm = null}
		on:updated={handleFormUpdated}
		on:deleted={handleFormDeleted}
	/>
{/if}

{#if configuringView}
	<ViewConfigPanel
		viewType={configuringView.type}
		config={configuringView.config || {}}
		{fields}
		on:save={saveViewConfig}
		on:close={() => configuringView = null}
	/>
{/if}

{#if showWebhooksPanel && base}
	<div class="panel-modal-overlay" on:click={() => showWebhooksPanel = false}>
		<div class="panel-modal" on:click|stopPropagation>
			<div class="panel-modal-header">
				<h3>Webhooks</h3>
				<button class="close-btn" on:click={() => showWebhooksPanel = false}>&times;</button>
			</div>
			<WebhooksPanel baseId={base.id} />
		</div>
	</div>
{/if}

{#if showAutomationsPanel && activeTable}
	<div class="panel-modal-overlay" on:click={() => showAutomationsPanel = false}>
		<div class="panel-modal panel-modal-lg" on:click|stopPropagation>
			<div class="panel-modal-header">
				<h3>Automations - {activeTable.name}</h3>
				<button class="close-btn" on:click={() => showAutomationsPanel = false}>&times;</button>
			</div>
			<AutomationPanel tableId={activeTable.id} />
		</div>
	</div>
{/if}

<style>
	.base-view {
		height: 100vh;
		display: flex;
		flex-direction: column;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-sm) var(--spacing-lg);
		background: white;
		border-bottom: 1px solid var(--color-border);
		flex-shrink: 0;
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
	}

	.back-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		background: var(--color-gray-100);
		border-radius: var(--radius-md);
		text-decoration: none;
		color: var(--color-text);
		font-size: 1.2rem;
	}

	.back-btn:hover {
		background: var(--color-gray-200);
	}

	.logo-icon {
		font-size: 1.25rem;
	}

	.header h1 {
		font-size: var(--font-size-base);
		margin: 0;
	}

	.header-right {
		display: flex;
		align-items: center;
		gap: var(--spacing-md);
	}

	.user-email {
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
	}

	.share-btn {
		padding: var(--spacing-xs) var(--spacing-md);
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-md);
		font-weight: 500;
		cursor: pointer;
		font-size: var(--font-size-sm);
	}

	.share-btn:hover {
		background: var(--color-primary-hover);
	}

	.logout-btn {
		padding: var(--spacing-xs) var(--spacing-sm);
		background: none;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		color: var(--color-text-muted);
		cursor: pointer;
		font-size: var(--font-size-sm);
	}

	.content {
		display: flex;
		flex: 1;
		overflow: hidden;
	}

	.sidebar {
		width: 260px;
		background: white;
		border-right: 1px solid var(--color-border);
		display: flex;
		flex-direction: column;
		flex-shrink: 0;
		overflow: hidden;
	}

	.sidebar-section {
		display: flex;
		flex-direction: column;
	}

	.sidebar-section:first-child {
		flex: 1;
		min-height: 0;
		overflow: hidden;
	}

	.sidebar-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-sm) var(--spacing-md);
		font-weight: 600;
		font-size: var(--font-size-sm);
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		border-bottom: 1px solid var(--color-border);
		background: white;
	}

	.sidebar-header.clickable {
		cursor: pointer;
		border: none;
		width: 100%;
		text-align: left;
		border-top: 1px solid var(--color-border);
	}

	.sidebar-header.clickable:hover {
		background: var(--color-gray-50);
	}

	.expand-icon {
		font-size: 10px;
		color: var(--color-text-muted);
	}

	.activity-section {
		border-top: 1px solid var(--color-border);
	}

	.activity-panel {
		max-height: 300px;
		overflow-y: auto;
		border-top: 1px solid var(--color-border);
	}

	.add-btn {
		width: 24px;
		height: 24px;
		background: var(--color-gray-100);
		border: none;
		border-radius: var(--radius-sm);
		cursor: pointer;
		font-size: 1rem;
		color: var(--color-text-muted);
	}

	.add-btn:hover {
		background: var(--color-gray-200);
		color: var(--color-text);
	}

	.tables-nav {
		flex: 1;
		overflow-y: auto;
		padding: var(--spacing-xs);
	}

	.table-item {
		display: block;
		width: 100%;
		padding: var(--spacing-sm) var(--spacing-md);
		background: none;
		border: none;
		border-radius: var(--radius-md);
		text-align: left;
		cursor: pointer;
		font-size: var(--font-size-sm);
		color: var(--color-text);
	}

	.table-item:hover {
		background: var(--color-gray-100);
	}

	.table-item.active {
		background: var(--color-primary-light);
		color: var(--color-primary);
		font-weight: 500;
	}

	.no-tables {
		padding: var(--spacing-md);
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
		text-align: center;
	}

	.main {
		flex: 1;
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.view-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-xs) var(--spacing-md);
		background: white;
		border-bottom: 1px solid var(--color-border);
		flex-shrink: 0;
	}

	.view-actions {
		display: flex;
		gap: var(--spacing-xs);
	}

	.action-btn {
		padding: var(--spacing-xs) var(--spacing-sm);
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: var(--font-size-sm);
		cursor: pointer;
		color: var(--color-text-muted);
	}

	.action-btn:hover {
		background: var(--color-gray-100);
		color: var(--color-text);
	}

	.view-tabs {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
	}

	.view-tab {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
		padding: var(--spacing-xs) var(--spacing-sm);
		background: transparent;
		border: none;
		border-radius: var(--radius-md);
		font-size: var(--font-size-sm);
		cursor: pointer;
		color: var(--color-text-muted);
		transition: background-color 0.15s, color 0.15s;
		position: relative;
	}

	.view-tab:hover {
		background: var(--color-gray-100);
		color: var(--color-text);
	}

	.view-tab.active {
		background: var(--color-primary-light);
		color: var(--color-primary);
		font-weight: 500;
	}

	.view-icon {
		font-size: 12px;
	}

	.delete-view-btn {
		display: none;
		width: 16px;
		height: 16px;
		padding: 0;
		background: var(--color-gray-200);
		border: none;
		border-radius: 50%;
		font-size: 12px;
		line-height: 1;
		cursor: pointer;
		color: var(--color-text-muted);
	}

	.view-tab:hover .delete-view-btn {
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.delete-view-btn:hover {
		background: var(--color-error);
		color: white;
	}

	.add-view-wrapper {
		position: relative;
	}

	.add-view-btn {
		padding: var(--spacing-xs) var(--spacing-sm);
		background: transparent;
		border: 1px dashed var(--color-border);
		border-radius: var(--radius-md);
		font-size: var(--font-size-sm);
		cursor: pointer;
		color: var(--color-text-muted);
	}

	.add-view-btn:hover {
		border-color: var(--color-primary);
		color: var(--color-primary);
	}

	.add-view-menu {
		position: absolute;
		top: 100%;
		left: 0;
		margin-top: var(--spacing-xs);
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-lg);
		padding: var(--spacing-md);
		width: 240px;
		z-index: 20;
	}

	.add-view-menu input[type="text"] {
		width: 100%;
		padding: var(--spacing-sm);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: var(--font-size-sm);
		margin-bottom: var(--spacing-sm);
	}

	.view-type-options {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--spacing-xs) var(--spacing-md);
		margin-bottom: var(--spacing-md);
		font-size: var(--font-size-sm);
	}

	.view-type-options label {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
		cursor: pointer;
	}

	.menu-actions {
		display: flex;
		gap: var(--spacing-sm);
		justify-content: flex-end;
	}

	.cancel-btn {
		padding: var(--spacing-xs) var(--spacing-sm);
		background: var(--color-gray-100);
		border: none;
		border-radius: var(--radius-sm);
		cursor: pointer;
		font-size: var(--font-size-sm);
	}

	.create-btn {
		padding: var(--spacing-xs) var(--spacing-sm);
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-sm);
		cursor: pointer;
		font-size: var(--font-size-sm);
	}

	.create-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.loading, .empty-state {
		display: flex;
		align-items: center;
		justify-content: center;
		flex: 1;
		color: var(--color-text-muted);
	}

	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 100;
	}

	.modal {
		background: white;
		border-radius: var(--radius-lg);
		padding: var(--spacing-lg);
		width: 100%;
		max-width: 400px;
		box-shadow: var(--shadow-lg);
	}

	.modal h3 {
		margin: 0 0 var(--spacing-md);
	}

	.modal input {
		width: 100%;
		padding: var(--spacing-sm) var(--spacing-md);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: var(--font-size-base);
		margin-bottom: var(--spacing-md);
	}

	.modal input:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: 0 0 0 3px var(--color-primary-light);
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: var(--spacing-sm);
	}

	.primary-btn {
		padding: var(--spacing-sm) var(--spacing-md);
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-md);
		font-weight: 500;
		cursor: pointer;
	}

	.primary-btn:hover:not(:disabled) {
		background: var(--color-primary-hover);
	}

	.primary-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.secondary-btn {
		padding: var(--spacing-sm) var(--spacing-md);
		background: var(--color-gray-100);
		color: var(--color-text);
		border: none;
		border-radius: var(--radius-md);
		font-weight: 500;
		cursor: pointer;
	}

	.secondary-btn:hover {
		background: var(--color-gray-200);
	}

	/* Table name editing */
	.table-name-input {
		width: 100%;
		padding: var(--spacing-sm) var(--spacing-md);
		border: 2px solid var(--color-primary);
		border-radius: var(--radius-md);
		font-size: var(--font-size-sm);
		background: white;
	}

	/* Context menu */
	.context-menu-overlay {
		position: fixed;
		inset: 0;
		z-index: 200;
	}

	.context-menu {
		position: fixed;
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-lg);
		min-width: 120px;
		padding: var(--spacing-xs);
		z-index: 201;
	}

	.context-menu-item {
		display: block;
		width: 100%;
		padding: var(--spacing-xs) var(--spacing-sm);
		background: none;
		border: none;
		border-radius: var(--radius-sm);
		text-align: left;
		font-size: var(--font-size-sm);
		cursor: pointer;
		color: var(--color-text);
	}

	.context-menu-item:hover {
		background: var(--color-gray-100);
	}

	.context-menu-item.danger {
		color: var(--color-error);
	}

	.context-menu-item.danger:hover {
		background: #fee2e2;
	}

	/* Forms menu */
	.forms-menu-wrapper {
		position: relative;
	}

	.forms-dropdown {
		position: absolute;
		top: 100%;
		right: 0;
		margin-top: 4px;
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
		min-width: 280px;
		z-index: 100;
		padding: 8px;
	}

	.no-forms {
		text-align: center;
		color: var(--color-gray-500);
		font-size: var(--font-size-sm);
		padding: 12px 0;
		margin: 0;
	}

	.form-item {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px;
		border-radius: var(--radius-sm);
		margin-bottom: 4px;
	}

	.form-item:hover {
		background: var(--color-gray-100);
	}

	.form-name {
		flex: 1;
		font-size: var(--font-size-sm);
		font-weight: 500;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.form-status {
		font-size: 11px;
		padding: 2px 6px;
		border-radius: 10px;
		flex-shrink: 0;
	}

	.form-status.active {
		background: #dcfce7;
		color: #166534;
	}

	.form-status.inactive {
		background: var(--color-gray-100);
		color: var(--color-gray-500);
	}

	.form-action-btn {
		background: none;
		border: none;
		padding: 4px;
		cursor: pointer;
		color: var(--color-gray-500);
		border-radius: var(--radius-sm);
	}

	.form-action-btn:hover {
		background: var(--color-gray-200);
		color: var(--color-text);
	}

	.create-form-btn {
		width: 100%;
		padding: 8px 12px;
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
		font-weight: 500;
		cursor: pointer;
		margin-top: 4px;
	}

	.create-form-btn:hover:not(:disabled) {
		background: var(--color-primary-hover);
	}

	.create-form-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	/* View sharing styles */
	.view-tab-wrapper {
		position: relative;
	}

	.public-indicator {
		font-size: 10px;
		opacity: 0.7;
	}

	.share-view-btn,
	.config-view-btn {
		display: none;
		width: 20px;
		height: 20px;
		padding: 0;
		background: var(--color-gray-200);
		border: none;
		border-radius: 4px;
		cursor: pointer;
		color: var(--color-text-muted);
		align-items: center;
		justify-content: center;
	}

	.view-tab:hover .share-view-btn,
	.view-tab:hover .config-view-btn {
		display: flex;
	}

	.share-view-btn:hover,
	.config-view-btn:hover {
		background: var(--color-primary-light);
		color: var(--color-primary);
	}

	.share-popover {
		position: absolute;
		top: 100%;
		left: 0;
		margin-top: 8px;
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
		width: 320px;
		z-index: 100;
		padding: 16px;
	}

	.share-popover-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16px;
	}

	.share-popover-header h4 {
		margin: 0;
		font-size: 14px;
		font-weight: 600;
	}

	.close-popover {
		background: none;
		border: none;
		font-size: 18px;
		cursor: pointer;
		color: var(--color-text-muted);
		padding: 0;
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 4px;
	}

	.close-popover:hover {
		background: var(--color-gray-100);
	}

	.share-option {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		padding: 12px 0;
		border-bottom: 1px solid var(--color-border);
	}

	.share-option-info {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.share-option-title {
		font-size: 14px;
		font-weight: 500;
	}

	.share-option-desc {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	/* Toggle switch */
	.toggle {
		position: relative;
		display: inline-block;
		width: 44px;
		height: 24px;
		flex-shrink: 0;
	}

	.toggle input {
		opacity: 0;
		width: 0;
		height: 0;
	}

	.toggle-slider {
		position: absolute;
		cursor: pointer;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background-color: var(--color-gray-300);
		transition: 0.3s;
		border-radius: 24px;
	}

	.toggle-slider:before {
		position: absolute;
		content: "";
		height: 18px;
		width: 18px;
		left: 3px;
		bottom: 3px;
		background-color: white;
		transition: 0.3s;
		border-radius: 50%;
	}

	.toggle input:checked + .toggle-slider {
		background-color: var(--color-primary);
	}

	.toggle input:checked + .toggle-slider:before {
		transform: translateX(20px);
	}

	.toggle input:disabled + .toggle-slider {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.share-link-section {
		display: flex;
		gap: 8px;
		margin-top: 12px;
	}

	.share-link-input {
		flex: 1;
		padding: 8px 12px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: 12px;
		background: var(--color-gray-50);
		color: var(--color-text);
	}

	.copy-link-btn {
		padding: 8px 16px;
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-sm);
		font-size: 12px;
		font-weight: 500;
		cursor: pointer;
		white-space: nowrap;
	}

	.copy-link-btn:hover {
		background: var(--color-primary-hover);
	}

	/* Settings/Webhooks button */
	.settings-btn {
		padding: var(--spacing-xs) var(--spacing-md);
		background: var(--color-gray-100);
		color: var(--color-text);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-weight: 500;
		cursor: pointer;
		font-size: var(--font-size-sm);
	}

	.settings-btn:hover {
		background: var(--color-gray-200);
	}

	/* Panel modal styles */
	.panel-modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 100;
	}

	.panel-modal {
		background: white;
		border-radius: var(--radius-lg);
		width: 100%;
		max-width: 600px;
		max-height: 80vh;
		overflow: hidden;
		display: flex;
		flex-direction: column;
		box-shadow: var(--shadow-lg);
	}

	.panel-modal-lg {
		max-width: 800px;
	}

	.panel-modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-md) var(--spacing-lg);
		border-bottom: 1px solid var(--color-border);
	}

	.panel-modal-header h3 {
		margin: 0;
		font-size: var(--font-size-lg);
	}

	.close-btn {
		background: none;
		border: none;
		font-size: 24px;
		cursor: pointer;
		color: var(--color-text-muted);
		padding: 0;
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: var(--radius-md);
	}

	.close-btn:hover {
		background: var(--color-gray-100);
		color: var(--color-text);
	}
</style>
