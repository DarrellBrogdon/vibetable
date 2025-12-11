<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Field, Record } from '$lib/types';

	export let fields: Field[] = [];
	export let records: Record[] = [];
	export let titleFieldId: string = '';
	export let coverFieldId: string = '';
	export let readonly: boolean = false;

	const dispatch = createEventDispatcher<{
		addRecord: void;
		selectRecord: { id: string };
	}>();

	// Get title field if not specified (first text field)
	$: effectiveTitleFieldId = titleFieldId || fields.find(f => f.field_type === 'text')?.id || '';

	// Get the display fields (exclude title and cover if set)
	$: displayFields = fields.filter(f =>
		f.id !== effectiveTitleFieldId &&
		f.id !== coverFieldId &&
		f.field_type !== 'linked_record' // Don't display linked records in cards
	).slice(0, 4); // Show max 4 fields per card

	function getRecordTitle(record: Record): string {
		if (effectiveTitleFieldId && record.values[effectiveTitleFieldId]) {
			return String(record.values[effectiveTitleFieldId]);
		}
		// Fall back to first non-empty text value
		for (const field of fields) {
			if (field.field_type === 'text' && record.values[field.id]) {
				return String(record.values[field.id]);
			}
		}
		return 'Untitled';
	}

	// Reactive map of record id to cover URL - recomputes when records or coverFieldId changes
	$: coverImages = new Map<string, string | null>(
		records.map(record => {
			if (!coverFieldId || !record.values[coverFieldId]) {
				return [record.id, null];
			}
			const value = String(record.values[coverFieldId]);
			// Check if it looks like a URL
			if (value.startsWith('http://') || value.startsWith('https://')) {
				return [record.id, value];
			}
			// Also support // protocol-relative URLs
			if (value.startsWith('//')) {
				return [record.id, 'https:' + value];
			}
			return [record.id, null];
		})
	);

	function getCoverImage(recordId: string): string | null {
		return coverImages.get(recordId) || null;
	}

	// Track images that failed to load
	let failedImages = new Set<string>();

	function handleImageError(recordId: string) {
		failedImages.add(recordId);
		failedImages = failedImages; // Trigger reactivity
	}

	function formatFieldValue(field: Field, value: any): string {
		if (value === null || value === undefined || value === '') return '—';

		switch (field.field_type) {
			case 'checkbox':
				return value ? '✓' : '✗';
			case 'date':
				try {
					return new Date(value).toLocaleDateString();
				} catch {
					return String(value);
				}
			case 'number':
				if (typeof value === 'number') {
					const precision = field.options?.precision ?? 0;
					return value.toFixed(precision);
				}
				return String(value);
			case 'single_select':
				const option = field.options?.options?.find((o: any) => o.id === value);
				return option?.name || String(value);
			case 'multi_select':
				if (Array.isArray(value)) {
					return value
						.map(v => field.options?.options?.find((o: any) => o.id === v)?.name || v)
						.join(', ');
				}
				return String(value);
			default:
				return String(value);
		}
	}

	function getSelectColor(field: Field, value: any): string | null {
		if (field.field_type !== 'single_select') return null;
		const option = field.options?.options?.find((o: any) => o.id === value);
		return option?.color || null;
	}

	function handleCardClick(record: Record) {
		dispatch('selectRecord', { id: record.id });
	}

	function handleAddRecord() {
		if (readonly) return;
		dispatch('addRecord');
	}
</script>

