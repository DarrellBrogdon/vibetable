<script lang="ts">
	import { createEventDispatcher, onMount, onDestroy } from 'svelte';
	import type { Field, Record, Table, ViewFilter, ViewSort, RecordColor, User } from '$lib/types';
	import { records as recordsApi, fields as fieldsApi } from '$lib/api/client';
	import RecordPicker from './RecordPicker.svelte';
	import KeyboardShortcutsModal from './KeyboardShortcutsModal.svelte';
	import RecordModal from './RecordModal.svelte';
	import ContextualHelp from './ContextualHelp.svelte';

	// Color options for records
	const recordColors: { value: RecordColor; label: string; bg: string }[] = [
		{ value: 'red', label: 'Red', bg: '#fee2e2' },
		{ value: 'orange', label: 'Orange', bg: '#ffedd5' },
		{ value: 'yellow', label: 'Yellow', bg: '#fef3c7' },
		{ value: 'green', label: 'Green', bg: '#dcfce7' },
		{ value: 'blue', label: 'Blue', bg: '#dbeafe' },
		{ value: 'purple', label: 'Purple', bg: '#f3e8ff' },
		{ value: 'pink', label: 'Pink', bg: '#fce7f3' },
		{ value: 'gray', label: 'Gray', bg: '#f3f4f6' },
	];

	export let fields: Field[] = [];
	export let records: Record[] = [];
	export let readonly: boolean = false;
	export let tables: Table[] = []; // Available tables for linking
	export let currentTableId: string = ''; // Current table ID (to exclude from linking options)
	export let currentUser: User | null = null; // Current user for comments

	// View state props (for persistence)
	export let initialFilters: ViewFilter[] = [];
	export let initialSort: ViewSort | null = null;

	// If set, the grid will automatically enter edit mode for the first editable field of this record
	export let editNewRecordId: string | null = null;

	const dispatch = createEventDispatcher<{
		addField: { name: string; fieldType: string; options?: any };
		updateField: { id: string; name: string };
		updateFieldOptions: { id: string; options: any };
		deleteField: { id: string };
		reorderFields: { fieldIds: string[] };
		addRecord: void;
		updateRecord: { id: string; values: { [key: string]: any } };
		updateRecordColor: { id: string; color: RecordColor | null };
		deleteRecord: { id: string };
		viewChange: { filters: ViewFilter[]; sort: ViewSort | null };
		editNewRecordHandled: void;
	}>();

	let editingCell: { recordId: string; fieldId: string } | null = null;
	let editValue: any = '';

	// Keyboard navigation state
	let selectedCell: { rowIndex: number; colIndex: number } | null = null;
	let gridContainer: HTMLDivElement;
	let showShortcutsModal = false;
	let clipboard: { value: any; fieldType: string } | null = null;

	let showAddFieldMenu = false;
	let newFieldName = '';
	let newFieldType = 'text';
	let newFieldLinkedTableId = '';
	let newSelectOptions: { id: string; name: string; color: string }[] = [];
	let newOptionName = '';
	let newOptionColor = 'gray';

	// Formula/Rollup/Lookup field configuration
	let newFormulaExpression = '';
	let newFormulaResultType = 'text';
	let newRollupLinkedFieldId = '';
	let newRollupFieldId = '';
	let newAggregationFunction = 'COUNT';
	let newLookupLinkedFieldId = '';
	let newLookupFieldId = '';

	const selectColors = ['red', 'orange', 'yellow', 'green', 'blue', 'purple', 'pink', 'gray'];

	// Field context menu state
	let fieldContextMenu: { fieldId: string; x: number; y: number } | null = null;
	let editingFieldId: string | null = null;
	let editingFieldName = '';

	// Field drag-and-drop state
	let draggingFieldId: string | null = null;
	let dragOverFieldId: string | null = null;

	// Record context menu state (for color picking)
	let recordContextMenu: { recordId: string; x: number; y: number } | null = null;

	// Record expansion modal
	let expandedRecord: Record | null = null;

	// Linked records picker state
	let showRecordPicker = false;
	let pickerFieldId = '';
	let pickerRecordId = '';
	let pickerTableId = '';
	let pickerSelectedIds: string[] = [];

	// Select field editing state
	let selectDropdown: { recordId: string; fieldId: string; x: number; y: number } | null = null;

	// Select options editor state
	let editingOptionsFieldId: string | null = null;
	let editingOptions: { id: string; name: string; color: string }[] = [];
	let newEditOptionName = '';
	let newEditOptionColor = 'gray';

	// Computed field (formula/rollup/lookup) editor state
	let editingComputedFieldId: string | null = null;
	let editingFormulaExpression = '';
	let editingFormulaResultType = 'text';
	let editingRollupLinkedFieldId = '';
	let editingRollupFieldId = '';
	let editingAggregationFunction = 'COUNT';
	let editingLookupLinkedFieldId = '';
	let editingLookupFieldId = '';

	// Attachment modal state
	let attachmentModal: { recordId: string; fieldId: string } | null = null;
	let uploadingAttachment = false;
	let attachmentList: any[] = [];
	let loadingAttachments = false;

	// Column resizing state
	let columnWidths: { [fieldId: string]: number } = {};
	let resizingColumn: { fieldId: string; startX: number; startWidth: number } | null = null;
	const MIN_COLUMN_WIDTH = 80;
	const DEFAULT_COLUMN_WIDTH = 150;

	// Search/quick find state
	let searchQuery = '';
	let showSearch = false;
	let searchInputRef: HTMLInputElement;

	// Bulk selection state
	let selectedRecordIds: Set<string> = new Set();
	let showBulkColorMenu = false;

	// Cache for linked record titles: stored as plain object for Svelte reactivity
	// Structure: { [tableId]: { [recordId]: title } }
	let linkedRecordTitles: { [tableId: string]: { [recordId: string]: string } } = {};
	let loadingLinkedRecords = false;
	let linkedCacheVersion = 0; // Increment to force re-render

	// Tables available for linking (exclude current table)
	$: linkableTables = tables.filter(t => t.id !== currentTableId);

	// Get linked record fields
	$: linkedRecordFields = fields.filter(f => f.field_type === 'linked_record' && f.options?.linked_table_id);

	// Load linked record titles when fields or records change
	$: if (linkedRecordFields.length > 0 && records.length > 0) {
		loadLinkedRecordTitles();
	}

	async function loadLinkedRecordTitles() {
		if (loadingLinkedRecords) return;
		loadingLinkedRecords = true;

		try {
			// Find all unique table IDs we need to fetch
			const tableIdsToFetch: string[] = [];
			for (const field of linkedRecordFields) {
				const tableId = field.options?.linked_table_id;
				if (tableId && !linkedRecordTitles[tableId]) {
					tableIdsToFetch.push(tableId);
				}
			}

			if (tableIdsToFetch.length === 0) {
				return;
			}

			// Fetch records for each linked table
			const newTitles: { [tableId: string]: { [recordId: string]: string } } = {};

			await Promise.all(tableIdsToFetch.map(async (tableId) => {
				try {
					const [fieldsResult, recordsResult] = await Promise.all([
						fieldsApi.list(tableId),
						recordsApi.list(tableId)
					]);

					// Find the primary field (first text field, or first field)
					const primaryField = fieldsResult.fields.find((f: Field) => f.field_type === 'text') || fieldsResult.fields[0];

					// Build the cache for this table
					const titleMap: { [recordId: string]: string } = {};
					for (const record of recordsResult.records) {
						const title = primaryField ? (record.values[primaryField.id] || 'Untitled') : 'Untitled';
						titleMap[record.id] = String(title);
					}
					newTitles[tableId] = titleMap;
				} catch (e) {
					console.error(`Failed to load linked records for table ${tableId}:`, e);
					newTitles[tableId] = {}; // Empty object to prevent retrying
				}
			}));

			// Update cache with new data - create new object for reactivity
			linkedRecordTitles = { ...linkedRecordTitles, ...newTitles };
			linkedCacheVersion += 1;
		} finally {
			loadingLinkedRecords = false;
		}
	}

	// Filtering
	type FilterOperator = 'contains' | 'equals' | 'not_empty' | 'empty' | 'gt' | 'lt';
	interface Filter {
		fieldId: string;
		operator: FilterOperator;
		value: string;
	}
	// Initialize from props
	let filters: Filter[] = initialFilters.map(f => ({
		fieldId: f.field_id,
		operator: f.operator as FilterOperator,
		value: f.value
	}));
	let showFilterMenu = false;
	let newFilter: Filter = { fieldId: '', operator: 'contains', value: '' };

	// Sorting
	interface Sort {
		fieldId: string;
		direction: 'asc' | 'desc';
	}
	let sort: Sort | null = initialSort ? {
		fieldId: initialSort.field_id,
		direction: initialSort.direction
	} : null;
	let showSortMenu = false;

	// Emit view config changes
	function emitViewChange() {
		dispatch('viewChange', {
			filters: filters.map(f => ({
				field_id: f.fieldId,
				operator: f.operator,
				value: f.value
			})),
			sorts: sort ? [{
				field_id: sort.fieldId,
				direction: sort.direction
			}] : []
		});
	}

	// Computed filtered and sorted records
	$: filteredRecords = applyFilters(records, filters);
	$: searchedRecords = applySearch(filteredRecords, searchQuery);
	$: sortedRecords = applySort(searchedRecords, sort);

	function applyFilters(recs: Record[], filters: Filter[]): Record[] {
		if (filters.length === 0) return recs;

		return recs.filter(record => {
			return filters.every(filter => {
				const value = record.values[filter.fieldId];
				const field = fields.find(f => f.id === filter.fieldId);
				if (!field) return true;

				switch (filter.operator) {
					case 'contains':
						return String(value || '').toLowerCase().includes(filter.value.toLowerCase());
					case 'equals':
						if (field.field_type === 'number') {
							return Number(value) === Number(filter.value);
						}
						return String(value || '').toLowerCase() === filter.value.toLowerCase();
					case 'not_empty':
						return value !== null && value !== undefined && value !== '';
					case 'empty':
						return value === null || value === undefined || value === '';
					case 'gt':
						return Number(value) > Number(filter.value);
					case 'lt':
						return Number(value) < Number(filter.value);
					default:
						return true;
				}
			});
		});
	}

	function applySearch(recs: Record[], query: string): Record[] {
		if (!query.trim()) return recs;

		const lowerQuery = query.toLowerCase().trim();

		return recs.filter(record => {
			return fields.some(field => {
				const value = record.values[field.id];
				if (value === null || value === undefined) return false;

				switch (field.field_type) {
					case 'text':
					case 'number':
						return String(value).toLowerCase().includes(lowerQuery);
					case 'date':
						const dateStr = value ? new Date(value).toLocaleDateString() : '';
						return dateStr.toLowerCase().includes(lowerQuery);
					case 'checkbox':
						// Allow searching for 'checked', 'true', 'yes' etc.
						if (value === true) {
							return ['checked', 'true', 'yes', '✓'].some(v => v.includes(lowerQuery));
						}
						return false;
					case 'single_select':
						const option = getSelectOptionById(field, value);
						return option ? option.name.toLowerCase().includes(lowerQuery) : false;
					case 'multi_select':
						const selectedIds = value || [];
						return selectedIds.some((id: string) => {
							const opt = getSelectOptionById(field, id);
							return opt ? opt.name.toLowerCase().includes(lowerQuery) : false;
						});
					case 'linked_record':
						const titles = getLinkedRecordTitles(value, field);
						return titles.some(t => t.title.toLowerCase().includes(lowerQuery));
					default:
						return String(value).toLowerCase().includes(lowerQuery);
				}
			});
		});
	}

	function applySort(recs: Record[], sort: Sort | null): Record[] {
		if (!sort) return recs;

		const field = fields.find(f => f.id === sort.fieldId);
		if (!field) return recs;

		return [...recs].sort((a, b) => {
			const aVal = a.values[sort.fieldId];
			const bVal = b.values[sort.fieldId];

			let comparison = 0;
			if (field.field_type === 'number') {
				comparison = (Number(aVal) || 0) - (Number(bVal) || 0);
			} else if (field.field_type === 'checkbox') {
				comparison = (aVal ? 1 : 0) - (bVal ? 1 : 0);
			} else if (field.field_type === 'date') {
				comparison = new Date(aVal || 0).getTime() - new Date(bVal || 0).getTime();
			} else {
				comparison = String(aVal || '').localeCompare(String(bVal || ''));
			}

			return sort.direction === 'desc' ? -comparison : comparison;
		});
	}

	function addFilter() {
		if (!newFilter.fieldId) return;
		filters = [...filters, { ...newFilter }];
		newFilter = { fieldId: fields[0]?.id || '', operator: 'contains', value: '' };
		showFilterMenu = false;
		emitViewChange();
	}

	function removeFilter(index: number) {
		filters = filters.filter((_, i) => i !== index);
		emitViewChange();
	}

	function setSort(fieldId: string, direction: 'asc' | 'desc') {
		sort = { fieldId, direction };
		showSortMenu = false;
		emitViewChange();
	}

	function clearSort() {
		sort = null;
		showSortMenu = false;
		emitViewChange();
	}

	function toggleSearch() {
		showSearch = !showSearch;
		if (showSearch) {
			// Focus the search input after it renders
			setTimeout(() => searchInputRef?.focus(), 0);
		} else {
			searchQuery = '';
		}
	}

	function clearSearch() {
		searchQuery = '';
		searchInputRef?.focus();
	}

	// Bulk selection functions
	$: allSelected = sortedRecords.length > 0 && sortedRecords.every(r => selectedRecordIds.has(r.id));
	$: someSelected = selectedRecordIds.size > 0;

	function toggleRecordSelection(recordId: string) {
		if (selectedRecordIds.has(recordId)) {
			selectedRecordIds.delete(recordId);
		} else {
			selectedRecordIds.add(recordId);
		}
		selectedRecordIds = selectedRecordIds; // Trigger reactivity
	}

	function toggleSelectAll() {
		if (allSelected) {
			// Deselect all
			selectedRecordIds.clear();
		} else {
			// Select all visible records
			sortedRecords.forEach(r => selectedRecordIds.add(r.id));
		}
		selectedRecordIds = selectedRecordIds; // Trigger reactivity
	}

	function clearSelection() {
		selectedRecordIds.clear();
		selectedRecordIds = selectedRecordIds;
		showBulkColorMenu = false;
	}

	function bulkDeleteRecords() {
		if (selectedRecordIds.size === 0) return;
		const count = selectedRecordIds.size;
		if (!confirm(`Are you sure you want to delete ${count} record${count > 1 ? 's' : ''}? This cannot be undone.`)) {
			return;
		}
		// Delete each selected record
		selectedRecordIds.forEach(id => {
			dispatch('deleteRecord', { id });
		});
		clearSelection();
	}

	function bulkSetColor(color: RecordColor | null) {
		if (selectedRecordIds.size === 0) return;
		selectedRecordIds.forEach(id => {
			dispatch('updateRecordColor', { id, color });
		});
		showBulkColorMenu = false;
	}

	const fieldTypes = [
		{ value: 'text', label: 'Text' },
		{ value: 'number', label: 'Number' },
		{ value: 'checkbox', label: 'Checkbox' },
		{ value: 'date', label: 'Date' },
		{ value: 'single_select', label: 'Single Select' },
		{ value: 'multi_select', label: 'Multi Select' },
		{ value: 'linked_record', label: 'Link to Table' },
		{ value: 'formula', label: 'Formula' },
		{ value: 'rollup', label: 'Rollup' },
		{ value: 'lookup', label: 'Lookup' },
		{ value: 'attachment', label: 'Attachment' }
	];

	const filterOperators = [
		{ value: 'contains', label: 'contains' },
		{ value: 'equals', label: 'equals' },
		{ value: 'not_empty', label: 'is not empty' },
		{ value: 'empty', label: 'is empty' },
		{ value: 'gt', label: '>' },
		{ value: 'lt', label: '<' }
	];

	function startEdit(record: Record, field: Field) {
		if (readonly) return;
		editingCell = { recordId: record.id, fieldId: field.id };
		editValue = record.values[field.id] ?? '';
	}

	function cancelEdit() {
		editingCell = null;
		editValue = '';
	}

	function saveEdit() {
		if (!editingCell) return;

		const record = records.find(r => r.id === editingCell!.recordId);
		const field = fields.find(f => f.id === editingCell!.fieldId);

		if (!record || !field) {
			cancelEdit();
			return;
		}

		let processedValue = editValue;
		if (field.field_type === 'number') {
			processedValue = editValue === '' ? null : Number(editValue);
		} else if (field.field_type === 'checkbox') {
			processedValue = Boolean(editValue);
		}

		if (record.values[field.id] !== processedValue) {
			dispatch('updateRecord', {
				id: editingCell.recordId,
				values: { [editingCell.fieldId]: processedValue }
			});
		}

		cancelEdit();
	}

	function toggleCheckbox(record: Record, field: Field) {
		if (readonly) return;
		const currentValue = record.values[field.id] ?? false;
		dispatch('updateRecord', {
			id: record.id,
			values: { [field.id]: !currentValue }
		});
	}

	function addSelectOption() {
		if (!newOptionName.trim()) return;
		newSelectOptions = [...newSelectOptions, {
			id: crypto.randomUUID(),
			name: newOptionName.trim(),
			color: newOptionColor
		}];
		newOptionName = '';
		newOptionColor = 'gray';
	}

	function removeSelectOption(optionId: string) {
		newSelectOptions = newSelectOptions.filter(o => o.id !== optionId);
	}

	function addField() {
		if (!newFieldName.trim()) return;
		if (newFieldType === 'linked_record' && !newFieldLinkedTableId) return;
		if ((newFieldType === 'single_select' || newFieldType === 'multi_select') && newSelectOptions.length === 0) return;
		if (newFieldType === 'formula' && !newFormulaExpression.trim()) return;
		if (newFieldType === 'rollup' && !newRollupLinkedFieldId) return;
		if (newFieldType === 'lookup' && (!newLookupLinkedFieldId || !newLookupFieldId)) return;

		let options: any = undefined;
		if (newFieldType === 'linked_record') {
			options = { linked_table_id: newFieldLinkedTableId };
		} else if (newFieldType === 'single_select' || newFieldType === 'multi_select') {
			options = { options: newSelectOptions };
		} else if (newFieldType === 'formula') {
			options = { expression: newFormulaExpression, result_type: newFormulaResultType };
		} else if (newFieldType === 'rollup') {
			options = {
				rollup_linked_field_id: newRollupLinkedFieldId,
				rollup_field_id: newRollupFieldId || undefined,
				aggregation_function: newAggregationFunction
			};
		} else if (newFieldType === 'lookup') {
			options = {
				lookup_linked_field_id: newLookupLinkedFieldId,
				lookup_field_id: newLookupFieldId
			};
		}

		dispatch('addField', {
			name: newFieldName.trim(),
			fieldType: newFieldType,
			options
		});
		newFieldName = '';
		newFieldType = 'text';
		newFieldLinkedTableId = '';
		newSelectOptions = [];
		newOptionName = '';
		newOptionColor = 'gray';
		newFormulaExpression = '';
		newFormulaResultType = 'text';
		newRollupLinkedFieldId = '';
		newRollupFieldId = '';
		newAggregationFunction = 'COUNT';
		newLookupLinkedFieldId = '';
		newLookupFieldId = '';
		showAddFieldMenu = false;
	}

	// Field context menu functions
	function showFieldContextMenu(e: MouseEvent, fieldId: string) {
		if (readonly) return;
		e.preventDefault();
		fieldContextMenu = { fieldId, x: e.clientX, y: e.clientY };
	}

	function closeFieldContextMenu() {
		fieldContextMenu = null;
	}

	function startEditField(fieldId: string) {
		const field = fields.find(f => f.id === fieldId);
		if (field) {
			editingFieldId = fieldId;
			editingFieldName = field.name;
		}
		closeFieldContextMenu();
	}

	function saveFieldName() {
		if (!editingFieldId || !editingFieldName.trim()) {
			editingFieldId = null;
			return;
		}
		dispatch('updateField', {
			id: editingFieldId,
			name: editingFieldName.trim()
		});
		editingFieldId = null;
		editingFieldName = '';
	}

	function deleteField(fieldId: string) {
		if (!confirm('Are you sure you want to delete this field? All data in this column will be lost.')) {
			closeFieldContextMenu();
			return;
		}
		dispatch('deleteField', { id: fieldId });
		closeFieldContextMenu();
	}

	// Field drag-and-drop functions
	function handleFieldDragStart(e: DragEvent, fieldId: string) {
		if (readonly) return;
		draggingFieldId = fieldId;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
			e.dataTransfer.setData('text/plain', fieldId);
		}
	}

	function handleFieldDragOver(e: DragEvent, fieldId: string) {
		if (readonly || !draggingFieldId || draggingFieldId === fieldId) return;
		e.preventDefault();
		if (e.dataTransfer) {
			e.dataTransfer.dropEffect = 'move';
		}
		dragOverFieldId = fieldId;
	}

	function handleFieldDragLeave() {
		dragOverFieldId = null;
	}

	function handleFieldDrop(e: DragEvent, targetFieldId: string) {
		e.preventDefault();
		if (readonly || !draggingFieldId || draggingFieldId === targetFieldId) {
			draggingFieldId = null;
			dragOverFieldId = null;
			return;
		}

		// Calculate new field order
		const currentOrder = fields.map(f => f.id);
		const sourceIndex = currentOrder.indexOf(draggingFieldId);
		const targetIndex = currentOrder.indexOf(targetFieldId);

		if (sourceIndex === -1 || targetIndex === -1) {
			draggingFieldId = null;
			dragOverFieldId = null;
			return;
		}

		// Remove from current position and insert at new position
		const newOrder = [...currentOrder];
		newOrder.splice(sourceIndex, 1);
		newOrder.splice(targetIndex, 0, draggingFieldId);

		dispatch('reorderFields', { fieldIds: newOrder });

		draggingFieldId = null;
		dragOverFieldId = null;
	}

	function handleFieldDragEnd() {
		draggingFieldId = null;
		dragOverFieldId = null;
	}

	// Column resize functions
	function getColumnWidth(fieldId: string): number {
		return columnWidths[fieldId] || DEFAULT_COLUMN_WIDTH;
	}

	function startColumnResize(e: MouseEvent, fieldId: string) {
		e.preventDefault();
		e.stopPropagation();
		const currentWidth = getColumnWidth(fieldId);
		resizingColumn = {
			fieldId,
			startX: e.clientX,
			startWidth: currentWidth
		};
		document.body.style.cursor = 'col-resize';
		document.body.style.userSelect = 'none';
		document.addEventListener('mousemove', handleColumnResize);
		document.addEventListener('mouseup', stopColumnResize);
	}

	function handleColumnResize(e: MouseEvent) {
		if (!resizingColumn) return;
		e.preventDefault();
		const delta = e.clientX - resizingColumn.startX;
		const newWidth = Math.max(MIN_COLUMN_WIDTH, resizingColumn.startWidth + delta);
		columnWidths = { ...columnWidths, [resizingColumn.fieldId]: newWidth };
	}

	function stopColumnResize() {
		document.body.style.cursor = '';
		document.body.style.userSelect = '';
		resizingColumn = null;
		document.removeEventListener('mousemove', handleColumnResize);
		document.removeEventListener('mouseup', stopColumnResize);
	}

	function addRecord() {
		dispatch('addRecord');
	}

	function deleteRecord(recordId: string) {
		dispatch('deleteRecord', { id: recordId });
	}

	// Record context menu functions (for color)
	function showRecordContextMenu(e: MouseEvent, recordId: string) {
		if (readonly) return;
		e.preventDefault();
		recordContextMenu = { recordId, x: e.clientX, y: e.clientY };
	}

	function closeRecordContextMenu() {
		recordContextMenu = null;
	}

	function setRecordColor(recordId: string, color: RecordColor | null) {
		dispatch('updateRecordColor', { id: recordId, color });
		closeRecordContextMenu();
	}

	function getRecordColor(record: Record): string | undefined {
		if (!record.color) return undefined;
		const colorDef = recordColors.find(c => c.value === record.color);
		return colorDef?.bg;
	}

	// Record expansion modal functions
	function expandRecord(record: Record) {
		expandedRecord = record;
	}

	function handleRecordModalUpdate(e: CustomEvent<{ values: { [key: string]: any } }>) {
		if (!expandedRecord) return;
		dispatch('updateRecord', {
			id: expandedRecord.id,
			values: e.detail.values
		});
		// Update local reference
		expandedRecord = { ...expandedRecord, values: e.detail.values };
	}

	function handleRecordModalDelete() {
		if (!expandedRecord) return;
		dispatch('deleteRecord', { id: expandedRecord.id });
		expandedRecord = null;
	}

	// Linked records functions
	function openRecordPicker(record: Record, field: Field) {
		if (readonly) return;
		const linkedTableId = field.options?.linked_table_id;
		if (!linkedTableId) return;

		pickerFieldId = field.id;
		pickerRecordId = record.id;
		pickerTableId = linkedTableId;
		pickerSelectedIds = record.values[field.id] || [];
		showRecordPicker = true;
	}

	function handleRecordPickerChange(newIds: string[]) {
		pickerSelectedIds = newIds;
	}

	function handleRecordPickerClose() {
		// Save the selected records
		dispatch('updateRecord', {
			id: pickerRecordId,
			values: { [pickerFieldId]: pickerSelectedIds }
		});
		showRecordPicker = false;
		pickerFieldId = '';
		pickerRecordId = '';
		pickerTableId = '';
		pickerSelectedIds = [];
	}

	function getLinkedRecordTitles(value: any, field: Field, _version?: number): { id: string; title: string }[] {
		// The _version param ensures Svelte tracks reactivity when cache updates
		if (!value || !Array.isArray(value) || value.length === 0) {
			return [];
		}

		const tableId = field.options?.linked_table_id;
		if (!tableId) return [];

		const titleMap = linkedRecordTitles[tableId];
		if (!titleMap) {
			// Cache not loaded yet
			return value.map((id: string) => ({ id, title: 'Loading...' }));
		}

		return value.map((id: string) => ({
			id,
			title: titleMap[id] || 'Unknown'
		}));
	}

	function getLinkedRecordDisplay(value: any, field: Field, _version?: number): string {
		const titles = getLinkedRecordTitles(value, field, _version);
		if (titles.length === 0) return '';
		if (titles.length <= 2) {
			return titles.map(t => t.title).join(', ');
		}
		return `${titles[0].title}, +${titles.length - 1} more`;
	}

	// Select field functions
	function openSelectDropdown(e: MouseEvent, record: Record, field: Field) {
		if (readonly) return;
		e.stopPropagation();
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		selectDropdown = {
			recordId: record.id,
			fieldId: field.id,
			x: rect.left,
			y: rect.bottom
		};
	}

	function closeSelectDropdown() {
		selectDropdown = null;
	}

	function selectSingleOption(optionId: string) {
		if (!selectDropdown) return;
		dispatch('updateRecord', {
			id: selectDropdown.recordId,
			values: { [selectDropdown.fieldId]: optionId }
		});
		closeSelectDropdown();
	}

	function toggleMultiOption(record: Record, field: Field, optionId: string) {
		if (readonly) return;
		const currentValue = record.values[field.id] || [];
		let newValue: string[];
		if (currentValue.includes(optionId)) {
			newValue = currentValue.filter((id: string) => id !== optionId);
		} else {
			newValue = [...currentValue, optionId];
		}
		dispatch('updateRecord', {
			id: record.id,
			values: { [field.id]: newValue }
		});
	}

	function clearSelectValue() {
		if (!selectDropdown) return;
		dispatch('updateRecord', {
			id: selectDropdown.recordId,
			values: { [selectDropdown.fieldId]: null }
		});
		closeSelectDropdown();
	}

	function getSelectOptionById(field: Field, optionId: string): { id: string; name: string; color?: string } | null {
		const options = field.options?.options || [];
		return options.find((o: any) => o.id === optionId) || null;
	}

	function getSelectOptionColor(color: string | undefined): string {
		const colors: { [key: string]: string } = {
			red: '#fee2e2',
			orange: '#ffedd5',
			yellow: '#fef3c7',
			green: '#dcfce7',
			blue: '#dbeafe',
			purple: '#f3e8ff',
			pink: '#fce7f3',
			gray: '#f3f4f6'
		};
		return colors[color || ''] || colors.gray;
	}

	// Options editor functions
	function openOptionsEditor(fieldId: string) {
		const field = fields.find(f => f.id === fieldId);
		if (!field || (field.field_type !== 'single_select' && field.field_type !== 'multi_select')) return;

		editingOptionsFieldId = fieldId;
		editingOptions = [...(field.options?.options || [])];
		newEditOptionName = '';
		newEditOptionColor = 'gray';
		closeFieldContextMenu();
	}

	function closeOptionsEditor() {
		editingOptionsFieldId = null;
		editingOptions = [];
		newEditOptionName = '';
		newEditOptionColor = 'gray';
	}

	function addEditOption() {
		if (!newEditOptionName.trim()) return;
		editingOptions = [...editingOptions, {
			id: crypto.randomUUID(),
			name: newEditOptionName.trim(),
			color: newEditOptionColor
		}];
		newEditOptionName = '';
		newEditOptionColor = 'gray';
	}

	function removeEditOption(optionId: string) {
		editingOptions = editingOptions.filter(o => o.id !== optionId);
	}

	function updateEditOptionName(optionId: string, newName: string) {
		editingOptions = editingOptions.map(o =>
			o.id === optionId ? { ...o, name: newName } : o
		);
	}

	function updateEditOptionColor(optionId: string, newColor: string) {
		editingOptions = editingOptions.map(o =>
			o.id === optionId ? { ...o, color: newColor } : o
		);
	}

	function saveOptionsChanges() {
		if (!editingOptionsFieldId) return;
		dispatch('updateFieldOptions', {
			id: editingOptionsFieldId,
			options: { options: editingOptions }
		});
		closeOptionsEditor();
	}

	// Computed field editor functions
	function openComputedFieldEditor(fieldId: string) {
		const field = fields.find(f => f.id === fieldId);
		if (!field) return;

		editingComputedFieldId = fieldId;

		if (field.field_type === 'formula') {
			editingFormulaExpression = field.options?.expression || '';
			editingFormulaResultType = field.options?.result_type || 'text';
		} else if (field.field_type === 'rollup') {
			editingRollupLinkedFieldId = field.options?.rollup_linked_field_id || '';
			editingRollupFieldId = field.options?.rollup_field_id || '';
			editingAggregationFunction = field.options?.aggregation_function || 'COUNT';
		} else if (field.field_type === 'lookup') {
			editingLookupLinkedFieldId = field.options?.lookup_linked_field_id || '';
			editingLookupFieldId = field.options?.lookup_field_id || '';
		}

		closeFieldContextMenu();
	}

	function closeComputedFieldEditor() {
		editingComputedFieldId = null;
		editingFormulaExpression = '';
		editingFormulaResultType = 'text';
		editingRollupLinkedFieldId = '';
		editingRollupFieldId = '';
		editingAggregationFunction = 'COUNT';
		editingLookupLinkedFieldId = '';
		editingLookupFieldId = '';
	}

	function saveComputedFieldChanges() {
		if (!editingComputedFieldId) return;
		const field = fields.find(f => f.id === editingComputedFieldId);
		if (!field) return;

		let options: any = {};
		if (field.field_type === 'formula') {
			options = {
				expression: editingFormulaExpression,
				result_type: editingFormulaResultType
			};
		} else if (field.field_type === 'rollup') {
			options = {
				rollup_linked_field_id: editingRollupLinkedFieldId,
				rollup_field_id: editingRollupFieldId,
				aggregation_function: editingAggregationFunction
			};
		} else if (field.field_type === 'lookup') {
			options = {
				lookup_linked_field_id: editingLookupLinkedFieldId,
				lookup_field_id: editingLookupFieldId
			};
		}

		dispatch('updateFieldOptions', {
			id: editingComputedFieldId,
			options
		});
		closeComputedFieldEditor();
	}

	// Attachment modal functions
	async function openAttachmentModal(recordId: string, fieldId: string) {
		attachmentModal = { recordId, fieldId };
		attachmentList = [];
		loadingAttachments = true;
		try {
			const { attachments } = await import('$lib/api/client');
			const result = await attachments.list(recordId, fieldId);
			attachmentList = result.attachments || [];
		} catch (error) {
			console.error('Failed to load attachments:', error);
			attachmentList = [];
		} finally {
			loadingAttachments = false;
		}
	}

	function closeAttachmentModal() {
		attachmentModal = null;
		uploadingAttachment = false;
		attachmentList = [];
		loadingAttachments = false;
	}

	async function handleAttachmentUpload(event: Event) {
		if (!attachmentModal) return;
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;

		uploadingAttachment = true;
		try {
			const { attachments } = await import('$lib/api/client');
			const newAttachment = await attachments.upload(attachmentModal.recordId, attachmentModal.fieldId, file);
			attachmentList = [...attachmentList, newAttachment];
		} catch (error) {
			console.error('Failed to upload attachment:', error);
			alert('Failed to upload attachment');
		} finally {
			uploadingAttachment = false;
			input.value = ''; // Reset input
		}
	}

	async function handleAttachmentDelete(attachmentId: string) {
		if (!confirm('Delete this attachment?')) return;
		try {
			const { attachments } = await import('$lib/api/client');
			await attachments.delete(attachmentId);
			attachmentList = attachmentList.filter(a => a.id !== attachmentId);
		} catch (error) {
			console.error('Failed to delete attachment:', error);
			alert('Failed to delete attachment');
		}
	}

	function getAuthenticatedUrl(url: string): string {
		const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;
		if (!token) return url;
		const separator = url.includes('?') ? '&' : '?';
		return `${url}${separator}token=${encodeURIComponent(token)}`;
	}

	function formatValue(value: any, field: Field): string {
		if (value === null || value === undefined) return '';

		switch (field.field_type) {
			case 'checkbox':
				return value ? '✓' : '';
			case 'date':
				return value ? new Date(value).toLocaleDateString() : '';
			case 'number':
				return value.toString();
			default:
				return String(value);
		}
	}

	function getFieldIcon(fieldType: string): string {
		switch (fieldType) {
			case 'text': return 'Aa';
			case 'number': return '#';
			case 'checkbox': return '☑';
			case 'date': return '📅';
			case 'single_select': return '◉';
			case 'multi_select': return '☰';
			case 'linked_record': return '🔗';
			case 'formula': return 'ƒx';
			case 'rollup': return 'Σ';
			case 'lookup': return '↗';
			case 'attachment': return '📎';
			default: return '?';
		}
	}

	// Keyboard navigation functions
	function selectCell(rowIndex: number, colIndex: number) {
		// Bounds checking
		if (rowIndex < 0 || rowIndex >= sortedRecords.length) return;
		if (colIndex < 0 || colIndex >= fields.length) return;
		selectedCell = { rowIndex, colIndex };
	}

	function handleCellClick(rowIndex: number, colIndex: number) {
		selectCell(rowIndex, colIndex);
	}

	function startEditSelectedCell() {
		if (!selectedCell || readonly) return;
		const record = sortedRecords[selectedCell.rowIndex];
		const field = fields[selectedCell.colIndex];
		if (!record || !field) return;

		// Don't start edit mode for special field types
		if (field.field_type === 'checkbox') {
			toggleCheckbox(record, field);
			return;
		}
		if (field.field_type === 'linked_record') {
			openRecordPicker(record, field);
			return;
		}

		startEdit(record, field);
	}

	function moveSelection(dRow: number, dCol: number) {
		if (!selectedCell) {
			// If nothing selected, select first cell
			if (sortedRecords.length > 0 && fields.length > 0) {
				selectCell(0, 0);
			}
			return;
		}

		const newRow = selectedCell.rowIndex + dRow;
		const newCol = selectedCell.colIndex + dCol;
		selectCell(newRow, newCol);
	}

	function copySelectedCell() {
		if (!selectedCell) return;
		const record = sortedRecords[selectedCell.rowIndex];
		const field = fields[selectedCell.colIndex];
		if (!record || !field) return;

		const value = record.values[field.id];
		clipboard = { value, fieldType: field.field_type };

		// Also copy to system clipboard as text
		const textValue = formatValue(value, field);
		navigator.clipboard?.writeText(textValue);
	}

	function pasteToSelectedCell() {
		if (!selectedCell || readonly) return;
		const record = sortedRecords[selectedCell.rowIndex];
		const field = fields[selectedCell.colIndex];
		if (!record || !field) return;

		// Try to use internal clipboard first (preserves type)
		if (clipboard && clipboard.fieldType === field.field_type) {
			dispatch('updateRecord', {
				id: record.id,
				values: { [field.id]: clipboard.value }
			});
			return;
		}

		// Fall back to system clipboard (text only)
		navigator.clipboard?.readText().then(text => {
			let processedValue: any = text;
			if (field.field_type === 'number') {
				const num = Number(text);
				processedValue = isNaN(num) ? null : num;
			} else if (field.field_type === 'checkbox') {
				processedValue = text.toLowerCase() === 'true' || text === '1' || text === '✓';
			}

			dispatch('updateRecord', {
				id: record.id,
				values: { [field.id]: processedValue }
			});
		});
	}

	function handleGridKeydown(e: KeyboardEvent) {
		// Don't handle if we're in an input field (editing) - except for search shortcuts
		const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
		const cmdOrCtrl = isMac ? e.metaKey : e.ctrlKey;

		// Search shortcut: Cmd/Ctrl + F (always available)
		if (cmdOrCtrl && e.key === 'f') {
			e.preventDefault();
			toggleSearch();
			return;
		}

		// Escape closes search if it's open
		if (e.key === 'Escape' && showSearch) {
			e.preventDefault();
			showSearch = false;
			searchQuery = '';
			gridContainer?.focus();
			return;
		}

		// Don't handle other shortcuts if we're in an input field (editing)
		if (editingCell) return;

		// Don't handle if a menu is open
		if (showAddFieldMenu || showFilterMenu || showSortMenu || showRecordPicker) return;

		// Show shortcuts modal: Cmd/Ctrl + /
		if (cmdOrCtrl && e.key === '/') {
			e.preventDefault();
			showShortcutsModal = true;
			return;
		}

		// Copy: Cmd/Ctrl + C
		if (cmdOrCtrl && e.key === 'c') {
			e.preventDefault();
			copySelectedCell();
			return;
		}

		// Paste: Cmd/Ctrl + V
		if (cmdOrCtrl && e.key === 'v') {
			e.preventDefault();
			pasteToSelectedCell();
			return;
		}

		// Delete selected cell content: Delete or Backspace
		if ((e.key === 'Delete' || e.key === 'Backspace') && selectedCell && !readonly) {
			e.preventDefault();
			const record = sortedRecords[selectedCell.rowIndex];
			const field = fields[selectedCell.colIndex];
			if (record && field && field.field_type !== 'linked_record') {
				let emptyValue: any = '';
				if (field.field_type === 'number') emptyValue = null;
				if (field.field_type === 'checkbox') emptyValue = false;
				dispatch('updateRecord', {
					id: record.id,
					values: { [field.id]: emptyValue }
				});
			}
			return;
		}

		// Arrow key navigation
		switch (e.key) {
			case 'ArrowUp':
				e.preventDefault();
				moveSelection(-1, 0);
				break;
			case 'ArrowDown':
				e.preventDefault();
				moveSelection(1, 0);
				break;
			case 'ArrowLeft':
				e.preventDefault();
				moveSelection(0, -1);
				break;
			case 'ArrowRight':
				e.preventDefault();
				moveSelection(0, 1);
				break;
			case 'Tab':
				e.preventDefault();
				if (e.shiftKey) {
					// Move left, or up to end of previous row
					if (selectedCell && selectedCell.colIndex === 0 && selectedCell.rowIndex > 0) {
						selectCell(selectedCell.rowIndex - 1, fields.length - 1);
					} else {
						moveSelection(0, -1);
					}
				} else {
					// Move right, or down to start of next row
					if (selectedCell && selectedCell.colIndex === fields.length - 1 && selectedCell.rowIndex < sortedRecords.length - 1) {
						selectCell(selectedCell.rowIndex + 1, 0);
					} else {
						moveSelection(0, 1);
					}
				}
				// Start editing the new cell after Tab navigation
				setTimeout(() => startEditSelectedCell(), 0);
				break;
			case 'Enter':
				e.preventDefault();
				startEditSelectedCell();
				break;
			case 'Escape':
				selectedCell = null;
				break;
		}
	}

	// When editing completes, move selection
	function handleEditKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.stopPropagation();
			saveEdit();
			// Move down after enter
			if (selectedCell) {
				moveSelection(1, 0);
			}
		} else if (e.key === 'Escape') {
			e.stopPropagation();
			cancelEdit();
		} else if (e.key === 'Tab') {
			e.preventDefault();
			e.stopPropagation();
			saveEdit();
			// Move right (or wrap to next row)
			if (selectedCell) {
				if (e.shiftKey) {
					if (selectedCell.colIndex === 0 && selectedCell.rowIndex > 0) {
						selectCell(selectedCell.rowIndex - 1, fields.length - 1);
					} else {
						moveSelection(0, -1);
					}
				} else {
					if (selectedCell.colIndex === fields.length - 1 && selectedCell.rowIndex < sortedRecords.length - 1) {
						selectCell(selectedCell.rowIndex + 1, 0);
					} else {
						moveSelection(0, 1);
					}
				}
				// Start editing the new cell
				setTimeout(() => startEditSelectedCell(), 0);
			}
		}
	}

	// Watch for editNewRecordId prop and auto-enter edit mode
	$: if (editNewRecordId) {
		// Find the record and the first editable text field
		const record = records.find(r => r.id === editNewRecordId);
		const editableField = fields.find(f => f.field_type === 'text');
		if (record && editableField) {
			// Use setTimeout to allow the DOM to update first
			setTimeout(() => {
				startEdit(record, editableField);
				dispatch('editNewRecordHandled');
			}, 100);
		} else {
			dispatch('editNewRecordHandled');
		}
	}

	// Focus grid on mount for keyboard navigation
	onMount(() => {
		if (gridContainer) {
			gridContainer.focus();
		}
	});
