<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { tables as tablesApi, fields as fieldsApi, records as recordsApi } from '$lib/api/client';
	import type { Table, Field, Record } from '$lib/types';

	export let tableId: string;
	export let selectedIds: string[] = [];
	export let multiple: boolean = true;

	const dispatch = createEventDispatcher();

	let table: Table | null = null;
	let fields: Field[] = [];
	let records: Record[] = [];
	let loading = true;
	let searchQuery = '';

	$: primaryField = fields.find(f => f.field_type === 'text') || fields[0];
	$: filteredRecords = searchQuery
		? records.filter(r => {
				const primaryValue = primaryField ? r.values[primaryField.id] : '';
				return String(primaryValue || '').toLowerCase().includes(searchQuery.toLowerCase());
			})
		: records;

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		loading = true;
		try {
			const [tableResult, fieldsResult, recordsResult] = await Promise.all([
				tablesApi.get(tableId),
				fieldsApi.list(tableId),
				recordsApi.list(tableId)
			]);
			table = tableResult;
			fields = fieldsResult.fields;
			records = recordsResult.records;
		} catch (e) {
			console.error('Failed to load linked table data:', e);
		} finally {
			loading = false;
		}
	}

	function getRecordTitle(record: Record): string {
		if (primaryField) {
			return record.values[primaryField.id] || 'Untitled';
		}
		return 'Untitled';
	}

	function isSelected(recordId: string): boolean {
		return selectedIds.includes(recordId);
	}

	function toggleRecord(recordId: string) {
		if (multiple) {
			if (isSelected(recordId)) {
				dispatch('change', selectedIds.filter(id => id !== recordId));
			} else {
				dispatch('change', [...selectedIds, recordId]);
			}
		} else {
			dispatch('change', isSelected(recordId) ? [] : [recordId]);
		}
	}

	function close() {
		dispatch('close');
	}
</script>

<div class="picker-overlay" on:click={close}>
	<div class="picker-modal" on:click|stopPropagation>
		<div class="picker-header">
			<h3>Link to {table?.name || 'records'}</h3>
			<button class="close-btn" on:click={close}>×</button>
		</div>

		<div class="picker-search">
			<input
				type="text"
				placeholder="Search records..."
				bind:value={searchQuery}
				autofocus
			/>
		</div>

		<div class="picker-content">
			{#if loading}
				<div class="loading">Loading records...</div>
			{:else if filteredRecords.length === 0}
				<div class="empty">
					{searchQuery ? 'No matching records' : 'No records in this table'}
				</div>
			{:else}
				<ul class="record-list">
					{#each filteredRecords as record}
						<li>
							<button
								class="record-item"
								class:selected={isSelected(record.id)}
								on:click={() => toggleRecord(record.id)}
							>
								<span class="checkbox">
									{isSelected(record.id) ? '✓' : ''}
								</span>
								<span class="record-title">{getRecordTitle(record)}</span>
							</button>
						</li>
					{/each}
				</ul>
			{/if}
		</div>

		<div class="picker-footer">
			<span class="selection-count">
				{selectedIds.length} selected
			</span>
			<button class="done-btn" on:click={close}>Done</button>
		</div>
	</div>
</div>

<style>
	.picker-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 200;
	}

	.picker-modal {
		background: white;
		border-radius: var(--radius-lg);
		width: 100%;
		max-width: 480px;
		max-height: 80vh;
		display: flex;
		flex-direction: column;
		box-shadow: var(--shadow-lg);
	}

	.picker-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-md) var(--spacing-lg);
		border-bottom: 1px solid var(--color-border);
	}

	.picker-header h3 {
		margin: 0;
		font-size: var(--font-size-lg);
	}

	.close-btn {
		width: 32px;
		height: 32px;
		background: none;
		border: none;
		font-size: 1.5rem;
		color: var(--color-text-muted);
		cursor: pointer;
		border-radius: var(--radius-md);
	}

	.close-btn:hover {
		background: var(--color-gray-100);
	}

	.picker-search {
		padding: var(--spacing-sm) var(--spacing-lg);
		border-bottom: 1px solid var(--color-border);
	}

	.picker-search input {
		width: 100%;
		padding: var(--spacing-sm) var(--spacing-md);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: var(--font-size-sm);
	}

	.picker-search input:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: 0 0 0 3px var(--color-primary-light);
	}

	.picker-content {
		flex: 1;
		overflow-y: auto;
		padding: var(--spacing-sm) 0;
	}

	.loading, .empty {
		padding: var(--spacing-xl);
		text-align: center;
		color: var(--color-text-muted);
	}

	.record-list {
		list-style: none;
		margin: 0;
		padding: 0;
	}

	.record-item {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		width: 100%;
		padding: var(--spacing-sm) var(--spacing-lg);
		background: none;
		border: none;
		text-align: left;
		cursor: pointer;
		font-size: var(--font-size-sm);
	}

	.record-item:hover {
		background: var(--color-gray-50);
	}

	.record-item.selected {
		background: var(--color-primary-light);
	}

	.checkbox {
		width: 20px;
		height: 20px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 12px;
		color: var(--color-primary);
		flex-shrink: 0;
	}

	.record-item.selected .checkbox {
		background: var(--color-primary);
		border-color: var(--color-primary);
		color: white;
	}

	.record-title {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.picker-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-md) var(--spacing-lg);
		border-top: 1px solid var(--color-border);
	}

	.selection-count {
		font-size: var(--font-size-sm);
		color: var(--color-text-muted);
	}

	.done-btn {
		padding: var(--spacing-sm) var(--spacing-lg);
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-md);
		font-weight: 500;
		cursor: pointer;
	}

	.done-btn:hover {
		background: var(--color-primary-hover);
	}
</style>