<div class="gallery-view">
	{#if fields.length === 0}
		<div class="empty-state">
			<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
				<rect x="3" y="3" width="7" height="7"/>
				<rect x="14" y="3" width="7" height="7"/>
				<rect x="3" y="14" width="7" height="7"/>
				<rect x="14" y="14" width="7" height="7"/>
			</svg>
			<p>Add fields to display records in Gallery view</p>
		</div>
	{:else}
		<div class="gallery-grid">
			{#each records as record (record.id)}
				{@const coverUrl = getCoverImage(record.id)}
				<button
					class="gallery-card"
					class:has-color={record.color}
					style:border-left-color={record.color ? `var(--color-${record.color})` : undefined}
					on:click={() => handleCardClick(record)}
				>
					{#if coverUrl && !failedImages.has(record.id)}
						<div class="card-cover">
							<img
								src={coverUrl}
								alt=""
								loading="lazy"
								on:error={() => handleImageError(record.id)}
							/>
						</div>
					{:else}
						<div class="card-cover placeholder">
							<svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
								<rect x="3" y="3" width="18" height="18" rx="2"/>
								<circle cx="8.5" cy="8.5" r="1.5"/>
								<path d="M21 15l-5-5L5 21"/>
							</svg>
						</div>
					{/if}
					<div class="card-content">
						<h3 class="card-title">{getRecordTitle(record)}</h3>
						<div class="card-fields">
							{#each displayFields as field}
								{@const value = record.values[field.id]}
								{@const selectColor = getSelectColor(field, value)}
								<div class="card-field">
									<span class="field-name">{field.name}</span>
									<span
										class="field-value"
										class:select-value={field.field_type === 'single_select' && selectColor}
										style:background-color={selectColor || undefined}
									>
										{formatFieldValue(field, value)}
									</span>
								</div>
							{/each}
						</div>
					</div>
				</button>
			{/each}

			{#if !readonly}
				<button class="add-card" on:click={handleAddRecord}>
					<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="12" y1="5" x2="12" y2="19"/>
						<line x1="5" y1="12" x2="19" y2="12"/>
					</svg>
					<span>Add record</span>
				</button>
			{/if}
		</div>

		{#if records.length === 0}
			<div class="no-records">
				<p>No records yet</p>
				{#if !readonly}
					<button class="add-first-btn" on:click={handleAddRecord}>
						Add your first record
					</button>
				{/if}
			</div>
		{/if}
	{/if}
</div>

<style>
	.gallery-view {
		height: 100%;
		overflow-y: auto;
		padding: 20px;
		background: #f9f9f9;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 100%;
		color: #888;
		gap: 16px;
	}

	.empty-state p {
		font-size: 16px;
		margin: 0;
	}

	.gallery-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 20px;
	}

	.gallery-card {
		background: white;
		border-radius: 8px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		overflow: hidden;
		cursor: pointer;
		transition: box-shadow 0.2s, transform 0.2s;
		border: 1px solid #e0e0e0;
		border-left: 4px solid transparent;
		text-align: left;
		padding: 0;
	}

	.gallery-card:hover {
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
		transform: translateY(-2px);
	}

	.gallery-card.has-color {
		border-left-width: 4px;
	}

	.card-cover {
		height: 140px;
		background: #f5f5f5;
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}

	.card-cover img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.card-cover.placeholder {
		color: #ccc;
	}

	.card-content {
		padding: 16px;
	}

	.card-title {
		font-size: 16px;
		font-weight: 600;
		color: #333;
		margin: 0 0 12px 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.card-fields {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.card-field {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 12px;
	}

	.field-name {
		font-size: 12px;
		color: #888;
		flex-shrink: 0;
	}

	.field-value {
		font-size: 13px;
		color: #333;
		text-align: right;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.field-value.select-value {
		padding: 2px 8px;
		border-radius: 12px;
		font-size: 12px;
	}

	.add-card {
		background: white;
		border: 2px dashed #ddd;
		border-radius: 8px;
		min-height: 200px;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 8px;
		cursor: pointer;
		color: #888;
		transition: all 0.2s;
	}

	.add-card:hover {
		border-color: var(--primary-color, #2d7ff9);
		color: var(--primary-color, #2d7ff9);
		background: #f8fbff;
	}

	.add-card span {
		font-size: 14px;
		font-weight: 500;
	}

	.no-records {
		text-align: center;
		padding: 60px 20px;
		color: #888;
	}

	.no-records p {
		margin: 0 0 16px 0;
		font-size: 16px;
	}

	.add-first-btn {
		background: var(--primary-color, #2d7ff9);
		color: white;
		border: none;
		border-radius: 6px;
		padding: 10px 20px;
		font-size: 14px;
		font-weight: 500;
		cursor: pointer;
	}

	.add-first-btn:hover {
		background: #1a6fe8;
	}

	/* Color variables */
	:global(:root) {
		--color-red: #ef4444;
		--color-orange: #f97316;
		--color-yellow: #eab308;
		--color-green: #22c55e;
		--color-blue: #3b82f6;
		--color-purple: #a855f7;
		--color-pink: #ec4899;
		--color-gray: #6b7280;
	}
</style>