</script>

<div
	class="grid-container"
	bind:this={gridContainer}
	tabindex="0"
	on:keydown={handleGridKeydown}
>
	<!-- Toolbar -->
	<div class="toolbar">
		<div class="toolbar-left">
			<div class="toolbar-item">
				<button class="toolbar-btn" class:active={filters.length > 0} on:click={() => showFilterMenu = !showFilterMenu}>
					Filter {filters.length > 0 ? `(${filters.length})` : ''}
				</button>
				{#if showFilterMenu}
					<div class="dropdown-menu filter-menu" on:click|stopPropagation>
						{#if filters.length > 0}
							<div class="active-filters">
								{#each filters as filter, i}
									{@const field = fields.find(f => f.id === filter.fieldId)}
									<div class="filter-tag">
										<span>{field?.name} {filter.operator} {filter.value}</span>
										<button class="remove-filter-btn" on:click={() => removeFilter(i)}>×</button>
									</div>
								{/each}
							</div>
						{/if}
						<div class="filter-form">
							<select bind:value={newFilter.fieldId}>
								<option value="">Select field...</option>
								{#each fields as field}
									<option value={field.id}>{field.name}</option>
								{/each}
							</select>
							<select bind:value={newFilter.operator}>
								{#each filterOperators as op}
									<option value={op.value}>{op.label}</option>
								{/each}
							</select>
							{#if newFilter.operator !== 'empty' && newFilter.operator !== 'not_empty'}
								<input type="text" placeholder="Value" bind:value={newFilter.value} />
							{/if}
							<button class="add-filter-btn" on:click={addFilter} disabled={!newFilter.fieldId}>
								Add
							</button>
						</div>
					</div>
				{/if}
			</div>

			<div class="toolbar-item">
				<button class="toolbar-btn" class:active={sort !== null} on:click={() => showSortMenu = !showSortMenu}>
					Sort {sort ? `(${fields.find(f => f.id === sort.fieldId)?.name})` : ''}
				</button>
				{#if showSortMenu}
					<div class="dropdown-menu sort-menu" on:click|stopPropagation>
						{#if sort}
							<button class="clear-sort-btn" on:click={clearSort}>Clear sort</button>
						{/if}
						{#each fields as field}
							<div class="sort-option">
								<span class="sort-field-name">{field.name}</span>
								<button
									class="sort-dir-btn"
									class:active={sort?.fieldId === field.id && sort?.direction === 'asc'}
									on:click={() => setSort(field.id, 'asc')}
								>
									↑ A-Z
								</button>
								<button
									class="sort-dir-btn"
									class:active={sort?.fieldId === field.id && sort?.direction === 'desc'}
									on:click={() => setSort(field.id, 'desc')}
								>
									↓ Z-A
								</button>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<div class="toolbar-right">
			{#if showSearch}
				<div class="search-box">
					<svg class="search-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="11" cy="11" r="8"/>
						<path d="m21 21-4.35-4.35"/>
					</svg>
					<input
						type="text"
						class="search-input"
						placeholder="Search records..."
						bind:value={searchQuery}
						bind:this={searchInputRef}
						on:keydown={(e) => {
							if (e.key === 'Escape') {
								showSearch = false;
								searchQuery = '';
								gridContainer?.focus();
							}
						}}
					/>
					{#if searchQuery}
						<button class="search-clear" on:click={clearSearch}>×</button>
					{/if}
					<button class="search-close" on:click={() => { showSearch = false; searchQuery = ''; }}>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M18 6L6 18M6 6l12 12"/>
						</svg>
					</button>
				</div>
			{:else}
				<button class="toolbar-btn search-btn" on:click={toggleSearch} title="Search (⌘F)">
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="11" cy="11" r="8"/>
						<path d="m21 21-4.35-4.35"/>
					</svg>
				</button>
			{/if}
			<span class="record-count">
				{#if searchQuery}
					{sortedRecords.length} of {filteredRecords.length} record{filteredRecords.length !== 1 ? 's' : ''}
				{:else}
					{sortedRecords.length} record{sortedRecords.length !== 1 ? 's' : ''}
				{/if}
			</span>
		</div>
	</div>

	{#if someSelected && !readonly}
		<div class="bulk-actions-bar">
			<div class="bulk-info">
				<span class="selected-count">{selectedRecordIds.size} selected</span>
				<button class="bulk-clear-btn" on:click={clearSelection}>Clear</button>
			</div>
			<div class="bulk-actions">
				<div class="bulk-action-item">
					<button class="bulk-action-btn" on:click={() => showBulkColorMenu = !showBulkColorMenu}>
						<span class="color-swatch"></span>
						Color
					</button>
					{#if showBulkColorMenu}
						<div class="bulk-color-menu" on:click|stopPropagation>
							<div class="bulk-color-grid">
								{#each recordColors as color}
									<button
										class="bulk-color-option"
										style="background: {color.bg}"
										title={color.label}
										on:click={() => bulkSetColor(color.value)}
									></button>
								{/each}
							</div>
							<button class="bulk-color-clear" on:click={() => bulkSetColor(null)}>Clear color</button>
						</div>
					{/if}
				</div>
				<button class="bulk-action-btn danger" on:click={bulkDeleteRecords}>
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M3 6h18M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"/>
					</svg>
					Delete
				</button>
			</div>
		</div>
	{/if}

	<div class="grid" role="grid">
		<!-- Header row -->
		<div class="grid-row header-row" role="row">
			<div class="cell row-number header-cell" role="columnheader">
				{#if !readonly}
					<input
						type="checkbox"
						class="row-checkbox"
						checked={allSelected}
						indeterminate={someSelected && !allSelected}
						on:change={toggleSelectAll}
						title="Select all"
					/>
				{/if}
			</div>
			{#each fields as field}
				{#if editingFieldId === field.id}
					<div
						class="cell header-cell editing"
						role="columnheader"
						style:width="{columnWidths[field.id] ?? 150}px"
						style:min-width="{columnWidths[field.id] ?? 150}px"
						style:max-width="{columnWidths[field.id] ?? 150}px"
					>
						<input
							type="text"
							class="field-name-input"
							bind:value={editingFieldName}
							on:blur={saveFieldName}
							on:keydown={(e) => {
								if (e.key === 'Enter') saveFieldName();
								if (e.key === 'Escape') { editingFieldId = null; }
							}}
							autofocus
						/>
					</div>
				{:else}
					<div
						class="cell header-cell"
						class:dragging={draggingFieldId === field.id}
						class:drag-over={dragOverFieldId === field.id}
						class:resizing={resizingColumn?.fieldId === field.id}
						role="columnheader"
						title={field.field_type}
						style:width="{columnWidths[field.id] ?? 150}px"
						style:min-width="{columnWidths[field.id] ?? 150}px"
						style:max-width="{columnWidths[field.id] ?? 150}px"
						draggable={!readonly && !resizingColumn}
						on:contextmenu={(e) => showFieldContextMenu(e, field.id)}
						on:dblclick={() => !readonly && startEditField(field.id)}
						on:dragstart={(e) => handleFieldDragStart(e, field.id)}
						on:dragover={(e) => handleFieldDragOver(e, field.id)}
						on:dragleave={handleFieldDragLeave}
						on:drop={(e) => handleFieldDrop(e, field.id)}
						on:dragend={handleFieldDragEnd}
					>
						{#if !readonly}
							<span class="drag-handle">⋮⋮</span>
						{/if}
						<span class="field-icon">{getFieldIcon(field.field_type)}</span>
						<span class="field-name">{field.name}</span>
						{#if !readonly}
							<div
								class="resize-handle"
								draggable="false"
								on:mousedown|stopPropagation={(e) => startColumnResize(e, field.id)}
								on:dragstart|preventDefault|stopPropagation
							></div>
						{/if}
					</div>
				{/if}
			{/each}
			{#if !readonly}
				<div class="cell add-field-cell" role="columnheader">
					<button class="add-field-btn" on:click={() => showAddFieldMenu = !showAddFieldMenu}>
						+
					</button>
					{#if showAddFieldMenu}
						<div class="add-field-menu" on:click|stopPropagation>
							<input
								type="text"
								placeholder="Field name"
								bind:value={newFieldName}
								autofocus
							/>
							<div class="field-type-row">
								<select bind:value={newFieldType}>
									{#each fieldTypes as ft}
										<option value={ft.value}>{ft.label}</option>
									{/each}
								</select>
								<ContextualHelp topic="fields" position="right" />
							</div>
							{#if newFieldType === 'linked_record'}
								<select bind:value={newFieldLinkedTableId}>
									<option value="">Select table to link...</option>
									{#each linkableTables as table}
										<option value={table.id}>{table.name}</option>
									{/each}
								</select>
							{/if}
							{#if newFieldType === 'single_select' || newFieldType === 'multi_select'}
								<div class="select-options-editor">
									<label>Options:</label>
									{#if newSelectOptions.length > 0}
										<div class="options-list">
											{#each newSelectOptions as option}
												<div class="option-item">
													<span class="option-color-dot" style="background: {getSelectOptionColor(option.color)}"></span>
													<span class="option-name">{option.name}</span>
													<button class="remove-option-btn" on:click={() => removeSelectOption(option.id)}>×</button>
												</div>
											{/each}
										</div>
									{/if}
									<div class="add-option-row">
										<input
											type="text"
											placeholder="Option name"
											bind:value={newOptionName}
											on:keydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addSelectOption(); }}}
										/>
										<select bind:value={newOptionColor} class="color-select">
											{#each selectColors as color}
												<option value={color}>{color}</option>
											{/each}
										</select>
										<button class="add-option-btn" on:click={addSelectOption} disabled={!newOptionName.trim()}>+</button>
									</div>
								</div>
							{/if}
							{#if newFieldType === 'formula'}
								<div class="formula-config">
									<label>Formula expression:</label>
									<textarea
										bind:value={newFormulaExpression}
										placeholder="e.g., CONCAT({"{Name}"}, ' - ', {"{Status}"})"
										rows="3"
									></textarea>
									<small class="formula-help">
										Use {"{Field Name}"} to reference fields. Functions: CONCAT, UPPER, LOWER, IF, SUM, etc.
									</small>
									<label>Result type:</label>
									<select bind:value={newFormulaResultType}>
										<option value="text">Text</option>
										<option value="number">Number</option>
										<option value="date">Date</option>
										<option value="boolean">Checkbox</option>
									</select>
								</div>
							{/if}
							{#if newFieldType === 'rollup'}
								<div class="rollup-config">
									<label>Link field to rollup from:</label>
									<select bind:value={newRollupLinkedFieldId}>
										<option value="">Select linked field...</option>
										{#each linkedRecordFields as field}
											<option value={field.id}>{field.name}</option>
										{/each}
									</select>
									<label>Aggregation:</label>
									<select bind:value={newAggregationFunction}>
										<option value="COUNT">Count</option>
										<option value="COUNTA">Count (non-empty)</option>
										<option value="SUM">Sum</option>
										<option value="AVG">Average</option>
										<option value="MIN">Min</option>
										<option value="MAX">Max</option>
									</select>
									{#if newAggregationFunction !== 'COUNT'}
										<label>Field to aggregate (optional for count):</label>
										<input type="text" bind:value={newRollupFieldId} placeholder="Field ID from linked table" />
									{/if}
								</div>
							{/if}
							{#if newFieldType === 'lookup'}
								<div class="lookup-config">
									<label>Link field to lookup from:</label>
									<select bind:value={newLookupLinkedFieldId}>
										<option value="">Select linked field...</option>
										{#each linkedRecordFields as field}
											<option value={field.id}>{field.name}</option>
										{/each}
									</select>
									<label>Field ID to lookup:</label>
									<input type="text" bind:value={newLookupFieldId} placeholder="Field ID from linked table" />
								</div>
							{/if}
							<div class="menu-actions">
								<button class="cancel-btn" on:click={() => { showAddFieldMenu = false; newSelectOptions = []; }}>Cancel</button>
								<button
									class="create-btn"
									on:click={addField}
									disabled={!newFieldName.trim() ||
										(newFieldType === 'linked_record' && !newFieldLinkedTableId) ||
										((newFieldType === 'single_select' || newFieldType === 'multi_select') && newSelectOptions.length === 0) ||
										(newFieldType === 'formula' && !newFormulaExpression.trim()) ||
										(newFieldType === 'rollup' && !newRollupLinkedFieldId) ||
										(newFieldType === 'lookup' && (!newLookupLinkedFieldId || !newLookupFieldId))}
								>
									Create
								</button>
							</div>
						</div>
					{/if}
				</div>
			{/if}
		</div>

		<!-- Data rows -->
		{#each sortedRecords as record, index}
			{@const rowBgColor = getRecordColor(record)}
			<div class="grid-row" class:colored-row={rowBgColor} role="row" style={rowBgColor ? `--row-bg: ${rowBgColor}` : ''}>
				<div
					class="cell row-number"
					class:row-selected={selectedRecordIds.has(record.id)}
					role="rowheader"
					on:contextmenu={(e) => showRecordContextMenu(e, record.id)}
					on:dblclick={() => expandRecord(record)}
				>
					{#if record.color}
						<span class="color-indicator" style="background: {recordColors.find(c => c.value === record.color)?.bg}"></span>
					{/if}
					{#if !readonly}
						<input
							type="checkbox"
							class="row-checkbox data-row-checkbox"
							checked={selectedRecordIds.has(record.id)}
							on:change={() => toggleRecordSelection(record.id)}
							on:click|stopPropagation
						/>
					{/if}
					<span class="row-index" on:dblclick|stopPropagation={() => expandRecord(record)} title="Double-click to expand">{index + 1}</span>
					{#if !readonly}
						<button class="delete-row-btn" on:click|stopPropagation={() => deleteRecord(record.id)} title="Delete row">
							×
						</button>
					{/if}
				</div>
				{#each fields as field, colIndex}
					{@const isEditing = editingCell?.recordId === record.id && editingCell?.fieldId === field.id}
					{@const isSelected = selectedCell?.rowIndex === index && selectedCell?.colIndex === colIndex}
					<div
						class="cell data-cell"
						class:editing={isEditing}
						class:selected={isSelected && !isEditing}
						class:linked-cell={field.field_type === 'linked_record'}
						class:select-cell={field.field_type === 'single_select' || field.field_type === 'multi_select'}
						style:width="{columnWidths[field.id] ?? 150}px"
						style:min-width="{columnWidths[field.id] ?? 150}px"
						style:max-width="{columnWidths[field.id] ?? 150}px"
						role="gridcell"
						on:dblclick={() => {
							// Computed fields (formula, rollup, lookup) and attachment are not directly editable
							const nonEditableTypes = ['checkbox', 'linked_record', 'single_select', 'multi_select', 'formula', 'rollup', 'lookup', 'attachment'];
							if (!nonEditableTypes.includes(field.field_type)) {
								startEdit(record, field);
							}
						}}
						on:click={(e) => {
							handleCellClick(index, colIndex);
							if (field.field_type === 'checkbox') toggleCheckbox(record, field);
							else if (field.field_type === 'linked_record') openRecordPicker(record, field);
							else if (field.field_type === 'single_select') openSelectDropdown(e, record, field);
							// Formula, rollup, lookup fields are read-only - no action on click
						}}
					>
						{#if isEditing}
							{#if field.field_type === 'number'}
								<input
									type="number"
									class="cell-input"
									bind:value={editValue}
									on:blur={saveEdit}
									on:keydown={handleEditKeydown}
									autofocus
								/>
							{:else if field.field_type === 'date'}
								<input
									type="date"
									class="cell-input"
									bind:value={editValue}
									on:blur={saveEdit}
									on:keydown={handleEditKeydown}
									autofocus
								/>
							{:else}
								<input
									type="text"
									class="cell-input"
									bind:value={editValue}
									on:blur={saveEdit}
									on:keydown={handleEditKeydown}
									autofocus
								/>
							{/if}
						{:else}
							{#if field.field_type === 'checkbox'}
								<span class="checkbox-display" class:checked={record.values[field.id]}>
									{record.values[field.id] ? '✓' : ''}
								</span>
							{:else if field.field_type === 'linked_record'}
								{@const linkedTitles = getLinkedRecordTitles(record.values[field.id], field, linkedCacheVersion)}
								<span class="linked-value">
									{#if linkedTitles.length === 0}
										<span class="linked-placeholder">Click to link records</span>
									{:else}
										{#each linkedTitles.slice(0, 3) as linked}
											<span class="linked-pill" title={linked.title}>{linked.title}</span>
										{/each}
										{#if linkedTitles.length > 3}
											<span class="linked-more">+{linkedTitles.length - 3}</span>
										{/if}
									{/if}
								</span>
							{:else if field.field_type === 'single_select'}
								{@const option = getSelectOptionById(field, record.values[field.id])}
								<span class="select-value">
									{#if option}
										<span class="select-pill" style="background: {getSelectOptionColor(option.color)}">{option.name}</span>
									{:else}
										<span class="select-placeholder">Select option...</span>
									{/if}
								</span>
							{:else if field.field_type === 'multi_select'}
								{@const selectedIds = record.values[field.id] || []}
								<span class="multi-select-value" on:click={(e) => openSelectDropdown(e, record, field)}>
									{#if selectedIds.length === 0}
										<span class="select-placeholder">Select options...</span>
									{:else}
										{#each selectedIds.slice(0, 3) as optId}
											{@const opt = getSelectOptionById(field, optId)}
											{#if opt}
												<span class="select-pill" style="background: {getSelectOptionColor(opt.color)}">{opt.name}</span>
											{/if}
										{/each}
										{#if selectedIds.length > 3}
											<span class="select-more">+{selectedIds.length - 3}</span>
										{/if}
									{/if}
								</span>
							{:else if field.field_type === 'formula' || field.field_type === 'rollup' || field.field_type === 'lookup'}
								<span class="cell-value computed-value" title="Computed field (read-only)">
									{formatValue(record.values[field.id], field)}
								</span>
							{:else if field.field_type === 'attachment'}
								{@const attachments = record.values[field.id] || []}
								<span class="attachment-value clickable" on:click|stopPropagation={() => openAttachmentModal(record.id, field.id)}>
									{#if Array.isArray(attachments) && attachments.length > 0}
										<span class="attachment-count">📎 {attachments.length} file{attachments.length > 1 ? 's' : ''}</span>
									{:else}
										<span class="attachment-placeholder">+ Add files</span>
									{/if}
								</span>
							{:else}
								<span class="cell-value">{formatValue(record.values[field.id], field)}</span>
							{/if}
						{/if}
					</div>
				{/each}
				{#if !readonly}
					<div class="cell spacer-cell"></div>
				{/if}
			</div>
		{/each}

		<!-- Add row button -->
		{#if !readonly}
			<div class="grid-row add-row" role="row">
				<div class="cell row-number">
					<button class="add-row-btn" on:click={addRecord}>+</button>
				</div>
				{#each fields as _}
					<div class="cell empty-cell"></div>
				{/each}
				<div class="cell spacer-cell"></div>
			</div>
		{/if}
	</div>

	{#if fields.length === 0}
		<div class="empty-grid">
			<p>No fields yet</p>
			{#if !readonly}
				<p class="hint">Click the + button to add your first field</p>
			{/if}
		</div>
	{/if}
</div>

{#if showRecordPicker && pickerTableId}
	<RecordPicker
		tableId={pickerTableId}
		selectedIds={pickerSelectedIds}
		multiple={true}
		on:change={(e) => handleRecordPickerChange(e.detail)}
		on:close={handleRecordPickerClose}
	/>
{/if}

{#if fieldContextMenu}
	{@const contextField = fields.find(f => f.id === fieldContextMenu.fieldId)}
	<div class="context-menu-overlay" on:click={closeFieldContextMenu}>
		<div
			class="context-menu"
			style="left: {fieldContextMenu.x}px; top: {fieldContextMenu.y}px;"
			on:click|stopPropagation
		>
			<button class="context-menu-item" on:click={() => startEditField(fieldContextMenu.fieldId)}>
				Rename
			</button>
			{#if contextField?.field_type === 'single_select' || contextField?.field_type === 'multi_select'}
				<button class="context-menu-item" on:click={() => openOptionsEditor(fieldContextMenu.fieldId)}>
					Edit Options
				</button>
			{/if}
			{#if contextField?.field_type === 'formula' || contextField?.field_type === 'rollup' || contextField?.field_type === 'lookup'}
				<button class="context-menu-item" on:click={() => openComputedFieldEditor(fieldContextMenu.fieldId)}>
					Edit Configuration
				</button>
			{/if}
			<button class="context-menu-item danger" on:click={() => deleteField(fieldContextMenu.fieldId)}>
				Delete
			</button>
		</div>
	</div>
{/if}

{#if recordContextMenu}
	<div class="context-menu-overlay" on:click={closeRecordContextMenu}>
		<div
			class="context-menu color-menu"
			style="left: {recordContextMenu.x}px; top: {recordContextMenu.y}px;"
			on:click|stopPropagation
		>
			<div class="context-menu-header">Row Color</div>
			<div class="color-grid">
				{#each recordColors as color}
					<button
						class="color-option"
						style="background: {color.bg}"
						title={color.label}
						on:click={() => setRecordColor(recordContextMenu.recordId, color.value)}
					></button>
				{/each}
			</div>
			<button class="context-menu-item" on:click={() => setRecordColor(recordContextMenu.recordId, null)}>
				Clear color
			</button>
		</div>
	</div>
{/if}

{#if selectDropdown}
	{@const field = fields.find(f => f.id === selectDropdown.fieldId)}
	{@const record = records.find(r => r.id === selectDropdown.recordId)}
	{@const options = field?.options?.options || []}
	{@const isMulti = field?.field_type === 'multi_select'}
	{@const currentValue = record?.values[selectDropdown.fieldId]}
	<div class="context-menu-overlay" on:click={closeSelectDropdown}>
		<div
			class="select-dropdown"
			style="left: {selectDropdown.x}px; top: {selectDropdown.y}px;"
			on:click|stopPropagation
		>
			{#if options.length === 0}
				<div class="select-dropdown-empty">No options defined</div>
			{:else}
				{#each options as option}
					{@const isSelected = isMulti
						? (currentValue || []).includes(option.id)
						: currentValue === option.id}
					<button
						class="select-option"
						class:selected={isSelected}
						on:click={() => {
							if (isMulti && record && field) {
								toggleMultiOption(record, field, option.id);
							} else {
								selectSingleOption(option.id);
							}
						}}
					>
						{#if isMulti}
							<span class="option-checkbox">{isSelected ? '☑' : '☐'}</span>
						{/if}
						<span class="option-pill" style="background: {getSelectOptionColor(option.color)}">{option.name}</span>
					</button>
				{/each}
			{/if}
			{#if !isMulti && currentValue}
				<button class="select-option clear-option" on:click={clearSelectValue}>
					Clear selection
				</button>
			{/if}
			{#if isMulti}
				<button class="select-done-btn" on:click={closeSelectDropdown}>Done</button>
			{/if}
		</div>
	</div>
{/if}

{#if showShortcutsModal}
	<KeyboardShortcutsModal on:close={() => showShortcutsModal = false} />
{/if}

{#if editingOptionsFieldId}
	{@const editingField = fields.find(f => f.id === editingOptionsFieldId)}
	<div class="modal-overlay" on:click={closeOptionsEditor}>
		<div class="options-editor-modal" on:click|stopPropagation>
			<div class="modal-header">
				<h3>Edit {editingField?.field_type === 'single_select' ? 'Single Select' : 'Multi Select'} Options</h3>
				<button class="modal-close" on:click={closeOptionsEditor}>×</button>
			</div>
			<div class="modal-body">
				<div class="options-list-editor">
					{#each editingOptions as option, index}
						<div class="option-row">
							<select
								class="option-color-select"
								value={option.color}
								on:change={(e) => updateEditOptionColor(option.id, e.currentTarget.value)}
							>
								{#each selectColors as color}
									<option value={color}>{color}</option>
								{/each}
							</select>
							<span class="option-color-preview" style="background: {getSelectOptionColor(option.color)}"></span>
							<input
								type="text"
								class="option-name-input"
								value={option.name}
								on:input={(e) => updateEditOptionName(option.id, e.currentTarget.value)}
								on:keydown={(e) => { if (e.key === 'Enter') e.currentTarget.blur(); }}
							/>
							<button class="option-remove" on:click={() => removeEditOption(option.id)} title="Remove option">×</button>
						</div>
					{/each}
				</div>
				<div class="add-option-section">
					<div class="add-option-row">
						<select bind:value={newEditOptionColor} class="option-color-select">
							{#each selectColors as color}
								<option value={color}>{color}</option>
							{/each}
						</select>
						<span class="option-color-preview" style="background: {getSelectOptionColor(newEditOptionColor)}"></span>
						<input
							type="text"
							class="option-name-input"
							placeholder="New option name"
							bind:value={newEditOptionName}
							on:keydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addEditOption(); }}}
						/>
						<button class="option-add" on:click={addEditOption} disabled={!newEditOptionName.trim()}>Add</button>
					</div>
				</div>
			</div>
			<div class="modal-footer">
				<button class="btn-cancel" on:click={closeOptionsEditor}>Cancel</button>
				<button class="btn-save" on:click={saveOptionsChanges}>Save Changes</button>
			</div>
		</div>
	</div>
{/if}

{#if editingComputedFieldId}
	{@const editingField = fields.find(f => f.id === editingComputedFieldId)}
	<div class="modal-overlay" on:click={closeComputedFieldEditor}>
		<div class="computed-field-modal" on:click|stopPropagation>
			<div class="modal-header">
				<h3>Edit {editingField?.field_type === 'formula' ? 'Formula' : editingField?.field_type === 'rollup' ? 'Rollup' : 'Lookup'} Field</h3>
				<button class="modal-close" on:click={closeComputedFieldEditor}>×</button>
			</div>
			<div class="modal-body">
				{#if editingField?.field_type === 'formula'}
					<div class="form-group">
						<label for="formula-expression">Formula Expression</label>
						<textarea
							id="formula-expression"
							bind:value={editingFormulaExpression}
							placeholder="e.g., {'{'}Quantity{'}'} * {'{'}Price{'}'}"
							rows="3"
						></textarea>
						<p class="form-help">Reference fields using {'{'}FieldName{'}'}. Available functions: CONCAT, UPPER, LOWER, SUM, IF, AND, OR, TODAY, etc.</p>
					</div>
					<div class="form-group">
						<label for="formula-result-type">Result Type</label>
						<select id="formula-result-type" bind:value={editingFormulaResultType}>
							<option value="text">Text</option>
							<option value="number">Number</option>
							<option value="date">Date</option>
							<option value="boolean">Boolean</option>
						</select>
					</div>
				{:else if editingField?.field_type === 'rollup'}
					<div class="form-group">
						<label for="rollup-linked-field">Linked Record Field</label>
						<select id="rollup-linked-field" bind:value={editingRollupLinkedFieldId}>
							<option value="">Select a linked record field</option>
							{#each linkedRecordFields as f}
								<option value={f.id}>{f.name}</option>
							{/each}
						</select>
					</div>
					<div class="form-group">
						<label for="rollup-field">Field to Rollup</label>
						<input
							id="rollup-field"
							type="text"
							bind:value={editingRollupFieldId}
							placeholder="Enter field ID from linked table"
						/>
						<p class="form-help">The field ID from the linked table to aggregate</p>
					</div>
					<div class="form-group">
						<label for="rollup-aggregation">Aggregation Function</label>
						<select id="rollup-aggregation" bind:value={editingAggregationFunction}>
							<option value="COUNT">COUNT</option>
							<option value="COUNTA">COUNTA</option>
							<option value="SUM">SUM</option>
							<option value="AVG">AVG</option>
							<option value="MIN">MIN</option>
							<option value="MAX">MAX</option>
						</select>
					</div>
				{:else if editingField?.field_type === 'lookup'}
					<div class="form-group">
						<label for="lookup-linked-field">Linked Record Field</label>
						<select id="lookup-linked-field" bind:value={editingLookupLinkedFieldId}>
							<option value="">Select a linked record field</option>
							{#each linkedRecordFields as f}
								<option value={f.id}>{f.name}</option>
							{/each}
						</select>
					</div>
					<div class="form-group">
						<label for="lookup-field">Field to Lookup</label>
						<input
							id="lookup-field"
							type="text"
							bind:value={editingLookupFieldId}
							placeholder="Enter field ID from linked table"
						/>
						<p class="form-help">The field ID from the linked table to lookup</p>
					</div>
				{/if}
			</div>
			<div class="modal-footer">
				<button class="btn-cancel" on:click={closeComputedFieldEditor}>Cancel</button>
				<button class="btn-save" on:click={saveComputedFieldChanges}>Save Changes</button>
			</div>
		</div>
	</div>
{/if}

{#if attachmentModal}
	{@const attachmentField = fields.find(f => f.id === attachmentModal.fieldId)}
	<div class="modal-overlay" on:click={closeAttachmentModal}>
		<div class="attachment-modal" on:click|stopPropagation>
			<div class="modal-header">
				<h3>Attachments - {attachmentField?.name || 'Unknown'}</h3>
				<button class="modal-close" on:click={closeAttachmentModal}>×</button>
			</div>
			<div class="modal-body">
				{#if loadingAttachments}
					<p class="no-attachments">Loading attachments...</p>
				{:else if attachmentList.length > 0}
					<div class="attachment-list">
						{#each attachmentList as attachment}
							<div class="attachment-item">
								<div class="attachment-info">
									<span class="attachment-icon">📎</span>
									<div class="attachment-details">
										<span class="attachment-name">{attachment.filename}</span>
										<span class="attachment-size">{Math.round(attachment.size_bytes / 1024)} KB</span>
									</div>
								</div>
								<div class="attachment-actions">
									<a href={getAuthenticatedUrl(attachment.url)} target="_blank" class="btn-small">View</a>
									{#if !readonly}
										<button class="btn-small danger" on:click={() => handleAttachmentDelete(attachment.id)}>
											Delete
										</button>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<p class="no-attachments">No attachments yet</p>
				{/if}
				{#if !readonly}
					<div class="upload-section">
						<label class="upload-btn">
							{#if uploadingAttachment}
								Uploading...
							{:else}
								+ Add Attachment
							{/if}
							<input
								type="file"
								hidden
								on:change={handleAttachmentUpload}
								disabled={uploadingAttachment}
							/>
						</label>
					</div>
				{/if}
			</div>
			<div class="modal-footer">
				<button class="btn-cancel" on:click={closeAttachmentModal}>Close</button>
			</div>
		</div>
	</div>
{/if}

{#if expandedRecord}
	<RecordModal
		record={expandedRecord}
		{fields}
		{tables}
		{readonly}
		{currentUser}
		on:close={() => expandedRecord = null}
		on:update={handleRecordModalUpdate}
		on:delete={handleRecordModalDelete}
	/>
{/if}

<style>
	.grid-container {
		flex: 1;
		overflow: auto;
		background: white;
		display: flex;
		flex-direction: column;
	}

	.grid-container:focus {
		outline: none;
	}

	/* Toolbar styles */
	.toolbar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-xs) var(--spacing-md);
		border-bottom: 1px solid var(--color-border);
		background: var(--color-gray-50);
		flex-shrink: 0;
	}

	.toolbar-left {
		display: flex;
		gap: var(--spacing-sm);
	}

	.toolbar-item {
		position: relative;
	}

	.toolbar-btn {
		padding: var(--spacing-xs) var(--spacing-sm);
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: var(--font-size-sm);
		cursor: pointer;
		color: var(--color-text-muted);
	}

	.toolbar-btn:hover {
		background: var(--color-gray-100);
	}

	.toolbar-btn.active {
		background: var(--color-primary-light);
		border-color: var(--color-primary);
		color: var(--color-primary);
	}

	.dropdown-menu {
		position: absolute;
		top: 100%;
		left: 0;
		margin-top: var(--spacing-xs);
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-lg);
		z-index: 50;
		min-width: 280px;
		padding: var(--spacing-sm);
	}

	.filter-menu {
		min-width: 360px;
	}

	.active-filters {
		display: flex;
		flex-wrap: wrap;
		gap: var(--spacing-xs);
		margin-bottom: var(--spacing-sm);
		padding-bottom: var(--spacing-sm);
		border-bottom: 1px solid var(--color-border);
	}

	.filter-tag {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
		padding: 2px var(--spacing-xs);
		background: var(--color-primary-light);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-xs);
		color: var(--color-primary);
	}

	.remove-filter-btn {
		background: none;
		border: none;
		cursor: pointer;
		font-size: 14px;
		color: var(--color-primary);
		padding: 0;
		line-height: 1;
	}

	.filter-form {
		display: flex;
		gap: var(--spacing-xs);
		align-items: center;
	}

	.filter-form select,
	.filter-form input {
		padding: var(--spacing-xs) var(--spacing-sm);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
	}

	.filter-form select {
		min-width: 100px;
	}

	.filter-form input {
		width: 100px;
	}

	.add-filter-btn {
		padding: var(--spacing-xs) var(--spacing-sm);
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
		cursor: pointer;
	}

	.add-filter-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.sort-menu {
		min-width: 240px;
	}

	.clear-sort-btn {
		display: block;
		width: 100%;
		padding: var(--spacing-xs) var(--spacing-sm);
		background: var(--color-gray-100);
		border: none;
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
		cursor: pointer;
		margin-bottom: var(--spacing-sm);
		text-align: left;
	}

	.sort-option {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
		padding: var(--spacing-xs) 0;
	}

	.sort-field-name {
		flex: 1;
		font-size: var(--font-size-sm);
	}

	.sort-dir-btn {
		padding: 2px var(--spacing-xs);
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-xs);
		cursor: pointer;
	}

	.sort-dir-btn:hover {
		background: var(--color-gray-100);
	}

	.sort-dir-btn.active {
		background: var(--color-primary);
		border-color: var(--color-primary);
		color: white;
	}

	.toolbar-right {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
	}

	.record-count {
		font-size: var(--font-size-sm);
		color: var(--color-text-muted);
	}

	/* Search styles */
	.search-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		padding: var(--spacing-xs);
	}

	.search-box {
		display: flex;
		align-items: center;
		gap: 6px;
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		padding: 4px 8px;
		width: 240px;
	}

	.search-box:focus-within {
		border-color: var(--color-primary);
		box-shadow: 0 0 0 3px var(--color-primary-light);
	}

	.search-icon {
		color: var(--color-text-muted);
		flex-shrink: 0;
	}

	.search-input {
		flex: 1;
		border: none;
		outline: none;
		font-size: var(--font-size-sm);
		background: transparent;
		min-width: 0;
	}

	.search-input::placeholder {
		color: var(--color-text-muted);
	}

	.search-clear {
		width: 18px;
		height: 18px;
		border: none;
		background: var(--color-gray-200);
		color: var(--color-text-muted);
		border-radius: 50%;
		cursor: pointer;
		font-size: 14px;
		line-height: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.search-clear:hover {
		background: var(--color-gray-300);
		color: var(--color-text);
	}

	.search-close {
		width: 20px;
		height: 20px;
		border: none;
		background: transparent;
		color: var(--color-text-muted);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: var(--radius-sm);
		flex-shrink: 0;
	}

	.search-close:hover {
		background: var(--color-gray-100);
		color: var(--color-text);
	}

	.grid {
		display: flex;
		flex-direction: column;
		min-width: 100%;
	}

	.grid-row {
		display: flex;
		flex-direction: row;
	}

	.cell {
		flex-shrink: 0;
		border: 1px solid var(--color-border);
		border-left: none;
		padding: 0;
		height: 32px;
		display: flex;
		align-items: center;
		position: relative;
		box-sizing: border-box;
	}

	.cell:first-child {
		border-left: 1px solid var(--color-border);
	}

	.grid-row + .grid-row .cell {
		border-top: none;
	}

	.row-number {
		width: 70px;
		min-width: 70px;
		max-width: 70px;
		background: var(--color-gray-50);
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
		justify-content: center;
		position: relative;
		cursor: pointer;
	}

	.row-number:hover .delete-row-btn {
		opacity: 1;
	}

	.row-index {
		display: inline-block;
		cursor: pointer;
	}

	.row-index:hover {
		color: var(--color-primary);
	}

	.delete-row-btn {
		position: absolute;
		right: 2px;
		top: 50%;
		transform: translateY(-50%);
		width: 18px;
		height: 18px;
		border: none;
		background: var(--color-error);
		color: white;
		border-radius: 50%;
		cursor: pointer;
		font-size: 14px;
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0;
		transition: opacity 0.15s;
	}

	.header-row {
		background: var(--color-gray-50);
		position: sticky;
		top: 0;
		z-index: 10;
	}

	.header-cell {
		font-weight: 500;
		font-size: var(--font-size-sm);
		padding: var(--spacing-xs) var(--spacing-sm);
		white-space: nowrap;
		cursor: grab;
		user-select: none;
		position: relative;
		overflow: visible;
	}

	.header-cell:active {
		cursor: grabbing;
	}

	.header-cell.dragging {
		opacity: 0.5;
		background: var(--color-gray-200);
	}

	.header-cell.drag-over {
		background: var(--color-primary-light);
		border-left: 3px solid var(--color-primary);
	}

	.drag-handle {
		display: inline-block;
		width: 14px;
		color: var(--color-text-muted);
		font-size: 10px;
		letter-spacing: -2px;
		opacity: 0;
		transition: opacity 0.15s;
		cursor: grab;
	}

	.header-cell:hover .drag-handle {
		opacity: 1;
	}

	/* Column resize handle */
	.resize-handle {
		position: absolute;
		right: -3px;
		top: 0;
		bottom: 0;
		width: 8px;
		cursor: col-resize;
		background: transparent;
		z-index: 10;
	}

	.resize-handle:hover {
		background: rgba(45, 127, 249, 0.4);
	}

	.header-cell.resizing .resize-handle {
		background: var(--color-primary);
	}

	.header-cell.resizing {
		background: var(--color-primary-light);
		user-select: none;
	}

	.field-icon {
		display: inline-block;
		width: 20px;
		color: var(--color-text-muted);
		font-size: 12px;
	}

	.field-name {
		color: var(--color-text);
	}

	.data-cell {
		cursor: pointer;
		overflow: hidden;
		padding: 0 var(--spacing-sm);
	}

	.data-cell:hover {
		background: var(--color-gray-50);
	}

	.data-cell.editing {
		padding: 0;
		background: white;
	}

	.data-cell.selected {
		outline: 2px solid var(--color-primary);
		outline-offset: -2px;
		background: var(--color-primary-light);
	}

	.cell-input {
		width: 100%;
		height: 100%;
		border: 2px solid var(--color-primary);
		padding: var(--spacing-xs) var(--spacing-sm);
		font-size: var(--font-size-sm);
		outline: none;
		box-sizing: border-box;
	}

	.cell-value {
		display: block;
		padding: var(--spacing-xs) var(--spacing-sm);
		font-size: var(--font-size-sm);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.checkbox-display {
		display: flex;
		align-items: center;
		justify-content: center;
		height: 100%;
		font-size: 16px;
		color: var(--color-text-muted);
	}

	.checkbox-display.checked {
		color: var(--color-primary);
	}

	.linked-cell {
		cursor: pointer;
	}

	.linked-cell:hover {
		background: var(--color-primary-light);
	}

	.linked-value {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		padding: 4px 6px;
		align-items: center;
	}

	.linked-pill {
		display: inline-block;
		padding: 2px 8px;
		background: var(--color-primary-light);
		color: var(--color-primary);
		border-radius: 12px;
		font-size: 12px;
		font-weight: 500;
		max-width: 120px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.linked-more {
		display: inline-block;
		padding: 2px 6px;
		background: var(--color-gray-200);
		color: var(--color-text-muted);
		border-radius: 12px;
		font-size: 11px;
	}

	.linked-placeholder {
		color: var(--color-text-muted);
		font-size: 12px;
		font-style: italic;
	}

	.add-field-cell {
		width: 40px;
		min-width: 40px;
		justify-content: center;
		position: relative;
		flex-grow: 1;
	}

	.add-field-btn {
		width: 24px;
		height: 24px;
		border: 1px dashed var(--color-border);
		background: white;
		border-radius: var(--radius-sm);
		cursor: pointer;
		font-size: 16px;
		color: var(--color-text-muted);
	}

	.add-field-btn:hover {
		border-color: var(--color-primary);
		color: var(--color-primary);
	}

	.add-field-menu {
		position: absolute;
		top: 100%;
		right: 0;
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-lg);
		padding: var(--spacing-md);
		width: 220px;
		z-index: 20;
	}

	.add-field-menu input,
	.add-field-menu select {
		width: 100%;
		padding: var(--spacing-sm);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: var(--font-size-sm);
		margin-bottom: var(--spacing-sm);
	}

	.field-type-row {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
	}

	.field-type-row select {
		flex: 1;
		margin-bottom: 0;
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

	/* Select options editor */
	.select-options-editor {
		margin-bottom: var(--spacing-sm);
	}

	.select-options-editor label {
		display: block;
		font-size: var(--font-size-sm);
		font-weight: 500;
		margin-bottom: var(--spacing-xs);
		color: var(--color-text-muted);
	}

	.options-list {
		display: flex;
		flex-direction: column;
		gap: 4px;
		margin-bottom: var(--spacing-xs);
	}

	.option-item {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 4px 8px;
		background: var(--color-gray-50);
		border-radius: var(--radius-sm);
	}

	.option-color-dot {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.option-name {
		flex: 1;
		font-size: var(--font-size-sm);
	}

	.remove-option-btn {
		width: 18px;
		height: 18px;
		border: none;
		background: transparent;
		color: var(--color-text-muted);
		cursor: pointer;
		font-size: 14px;
		line-height: 1;
		border-radius: 50%;
	}

	.remove-option-btn:hover {
		background: var(--color-error);
		color: white;
	}

	.add-option-row {
		display: flex;
		gap: 4px;
	}

	.add-option-row input {
		flex: 1;
		padding: 6px 8px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
	}

	.color-select {
		width: 80px;
		padding: 6px 4px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
	}

	.add-option-btn {
		width: 28px;
		height: 28px;
		border: 1px solid var(--color-primary);
		background: var(--color-primary);
		color: white;
		border-radius: var(--radius-sm);
		cursor: pointer;
		font-size: 16px;
		line-height: 1;
	}

	.add-option-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.add-option-btn:hover:not(:disabled) {
		background: var(--color-primary-hover);
	}

	/* Formula/Rollup/Lookup config styles */
	.formula-config,
	.rollup-config,
	.lookup-config {
		margin-bottom: var(--spacing-sm);
	}

	.formula-config label,
	.rollup-config label,
	.lookup-config label {
		display: block;
		font-size: var(--font-size-sm);
		font-weight: 500;
		margin-bottom: var(--spacing-xs);
		margin-top: var(--spacing-sm);
		color: var(--color-text-muted);
	}

	.formula-config label:first-child,
	.rollup-config label:first-child,
	.lookup-config label:first-child {
		margin-top: 0;
	}

	.formula-config textarea {
		width: 100%;
		padding: 8px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
		font-family: monospace;
		resize: vertical;
	}

	.formula-help {
		display: block;
		font-size: 11px;
		color: var(--color-text-muted);
		margin-top: 4px;
	}

	.formula-config select,
	.rollup-config select,
	.lookup-config select,
	.rollup-config input,
	.lookup-config input {
		width: 100%;
		padding: 6px 8px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
	}

	.spacer-cell {
		width: 40px;
		min-width: 40px;
		border: none;
		flex-grow: 1;
	}

	.add-row {
		background: var(--color-gray-50);
	}

	.add-row-btn {
		width: 100%;
		height: 100%;
		border: none;
		background: transparent;
		cursor: pointer;
		font-size: 16px;
		color: var(--color-text-muted);
	}

	.add-row-btn:hover {
		color: var(--color-primary);
		background: var(--color-gray-100);
	}

	.empty-cell {
		background: var(--color-gray-50);
	}

	.empty-grid {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--spacing-xl);
		color: var(--color-text-muted);
	}

	.empty-grid p {
		margin: 0;
	}

	.empty-grid .hint {
		font-size: var(--font-size-sm);
		margin-top: var(--spacing-xs);
	}

	/* Field name editing */
	.field-name-input {
		width: 100%;
		padding: var(--spacing-xs);
		border: 2px solid var(--color-primary);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
		font-weight: 500;
	}

	.header-cell.editing {
		padding: 2px;
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

	/* Color menu styles */
	.color-menu {
		min-width: 160px;
	}

	.context-menu-header {
		padding: var(--spacing-xs) var(--spacing-sm);
		font-size: var(--font-size-xs);
		font-weight: 600;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.color-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 4px;
		padding: var(--spacing-xs) var(--spacing-sm);
	}

	.color-option {
		width: 28px;
		height: 28px;
		border: 2px solid transparent;
		border-radius: var(--radius-sm);
		cursor: pointer;
		transition: transform 0.1s, border-color 0.1s;
	}

	.color-option:hover {
		transform: scale(1.1);
		border-color: var(--color-text-muted);
	}

	/* Row coloring */
	.grid-row.colored-row .data-cell {
		background: var(--row-bg);
	}

	.grid-row.colored-row .data-cell:hover {
		background: var(--row-bg);
		filter: brightness(0.95);
	}

	.grid-row.colored-row .data-cell.selected {
		background: var(--row-bg);
	}

	/* Color indicator in row number */
	.color-indicator {
		position: absolute;
		left: 2px;
		top: 2px;
		bottom: 2px;
		width: 4px;
		border-radius: 2px;
	}

	/* Select field styles */
	.select-cell {
		cursor: pointer;
	}

	.select-cell:hover {
		background: var(--color-primary-light);
	}

	.select-value,
	.multi-select-value {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		padding: 4px 6px;
		align-items: center;
		min-height: 24px;
	}

	.select-pill {
		display: inline-block;
		padding: 2px 8px;
		border-radius: 12px;
		font-size: 12px;
		font-weight: 500;
		max-width: 120px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.select-placeholder {
		color: var(--color-text-muted);
		font-size: 12px;
		font-style: italic;
	}

	.select-more {
		display: inline-block;
		padding: 2px 6px;
		background: var(--color-gray-200);
		color: var(--color-text-muted);
		border-radius: 12px;
		font-size: 11px;
	}

	/* Computed field styles (formula, rollup, lookup) */
	.computed-value {
		color: var(--color-text-muted);
		font-style: italic;
	}

	/* Attachment field styles */
	.attachment-value {
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.attachment-value.clickable {
		cursor: pointer;
	}

	.attachment-value.clickable:hover {
		text-decoration: underline;
	}

	.attachment-count {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 2px 8px;
		background: var(--color-gray-100);
		border-radius: 12px;
		font-size: 12px;
		color: var(--color-text-secondary);
	}

	.attachment-placeholder {
		color: var(--color-primary);
		font-size: 12px;
		cursor: pointer;
	}

	.attachment-placeholder:hover {
		text-decoration: underline;
	}

	/* Computed field modal */
	.computed-field-modal {
		background: white;
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-lg);
		width: 90%;
		max-width: 500px;
		max-height: 90vh;
		overflow: hidden;
	}

	.form-group {
		margin-bottom: 16px;
	}

	.form-group label {
		display: block;
		margin-bottom: 6px;
		font-weight: 500;
		font-size: 14px;
	}

	.form-group input,
	.form-group select,
	.form-group textarea {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: 14px;
	}

	.form-group textarea {
		resize: vertical;
		font-family: monospace;
	}

	.form-help {
		margin-top: 4px;
		font-size: 12px;
		color: var(--color-text-muted);
	}

	/* Attachment modal */
	.attachment-modal {
		background: white;
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-lg);
		width: 90%;
		max-width: 500px;
		max-height: 90vh;
		overflow: hidden;
	}

	.attachment-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.attachment-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px;
		background: var(--color-gray-50);
		border-radius: var(--radius-md);
	}

	.attachment-info {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.attachment-icon {
		font-size: 24px;
	}

	.attachment-details {
		display: flex;
		flex-direction: column;
	}

	.attachment-name {
		font-weight: 500;
		word-break: break-all;
	}

	.attachment-size {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	.attachment-actions {
		display: flex;
		gap: 8px;
	}

	.btn-small {
		padding: 4px 12px;
		background: var(--color-gray-100);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: 12px;
		cursor: pointer;
		text-decoration: none;
		color: inherit;
	}

	.btn-small:hover {
		background: var(--color-gray-200);
	}

	.btn-small.danger {
		color: var(--color-danger);
	}

	.btn-small.danger:hover {
		background: #fee2e2;
	}

	.no-attachments {
		text-align: center;
		color: var(--color-text-muted);
		padding: 24px;
	}

	.upload-section {
		margin-top: 16px;
		padding-top: 16px;
		border-top: 1px solid var(--color-border);
	}

	.upload-btn {
		display: inline-block;
		padding: 8px 16px;
		background: var(--color-primary);
		color: white;
		border-radius: var(--radius-md);
		cursor: pointer;
		font-size: 14px;
	}

	.upload-btn:hover {
		background: var(--color-primary-dark);
	}

	/* Select dropdown */
	.select-dropdown {
		position: fixed;
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-lg);
		min-width: 180px;
		max-width: 300px;
		max-height: 300px;
		overflow-y: auto;
		z-index: 201;
		padding: var(--spacing-xs);
	}

	.select-dropdown-empty {
		padding: 12px;
		text-align: center;
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
	}

	.select-option {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 8px 12px;
		background: none;
		border: none;
		border-radius: var(--radius-sm);
		text-align: left;
		font-size: var(--font-size-sm);
		cursor: pointer;
	}

	.select-option:hover {
		background: var(--color-gray-100);
	}

	.select-option.selected {
		background: var(--color-primary-light);
	}

	.select-option.clear-option {
		border-top: 1px solid var(--color-border);
		margin-top: 4px;
		padding-top: 12px;
		color: var(--color-text-muted);
	}

	.option-checkbox {
		font-size: 14px;
		color: var(--color-primary);
	}

	.option-pill {
		display: inline-block;
		padding: 2px 10px;
		border-radius: 12px;
		font-size: 12px;
	}

	.select-done-btn {
		display: block;
		width: 100%;
		padding: 8px;
		margin-top: 8px;
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
		font-weight: 500;
		cursor: pointer;
	}

	.select-done-btn:hover {
		background: var(--color-primary-hover);
	}

	/* Bulk selection styles */
	.row-checkbox {
		width: 14px;
		height: 14px;
		cursor: pointer;
		accent-color: var(--color-primary);
		margin: 0;
		vertical-align: middle;
	}

	/* In data rows, position checkbox on left */
	.data-row-checkbox {
		position: absolute;
		left: 4px;
		top: 50%;
		transform: translateY(-50%);
		opacity: 0;
		transition: opacity 0.15s;
	}

	.row-number:hover .data-row-checkbox,
	.row-number.row-selected .data-row-checkbox {
		opacity: 1;
	}

	/* Header checkbox is always visible and centered */
	.header-cell.row-number .row-checkbox {
		opacity: 1;
	}

	.row-number.row-selected {
		background: var(--color-primary-light);
	}

	.bulk-actions-bar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-xs) var(--spacing-md);
		background: var(--color-primary);
		color: white;
		flex-shrink: 0;
	}

	.bulk-info {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
	}

	.selected-count {
		font-size: var(--font-size-sm);
		font-weight: 500;
	}

	.bulk-clear-btn {
		padding: 2px 8px;
		background: rgba(255, 255, 255, 0.2);
		border: none;
		border-radius: var(--radius-sm);
		color: white;
		font-size: var(--font-size-sm);
		cursor: pointer;
	}

	.bulk-clear-btn:hover {
		background: rgba(255, 255, 255, 0.3);
	}

	.bulk-actions {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
	}

	.bulk-action-item {
		position: relative;
	}

	.bulk-action-btn {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 12px;
		background: rgba(255, 255, 255, 0.15);
		border: none;
		border-radius: var(--radius-md);
		color: white;
		font-size: var(--font-size-sm);
		cursor: pointer;
	}

	.bulk-action-btn:hover {
		background: rgba(255, 255, 255, 0.25);
	}

	.bulk-action-btn.danger {
		background: rgba(239, 68, 68, 0.8);
	}

	.bulk-action-btn.danger:hover {
		background: rgba(239, 68, 68, 1);
	}

	.color-swatch {
		width: 14px;
		height: 14px;
		background: linear-gradient(135deg, #fee2e2 25%, #dcfce7 25%, #dcfce7 50%, #dbeafe 50%, #dbeafe 75%, #f3e8ff 75%);
		border-radius: 3px;
		border: 1px solid rgba(255, 255, 255, 0.3);
	}

	.bulk-color-menu {
		position: absolute;
		top: 100%;
		right: 0;
		margin-top: 4px;
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		box-shadow: var(--shadow-lg);
		padding: var(--spacing-sm);
		z-index: 100;
	}

	.bulk-color-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 4px;
		margin-bottom: var(--spacing-xs);
	}

	.bulk-color-option {
		width: 28px;
		height: 28px;
		border: 2px solid transparent;
		border-radius: var(--radius-sm);
		cursor: pointer;
	}

	.bulk-color-option:hover {
		transform: scale(1.1);
		border-color: var(--color-text-muted);
	}

	.bulk-color-clear {
		display: block;
		width: 100%;
		padding: 6px;
		background: none;
		border: none;
		border-top: 1px solid var(--color-border);
		font-size: var(--font-size-sm);
		color: var(--color-text-muted);
		cursor: pointer;
		text-align: center;
		margin-top: var(--spacing-xs);
	}

	.bulk-color-clear:hover {
		background: var(--color-gray-100);
		color: var(--color-text);
	}

	/* Options Editor Modal */
	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 300;
	}

	.options-editor-modal {
		background: white;
		border-radius: var(--radius-lg);
		width: 100%;
		max-width: 480px;
		max-height: 80vh;
		display: flex;
		flex-direction: column;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-md) var(--spacing-lg);
		border-bottom: 1px solid var(--color-border);
	}

	.modal-header h3 {
		margin: 0;
		font-size: 16px;
		font-weight: 600;
	}

	.modal-close {
		width: 28px;
		height: 28px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: none;
		border: none;
		border-radius: var(--radius-sm);
		font-size: 20px;
		color: var(--color-text-muted);
		cursor: pointer;
	}

	.modal-close:hover {
		background: var(--color-gray-100);
	}

	.modal-body {
		flex: 1;
		overflow-y: auto;
		padding: var(--spacing-lg);
	}

	.options-list-editor {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-sm);
		margin-bottom: var(--spacing-md);
	}

	.option-row,
	.add-option-row {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
	}

	.option-color-select {
		width: 80px;
		padding: 6px 8px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
		background: white;
	}

	.option-color-preview {
		width: 20px;
		height: 20px;
		border-radius: var(--radius-sm);
		flex-shrink: 0;
	}

	.option-name-input {
		flex: 1;
		padding: 8px 12px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
	}

	.option-name-input:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: 0 0 0 2px var(--color-primary-light);
	}

	.option-remove {
		width: 28px;
		height: 28px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: none;
		border: none;
		border-radius: var(--radius-sm);
		font-size: 18px;
		color: var(--color-text-muted);
		cursor: pointer;
	}

	.option-remove:hover {
		background: var(--color-error);
		color: white;
	}

	.add-option-section {
		padding-top: var(--spacing-md);
		border-top: 1px solid var(--color-border);
	}

	.option-add {
		padding: 8px 16px;
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
		cursor: pointer;
	}

	.option-add:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.option-add:hover:not(:disabled) {
		background: var(--color-primary-hover);
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--spacing-sm);
		padding: var(--spacing-md) var(--spacing-lg);
		border-top: 1px solid var(--color-border);
		background: var(--color-gray-50);
		border-radius: 0 0 var(--radius-lg) var(--radius-lg);
	}

	.btn-cancel {
		padding: 8px 16px;
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
		cursor: pointer;
	}

	.btn-cancel:hover {
		background: var(--color-gray-100);
	}

	.btn-save {
		padding: 8px 20px;
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
		font-weight: 500;
		cursor: pointer;
	}

	.btn-save:hover {
		background: var(--color-primary-hover);
	}
</style>
