<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Field, Record, SelectOption } from '$lib/types';
	import { fields as fieldsApi, records as recordsApi } from '$lib/api/client';

	export let fields: Field[] = [];
	export let records: Record[] = [];
	export let readonly: boolean = false;

	const dispatch = createEventDispatcher<{
		updateRecord: { id: string; values: { [key: string]: any } };
		addRecordWithValue: { fieldId: string; value: any };
		renameOption: { fieldId: string; optionId: string; newName: string };
		selectRecord: { id: string };
		reorderRecords: { recordIds: string[] };
	}>();

	// Column rename state
	let editingColumnId: string | null = null;
	let editingColumnName: string = '';

	// Cache for linked record titles
	let linkedRecordTitles: { [tableId: string]: { [recordId: string]: string } } = {};
	let loadingLinkedRecords = false;
	let linkedCacheVersion = 0;

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
			const tableIdsToFetch: string[] = [];
			for (const field of linkedRecordFields) {
				const tableId = field.options?.linked_table_id;
				if (tableId && !linkedRecordTitles[tableId]) {
					tableIdsToFetch.push(tableId);
				}
			}

			if (tableIdsToFetch.length === 0) return;

			const newTitles: { [tableId: string]: { [recordId: string]: string } } = {};

			await Promise.all(tableIdsToFetch.map(async (tableId) => {
				try {
					const [fieldsResult, recordsResult] = await Promise.all([
						fieldsApi.list(tableId),
						recordsApi.list(tableId)
					]);

					const primaryField = fieldsResult.fields.find((f: Field) => f.field_type === 'text') || fieldsResult.fields[0];

					const titleMap: { [recordId: string]: string } = {};
					for (const record of recordsResult.records) {
						const title = primaryField ? (record.values[primaryField.id] || 'Untitled') : 'Untitled';
						titleMap[record.id] = String(title);
					}
					newTitles[tableId] = titleMap;
				} catch (e) {
					console.error(`Failed to load linked records for table ${tableId}:`, e);
					newTitles[tableId] = {};
				}
			}));

			linkedRecordTitles = { ...linkedRecordTitles, ...newTitles };
			linkedCacheVersion += 1;
		} finally {
			loadingLinkedRecords = false;
		}
	}

	function getLinkedRecordDisplay(value: any, field: Field): string {
		if (!value || !Array.isArray(value) || value.length === 0) return '';

		const tableId = field.options?.linked_table_id;
		if (!tableId) return '';

		const titleMap = linkedRecordTitles[tableId];
		if (!titleMap) return 'Loading...';

		const titles = value.map((id: string) => titleMap[id] || 'Unknown');
		if (titles.length <= 2) return titles.join(', ');
		return `${titles[0]}, +${titles.length - 1} more`;
	}

	// Find single_select fields that can be used for grouping
	$: selectFields = fields.filter(f => f.field_type === 'single_select');

	// Selected grouping field
	let groupByFieldId: string = '';
	$: {
		if (!groupByFieldId && selectFields.length > 0) {
			groupByFieldId = selectFields[0].id;
		}
	}

	// Get the grouping field
	$: groupByField = fields.find(f => f.id === groupByFieldId);

	// Get options from the grouping field
	$: groupOptions = (groupByField?.options?.options || []) as SelectOption[];

	// Get the primary text field (first text field)
	$: primaryTextField = fields.find(f => f.field_type === 'text');

	// Group records by the selected field
	$: groupedRecords = groupRecordsByField(records, groupByFieldId, groupOptions);

	function groupRecordsByField(
		records: Record[],
		fieldId: string,
		options: SelectOption[]
	): Map<string, Record[]> {
		const groups = new Map<string, Record[]>();

		// Initialize groups for each option
		for (const option of options) {
			groups.set(option.id, []);
		}
		// Add an "uncategorized" group for records without a value
		groups.set('__uncategorized__', []);

		for (const record of records) {
			const value = record.values[fieldId];
			if (value && groups.has(value)) {
				groups.get(value)!.push(record);
			} else {
				groups.get('__uncategorized__')!.push(record);
			}
		}

		return groups;
	}

	function getOptionColor(optionId: string): string {
		const option = groupOptions.find(o => o.id === optionId);
		return option?.color || '#e5e7eb';
	}

	function getOptionName(optionId: string): string {
		if (optionId === '__uncategorized__') return 'Uncategorized';
		const option = groupOptions.find(o => o.id === optionId);
		return option?.name || 'Unknown';
	}

	function getRecordTitle(record: Record): string {
		if (primaryTextField) {
			return record.values[primaryTextField.id] || 'Untitled';
		}
		// Fallback to first non-empty value
		for (const field of fields) {
			const value = record.values[field.id];
			if (value && typeof value === 'string') {
				return value;
			}
		}
		return 'Untitled';
	}

	function getRecordPreview(record: Record): { field: Field; value: any }[] {
		const preview: { field: Field; value: any }[] = [];
		for (const field of fields) {
			if (field.id === groupByFieldId) continue; // Skip the grouping field
			if (field.id === primaryTextField?.id) continue; // Skip the title field
			const value = record.values[field.id];
			if (value !== null && value !== undefined && value !== '') {
				preview.push({ field, value });
				if (preview.length >= 2) break; // Only show first 2 fields
			}
		}
		return preview;
	}

	function formatPreviewValue(field: Field, value: any, _cacheVersion?: number): string {
		switch (field.field_type) {
			case 'checkbox':
				return value ? '✓' : '✗';
			case 'date':
				return new Date(value).toLocaleDateString();
			case 'number':
				return String(value);
			case 'linked_record':
				return getLinkedRecordDisplay(value, field);
			case 'single_select':
				const option = field.options?.options?.find((o: SelectOption) => o.id === value);
				return option?.name || String(value);
			case 'multi_select':
				if (!Array.isArray(value)) return '';
				return value.map((id: string) => {
					const opt = field.options?.options?.find((o: SelectOption) => o.id === id);
					return opt?.name || id;
				}).join(', ');
			default:
				return String(value).substring(0, 50);
		}
	}

	function moveRecord(recordId: string, newOptionId: string) {
		if (readonly) return;
		const actualValue = newOptionId === '__uncategorized__' ? null : newOptionId;
		dispatch('updateRecord', {
			id: recordId,
			values: { [groupByFieldId]: actualValue }
		});
	}

	function addRecordToColumn(optionId: string) {
		if (readonly) return;
		const value = optionId === '__uncategorized__' ? null : optionId;
		dispatch('addRecordWithValue', {
			fieldId: groupByFieldId,
			value
		});
	}

	// Drag and drop handling
	let draggedRecordId: string | null = null;
	let dragOverColumn: string | null = null;
	let dragOverRecordId: string | null = null;
	let dragOverPosition: 'above' | 'below' | null = null;

	function handleDragStart(e: DragEvent, recordId: string) {
		if (readonly) return;
		draggedRecordId = recordId;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
		}
	}

	function handleDragOver(e: DragEvent, columnId: string) {
		e.preventDefault();
		dragOverColumn = columnId;
	}

	function handleCardDragOver(e: DragEvent, recordId: string, columnId: string) {
		if (!draggedRecordId || draggedRecordId === recordId) return;
		e.preventDefault();
		e.stopPropagation();

		dragOverColumn = columnId;
		dragOverRecordId = recordId;

		// Determine if we're above or below the center of the card
		const rect = (e.currentTarget as HTMLElement).getBoundingClientRect();
		const y = e.clientY - rect.top;
		dragOverPosition = y < rect.height / 2 ? 'above' : 'below';
	}

	function handleCardDragLeave() {
		dragOverRecordId = null;
		dragOverPosition = null;
	}

	function handleDragLeave() {
		dragOverColumn = null;
	}

	function handleDrop(e: DragEvent, columnId: string) {
		e.preventDefault();
		if (draggedRecordId) {
			// Find the current column of the dragged record
			const record = records.find(r => r.id === draggedRecordId);
			const currentColumnId = record?.values[groupByFieldId] || '__uncategorized__';

			if (currentColumnId !== columnId) {
				// Moving to a different column - just update the column value
				moveRecord(draggedRecordId, columnId);
			} else if (dragOverRecordId && dragOverRecordId !== draggedRecordId) {
				// Reordering within the same column
				const columnRecords = groupedRecords.get(columnId) || [];
				const newOrder = [...columnRecords.map(r => r.id)];

				const draggedIndex = newOrder.indexOf(draggedRecordId);
				const targetIndex = newOrder.indexOf(dragOverRecordId);

				if (draggedIndex !== -1 && targetIndex !== -1) {
					// Remove from old position
					newOrder.splice(draggedIndex, 1);
					// Insert at new position
					const insertIndex = dragOverPosition === 'above' ? targetIndex : targetIndex + 1;
					const adjustedIndex = draggedIndex < targetIndex ? insertIndex - 1 : insertIndex;
					newOrder.splice(adjustedIndex, 0, draggedRecordId);

					// Emit reorder event for this column's records
					dispatch('reorderRecords', { recordIds: newOrder });
				}
			}
		}
		draggedRecordId = null;
		dragOverColumn = null;
		dragOverRecordId = null;
		dragOverPosition = null;
	}

	function handleDragEnd() {
		draggedRecordId = null;
		dragOverColumn = null;
		dragOverRecordId = null;
		dragOverPosition = null;
	}

	// Column rename functions
	function startColumnRename(optionId: string) {
		if (readonly || optionId === '__uncategorized__') return;
		editingColumnId = optionId;
		editingColumnName = getOptionName(optionId);
	}

	function saveColumnRename() {
		if (!editingColumnId || !editingColumnName.trim()) {
			cancelColumnRename();
			return;
		}
		dispatch('renameOption', {
			fieldId: groupByFieldId,
			optionId: editingColumnId,
			newName: editingColumnName.trim()
		});
		editingColumnId = null;
		editingColumnName = '';
	}

	function cancelColumnRename() {
		editingColumnId = null;
		editingColumnName = '';
	}

	function handleColumnRenameKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			saveColumnRename();
		} else if (e.key === 'Escape') {
			cancelColumnRename();
		}
	}
