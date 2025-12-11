<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import type { Field, Record, Table, User } from '$lib/types';
	import { records as recordsApi, fields as fieldsApi } from '$lib/api/client';
	import CommentThread from './CommentThread.svelte';

	export let record: Record;
	export let fields: Field[] = [];
	export let tables: Table[] = [];
	export let readonly: boolean = false;
	export let currentUser: User | null = null;

	type TabType = 'fields' | 'comments';
	let activeTab: TabType = 'fields';
	let commentCount = 0;

	const dispatch = createEventDispatcher<{
		close: void;
		update: { values: { [key: string]: any } };
		delete: void;
	}>();

	// Local copy of values for editing
	let editValues: { [key: string]: any } = {};
	let hasChanges = false;
	let saving = false;

	// Linked record caches
	let linkedRecordCache: Map<string, Map<string, string>> = new Map();
	let linkedRecordPickerOpen: string | null = null;
	let linkedRecordsForPicker: { id: string; title: string }[] = [];
	let linkedPickerSearch = '';

	$: filteredLinkedRecords = linkedPickerSearch
		? linkedRecordsForPicker.filter(r => r.title.toLowerCase().includes(linkedPickerSearch.toLowerCase()))
		: linkedRecordsForPicker;

	onMount(() => {
		// Copy record values to local state
		editValues = { ...record.values };
		loadLinkedRecordTitles();
	});

	async function loadLinkedRecordTitles() {
		const linkedFields = fields.filter(f => f.field_type === 'linked_record' && f.options?.linked_table_id);

		for (const field of linkedFields) {
			const tableId = field.options?.linked_table_id;
			if (!tableId || linkedRecordCache.has(tableId)) continue;

			try {
				const [fieldsResult, recordsResult] = await Promise.all([
					fieldsApi.list(tableId),
					recordsApi.list(tableId)
				]);

				const primaryField = fieldsResult.fields.find(f => f.field_type === 'text') || fieldsResult.fields[0];
				const titleMap = new Map<string, string>();

				for (const rec of recordsResult.records) {
					const title = primaryField ? (rec.values[primaryField.id] || 'Untitled') : 'Untitled';
					titleMap.set(rec.id, String(title));
				}
				linkedRecordCache.set(tableId, titleMap);
				linkedRecordCache = linkedRecordCache;
			} catch (e) {
				console.error('Failed to load linked records:', e);
			}
		}
	}

	function getLinkedRecordTitle(fieldId: string, recordId: string): string {
		const field = fields.find(f => f.id === fieldId);
		if (!field) return recordId;

		const tableId = field.options?.linked_table_id;
		if (!tableId) return recordId;

		const titleMap = linkedRecordCache.get(tableId);
		return titleMap?.get(recordId) || 'Loading...';
	}

	function handleValueChange(fieldId: string, value: any) {
		editValues[fieldId] = value;
		editValues = editValues;
		hasChanges = true;
	}

	function handleCheckboxToggle(fieldId: string) {
		editValues[fieldId] = !editValues[fieldId];
		editValues = editValues;
		hasChanges = true;
	}

	async function openLinkedPicker(field: Field) {
		const tableId = field.options?.linked_table_id;
		if (!tableId) return;

		linkedRecordPickerOpen = field.id;
		linkedPickerSearch = '';

		try {
			const [fieldsResult, recordsResult] = await Promise.all([
				fieldsApi.list(tableId),
				recordsApi.list(tableId)
			]);

			const primaryField = fieldsResult.fields.find(f => f.field_type === 'text') || fieldsResult.fields[0];
			linkedRecordsForPicker = recordsResult.records.map(rec => ({
				id: rec.id,
				title: primaryField ? String(rec.values[primaryField.id] || 'Untitled') : 'Untitled'
			}));
		} catch (e) {
			console.error('Failed to load records for picker:', e);
			linkedRecordsForPicker = [];
		}
	}

	function toggleLinkedRecord(fieldId: string, recordId: string) {
		const current = editValues[fieldId] || [];
		const isSelected = current.includes(recordId);

		if (isSelected) {
			editValues[fieldId] = current.filter((id: string) => id !== recordId);
		} else {
			editValues[fieldId] = [...current, recordId];
		}
		editValues = editValues;
		hasChanges = true;
	}

	function removeSelectOption(fieldId: string, optionId: string) {
		const current = editValues[fieldId] || [];
		handleValueChange(fieldId, current.filter((id: string) => id !== optionId));
	}

	function removeLinkedRecord(fieldId: string, recordId: string) {
		const current = editValues[fieldId] || [];
		handleValueChange(fieldId, current.filter((id: string) => id !== recordId));
	}

	function closeLinkedPicker() {
		linkedRecordPickerOpen = null;
		linkedRecordsForPicker = [];
		linkedPickerSearch = '';
	}

	async function saveChanges() {
		if (!hasChanges || saving) return;

		saving = true;
		try {
			dispatch('update', { values: editValues });
			hasChanges = false;
		} finally {
			saving = false;
		}
	}

	function handleDelete() {
		if (confirm('Are you sure you want to delete this record?')) {
			dispatch('delete');
		}
	}

	function close() {
		if (hasChanges) {
			if (!confirm('You have unsaved changes. Discard them?')) {
				return;
			}
		}
		dispatch('close');
	}

	function formatSelectValue(value: any, field: Field): { id: string; name: string; color?: string }[] {
		if (!value) return [];
		const options = field.options?.options || [];

		if (field.field_type === 'single_select') {
			const opt = options.find(o => o.id === value);
			return opt ? [opt] : [];
		}

		if (field.field_type === 'multi_select' && Array.isArray(value)) {
			return value.map(v => options.find(o => o.id === v)).filter(Boolean) as any[];
		}

		return [];
	}

	function getTableName(tableId: string): string {
		return tables.find(t => t.id === tableId)?.name || 'Unknown table';
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			if (linkedRecordPickerOpen) {
				closeLinkedPicker();
			} else {
				close();
			}
		}
		if ((e.metaKey || e.ctrlKey) && e.key === 's') {
			e.preventDefault();
			saveChanges();
		}
	}

	function handleCommentCountChange(event: CustomEvent<{ count: number }>) {
		commentCount = event.detail.count;
	}
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="modal-overlay" on:click|self={close}>
	<div class="modal">
		<div class="modal-header">
			<h2>Record Details</h2>
			<div class="header-actions">
				{#if !readonly}
					<button class="delete-btn" on:click={handleDelete} title="Delete record">
						<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
						</svg>
					</button>
				{/if}
				<button class="close-btn" on:click={close}>
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M18 6L6 18M6 6l12 12"/>
					</svg>
				</button>
			</div>
		</div>

		<div class="modal-tabs">
			<button
				class="tab"
				class:active={activeTab === 'fields'}
				on:click={() => activeTab = 'fields'}
			>
				Fields
			</button>
			<button
				class="tab"
				class:active={activeTab === 'comments'}
				on:click={() => activeTab = 'comments'}
			>
				Comments {#if commentCount > 0}<span class="tab-badge">{commentCount}</span>{/if}
			</button>
		</div>

		<div class="modal-content">
			{#if activeTab === 'fields'}
				{#each fields as field}
					<div class="field-row">
						<div class="field-label">
							<span class="field-icon">
							{#if field.field_type === 'text'}Aa
							{:else if field.field_type === 'number'}#
							{:else if field.field_type === 'checkbox'}☑
							{:else if field.field_type === 'date'}📅
							{:else if field.field_type === 'single_select'}◉
							{:else if field.field_type === 'multi_select'}☰
							{:else if field.field_type === 'linked_record'}🔗
							{/if}
						</span>
						{field.name}
					</div>
					<div class="field-value">
						{#if field.field_type === 'text'}
							<textarea
								class="text-input"
								value={editValues[field.id] || ''}
								on:input={(e) => handleValueChange(field.id, e.currentTarget.value)}
								disabled={readonly}
								placeholder="Enter text..."
								rows="2"
							></textarea>
						{:else if field.field_type === 'number'}
							<input
								type="number"
								class="number-input"
								value={editValues[field.id] ?? ''}
								on:input={(e) => handleValueChange(field.id, e.currentTarget.value ? Number(e.currentTarget.value) : null)}
								disabled={readonly}
								placeholder="Enter number..."
							/>
						{:else if field.field_type === 'checkbox'}
							<label class="checkbox-wrapper">
								<input
									type="checkbox"
									checked={editValues[field.id] || false}
									on:change={() => handleCheckboxToggle(field.id)}
									disabled={readonly}
								/>
								<span class="checkbox-custom"></span>
							</label>
						{:else if field.field_type === 'date'}
							<input
								type="date"
								class="date-input"
								value={editValues[field.id] || ''}
								on:input={(e) => handleValueChange(field.id, e.currentTarget.value)}
								disabled={readonly}
							/>
						{:else if field.field_type === 'single_select'}
							<select
								class="select-input"
								value={editValues[field.id] || ''}
								on:change={(e) => handleValueChange(field.id, e.currentTarget.value || null)}
								disabled={readonly}
							>
								<option value="">Select an option...</option>
								{#each field.options?.options || [] as opt}
									<option value={opt.id}>{opt.name}</option>
								{/each}
							</select>
						{:else if field.field_type === 'multi_select'}
							<div class="multi-select-wrapper">
								<div class="selected-options">
									{#each formatSelectValue(editValues[field.id], field) as opt}
										<span class="option-pill" style="background: {opt.color || '#e0e0e0'}20; border-color: {opt.color || '#e0e0e0'}">
											{opt.name}
											{#if !readonly}
												<button class="remove-option" on:click={() => removeSelectOption(field.id, opt.id)}>×</button>
											{/if}
										</span>
									{/each}
								</div>
								{#if !readonly}
									<select
										class="add-option-select"
										on:change={(e) => {
											const val = e.currentTarget.value;
											if (val) {
												const current = editValues[field.id] || [];
												if (!current.includes(val)) {
													handleValueChange(field.id, [...current, val]);
												}
												e.currentTarget.value = '';
											}
										}}
									>
										<option value="">Add option...</option>
										{#each (field.options?.options || []).filter(o => !(editValues[field.id] || []).includes(o.id)) as opt}
											<option value={opt.id}>{opt.name}</option>
										{/each}
									</select>
								{/if}
							</div>
						{:else if field.field_type === 'linked_record'}
							<div class="linked-records-wrapper">
								<div class="linked-records">
									{#each editValues[field.id] || [] as linkedId}
										<span class="linked-pill">
											{getLinkedRecordTitle(field.id, linkedId)}
											{#if !readonly}
												<button class="remove-linked" on:click={() => removeLinkedRecord(field.id, linkedId)}>×</button>
											{/if}
										</span>
									{/each}
								</div>
								{#if !readonly}
									<button class="add-linked-btn" on:click={() => openLinkedPicker(field)}>
										+ Link to {getTableName(field.options?.linked_table_id || '')}
									</button>
								{/if}
							</div>
						{/if}
					</div>
				</div>
				{/each}
			{:else if activeTab === 'comments'}
				<CommentThread
					recordId={record.id}
					{currentUser}
					{readonly}
					on:countChange={handleCommentCountChange}
				/>
			{/if}
		</div>

		{#if !readonly && activeTab === 'fields'}
			<div class="modal-footer">
				<div class="save-hint">
					{#if hasChanges}
						<span class="unsaved">Unsaved changes</span>
					{/if}
				</div>
				<div class="footer-actions">
					<button class="cancel-btn" on:click={close}>Cancel</button>
					<button class="save-btn" on:click={saveChanges} disabled={!hasChanges || saving}>
						{saving ? 'Saving...' : 'Save'}
					</button>
				</div>
			</div>
		{/if}
	</div>
</div>

{#if linkedRecordPickerOpen}
	<div class="picker-overlay" on:click|self={closeLinkedPicker}>
		<div class="picker-modal">
			<div class="picker-header">
				<h3>Link records</h3>
				<button class="close-btn" on:click={closeLinkedPicker}>×</button>
			</div>
			<div class="picker-search">
				<input
					type="text"
					placeholder="Search records..."
					bind:value={linkedPickerSearch}
					autofocus
				/>
			</div>
			<div class="picker-content">
				{#each filteredLinkedRecords as rec}
					{@const isSelected = (editValues[linkedRecordPickerOpen] || []).includes(rec.id)}
					<button
						class="picker-item"
						class:selected={isSelected}
						on:click={() => toggleLinkedRecord(linkedRecordPickerOpen || '', rec.id)}
					>
						<span class="picker-checkbox">{isSelected ? '✓' : ''}</span>
						<span class="picker-title">{rec.title}</span>
					</button>
				{/each}
				{#if filteredLinkedRecords.length === 0}
					<div class="picker-empty">No records found</div>
				{/if}
			</div>
			<div class="picker-footer">
				<button class="done-btn" on:click={closeLinkedPicker}>Done</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 200;
	}

	.modal {
		background: white;
		border-radius: var(--radius-lg);
		width: 100%;
		max-width: 640px;
		max-height: 85vh;
		display: flex;
		flex-direction: column;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 20px 24px;
		border-bottom: 1px solid var(--color-border);
	}

	.modal-header h2 {
		margin: 0;
		font-size: 18px;
		font-weight: 600;
	}

	.modal-tabs {
		display: flex;
		gap: 0;
		padding: 0 24px;
		border-bottom: 1px solid var(--color-border);
	}

	.tab {
		padding: 12px 16px;
		background: none;
		border: none;
		border-bottom: 2px solid transparent;
		font-size: 14px;
		font-weight: 500;
		color: var(--color-text-muted);
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 6px;
		margin-bottom: -1px;
	}

	.tab:hover {
		color: var(--color-text);
	}

	.tab.active {
		color: var(--color-primary);
		border-bottom-color: var(--color-primary);
	}

	.tab-badge {
		background: var(--color-primary);
		color: white;
		font-size: 11px;
		padding: 2px 6px;
		border-radius: 10px;
		font-weight: 600;
	}

	.header-actions {
		display: flex;
		gap: 8px;
	}

	.close-btn, .delete-btn {
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: none;
		border: none;
		border-radius: var(--radius-md);
		cursor: pointer;
		color: var(--color-text-muted);
	}

	.close-btn:hover {
		background: var(--color-gray-100);
	}

	.delete-btn:hover {
		background: #fee2e2;
		color: #dc2626;
	}

	.modal-content {
		flex: 1;
		overflow-y: auto;
		padding: 24px;
	}

	.field-row {
		display: grid;
		grid-template-columns: 140px 1fr;
		gap: 16px;
		margin-bottom: 20px;
		align-items: start;
	}

	.field-label {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
		font-weight: 500;
		color: var(--color-text-muted);
		padding-top: 8px;
	}

	.field-icon {
		font-size: 14px;
	}

	.field-value {
		min-height: 36px;
	}

	.text-input {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: 14px;
		resize: vertical;
		min-height: 36px;
	}

	.number-input, .date-input, .select-input {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: 14px;
		height: 36px;
	}

	.text-input:focus, .number-input:focus, .date-input:focus, .select-input:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: 0 0 0 3px var(--color-primary-light);
	}

	.text-input:disabled, .number-input:disabled, .date-input:disabled, .select-input:disabled {
		background: var(--color-gray-50);
		cursor: not-allowed;
	}

	.checkbox-wrapper {
		display: inline-flex;
		align-items: center;
		cursor: pointer;
	}

	.checkbox-wrapper input {
		display: none;
	}

	.checkbox-custom {
		width: 20px;
		height: 20px;
		border: 2px solid var(--color-border);
		border-radius: 4px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.checkbox-wrapper input:checked + .checkbox-custom {
		background: var(--color-primary);
		border-color: var(--color-primary);
	}

	.checkbox-wrapper input:checked + .checkbox-custom::after {
		content: '✓';
		color: white;
		font-size: 12px;
	}

	.multi-select-wrapper {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.selected-options {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	.option-pill {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 4px 10px;
		border-radius: 12px;
		font-size: 13px;
		border: 1px solid;
	}

	.remove-option, .remove-linked {
		background: none;
		border: none;
		padding: 0;
		margin-left: 2px;
		cursor: pointer;
		font-size: 14px;
		color: inherit;
		opacity: 0.6;
	}

	.remove-option:hover, .remove-linked:hover {
		opacity: 1;
	}

	.add-option-select {
		padding: 6px 10px;
		border: 1px dashed var(--color-border);
		border-radius: var(--radius-md);
		background: white;
		font-size: 13px;
		color: var(--color-text-muted);
		cursor: pointer;
	}

	.linked-records-wrapper {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.linked-records {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	.linked-pill {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 4px 10px;
		background: var(--color-primary-light);
		color: var(--color-primary);
		border-radius: 12px;
		font-size: 13px;
	}

	.add-linked-btn {
		padding: 6px 12px;
		border: 1px dashed var(--color-primary);
		border-radius: var(--radius-md);
		background: white;
		color: var(--color-primary);
		font-size: 13px;
		cursor: pointer;
		text-align: left;
	}

	.add-linked-btn:hover {
		background: var(--color-primary-light);
	}

	.modal-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 16px 24px;
		border-top: 1px solid var(--color-border);
		background: var(--color-gray-50);
	}

	.save-hint {
		font-size: 13px;
	}

	.unsaved {
		color: #f59e0b;
	}

	.footer-actions {
		display: flex;
		gap: 8px;
	}

	.cancel-btn {
		padding: 8px 16px;
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: 14px;
		cursor: pointer;
	}

	.cancel-btn:hover {
		background: var(--color-gray-50);
	}

	.save-btn {
		padding: 8px 20px;
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-md);
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
	}

	.save-btn:hover:not(:disabled) {
		background: var(--color-primary-hover);
	}

	.save-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	/* Linked record picker */
	.picker-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.3);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 300;
	}

	.picker-modal {
		background: white;
		border-radius: var(--radius-lg);
		width: 100%;
		max-width: 400px;
		max-height: 60vh;
		display: flex;
		flex-direction: column;
		box-shadow: var(--shadow-lg);
	}

	.picker-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 16px 20px;
		border-bottom: 1px solid var(--color-border);
	}

	.picker-header h3 {
		margin: 0;
		font-size: 16px;
	}

	.picker-search {
		padding: 12px 16px;
		border-bottom: 1px solid var(--color-border);
	}

	.picker-search input {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: 14px;
	}

	.picker-search input:focus {
		outline: none;
		border-color: var(--color-primary);
	}

	.picker-content {
		flex: 1;
		overflow-y: auto;
	}

	.picker-item {
		display: flex;
		align-items: center;
		gap: 10px;
		width: 100%;
		padding: 10px 16px;
		background: none;
		border: none;
		text-align: left;
		cursor: pointer;
		font-size: 14px;
	}

	.picker-item:hover {
		background: var(--color-gray-50);
	}

	.picker-item.selected {
		background: var(--color-primary-light);
	}

	.picker-checkbox {
		width: 18px;
		height: 18px;
		border: 1px solid var(--color-border);
		border-radius: 3px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 11px;
		color: var(--color-primary);
	}

	.picker-item.selected .picker-checkbox {
		background: var(--color-primary);
		border-color: var(--color-primary);
		color: white;
	}

	.picker-empty {
		padding: 24px;
		text-align: center;
		color: var(--color-text-muted);
	}

	.picker-footer {
		padding: 12px 16px;
		border-top: 1px solid var(--color-border);
		display: flex;
		justify-content: flex-end;
	}

	.done-btn {
		padding: 8px 16px;
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-md);
		font-size: 14px;
		cursor: pointer;
	}

	.done-btn:hover {
		background: var(--color-primary-hover);
	}
</style>