</script>

<div class="kanban-container">
	{#if selectFields.length === 0}
		<div class="no-select-field">
			<p>No single-select fields available</p>
			<p class="hint">Add a single-select field to use Kanban view</p>
		</div>
	{:else}
		<div class="kanban-toolbar">
			<label class="group-by-label">
				Group by:
				<select bind:value={groupByFieldId}>
					{#each selectFields as field}
						<option value={field.id}>{field.name}</option>
					{/each}
				</select>
			</label>
		</div>

		<div class="kanban-board">
			{#each [...groupedRecords] as [optionId, columnRecords]}
				<div
					class="kanban-column"
					class:drag-over={dragOverColumn === optionId}
					on:dragover={(e) => handleDragOver(e, optionId)}
					on:dragleave={handleDragLeave}
					on:drop={(e) => handleDrop(e, optionId)}
				>
					<div class="column-header" style="border-top-color: {getOptionColor(optionId)}">
						{#if editingColumnId === optionId}
							<input
								type="text"
								class="column-title-input"
								bind:value={editingColumnName}
								on:blur={saveColumnRename}
								on:keydown={handleColumnRenameKeydown}
								autofocus
							/>
						{:else}
							<span
								class="column-title"
								class:editable={!readonly && optionId !== '__uncategorized__'}
								on:dblclick={() => startColumnRename(optionId)}
								title={!readonly && optionId !== '__uncategorized__' ? 'Double-click to rename' : ''}
							>{getOptionName(optionId)}</span>
						{/if}
						<span class="column-count">{columnRecords.length}</span>
					</div>

					<div class="column-cards">
						{#each columnRecords as record}
							<div
								class="kanban-card"
								class:dragging={draggedRecordId === record.id}
								class:drag-over-above={dragOverRecordId === record.id && dragOverPosition === 'above'}
								class:drag-over-below={dragOverRecordId === record.id && dragOverPosition === 'below'}
								draggable={!readonly}
								on:dragstart={(e) => handleDragStart(e, record.id)}
								on:dragover={(e) => handleCardDragOver(e, record.id, optionId)}
								on:dragleave={handleCardDragLeave}
								on:dragend={handleDragEnd}
								on:click={() => dispatch('selectRecord', { id: record.id })}
							>
								<div class="card-title">{getRecordTitle(record)}</div>
								{#each getRecordPreview(record) as preview}
									<div class="card-field">
										<span class="field-name">{preview.field.name}:</span>
										<span class="field-value">{formatPreviewValue(preview.field, preview.value, linkedCacheVersion)}</span>
									</div>
								{/each}
							</div>
						{/each}
					</div>

					{#if !readonly}
						<button class="add-card-btn" on:click={() => addRecordToColumn(optionId)}>
							+ Add card
						</button>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.kanban-container {
		flex: 1;
		display: flex;
		flex-direction: column;
		background: var(--color-gray-50);
		overflow: hidden;
	}

	.no-select-field {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		flex: 1;
		color: var(--color-text-muted);
	}

	.no-select-field p {
		margin: 0;
	}

	.hint {
		font-size: var(--font-size-sm);
		margin-top: var(--spacing-xs);
	}

	.kanban-toolbar {
		padding: var(--spacing-sm) var(--spacing-md);
		background: white;
		border-bottom: 1px solid var(--color-border);
		flex-shrink: 0;
	}

	.group-by-label {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		font-size: var(--font-size-sm);
		color: var(--color-text-muted);
	}

	.group-by-label select {
		padding: var(--spacing-xs) var(--spacing-sm);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: var(--font-size-sm);
		background: white;
	}

	.kanban-board {
		display: flex;
		gap: var(--spacing-md);
		padding: var(--spacing-md);
		overflow-x: auto;
		flex: 1;
		align-items: flex-start;
	}

	.kanban-column {
		min-width: 280px;
		max-width: 280px;
		background: var(--color-gray-100);
		border-radius: var(--radius-lg);
		display: flex;
		flex-direction: column;
		max-height: 100%;
	}

	.kanban-column.drag-over {
		background: var(--color-primary-light);
	}

	.column-header {
		padding: var(--spacing-sm) var(--spacing-md);
		border-top: 3px solid var(--color-gray-400);
		border-radius: var(--radius-lg) var(--radius-lg) 0 0;
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-shrink: 0;
	}

	.column-title {
		font-weight: 600;
		font-size: var(--font-size-sm);
	}

	.column-title.editable {
		cursor: pointer;
		padding: 2px 4px;
		border-radius: var(--radius-sm);
	}

	.column-title.editable:hover {
		background: rgba(0, 0, 0, 0.05);
	}

	.column-title-input {
		font-weight: 600;
		font-size: var(--font-size-sm);
		padding: 2px 4px;
		border: 1px solid var(--color-primary);
		border-radius: var(--radius-sm);
		outline: none;
		width: 100%;
		max-width: 180px;
		background: white;
	}

	.column-count {
		background: var(--color-gray-200);
		padding: 2px var(--spacing-xs);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-xs);
		color: var(--color-text-muted);
	}

	.column-cards {
		flex: 1;
		overflow-y: auto;
		padding: var(--spacing-xs) var(--spacing-sm);
		display: flex;
		flex-direction: column;
		gap: var(--spacing-xs);
	}

	.kanban-card {
		background: white;
		border-radius: var(--radius-md);
		padding: var(--spacing-sm);
		box-shadow: var(--shadow-sm);
		cursor: grab;
		transition: box-shadow 0.15s, opacity 0.15s;
	}

	.kanban-card:hover {
		box-shadow: var(--shadow-md);
	}

	.kanban-card.dragging {
		opacity: 0.5;
	}

	.kanban-card.drag-over-above {
		border-top: 3px solid var(--color-primary);
		margin-top: -3px;
	}

	.kanban-card.drag-over-below {
		border-bottom: 3px solid var(--color-primary);
		margin-bottom: -3px;
	}

	.card-title {
		font-weight: 500;
		font-size: var(--font-size-sm);
		margin-bottom: var(--spacing-xs);
	}

	.card-field {
		font-size: var(--font-size-xs);
		color: var(--color-text-muted);
		display: flex;
		gap: var(--spacing-xs);
	}

	.field-name {
		font-weight: 500;
	}

	.add-card-btn {
		margin: var(--spacing-xs) var(--spacing-sm) var(--spacing-sm);
		padding: var(--spacing-sm);
		background: transparent;
		border: 1px dashed var(--color-border);
		border-radius: var(--radius-md);
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
		cursor: pointer;
		text-align: center;
	}

	.add-card-btn:hover {
		background: white;
		border-color: var(--color-primary);
		color: var(--color-primary);
	}
</style>
