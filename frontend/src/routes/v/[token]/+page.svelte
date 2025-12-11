<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { publicViews } from '$lib/api/client';
	import type { PublicView, Field, Record, ViewConfig } from '$lib/types';
	import Grid from '$lib/components/Grid.svelte';
	import Kanban from '$lib/components/Kanban.svelte';
	import Calendar from '$lib/components/Calendar.svelte';
	import Gallery from '$lib/components/Gallery.svelte';

	let publicView: PublicView | null = null;
	let loading = true;
	let error = '';

	$: token = $page.params.token;
	$: viewConfig = publicView?.view?.config || {};
	$: viewType = publicView?.view?.type || 'grid';

	onMount(async () => {
		await loadView();
	});

	async function loadView() {
		loading = true;
		error = '';
		try {
			publicView = await publicViews.get(token);
		} catch (err: any) {
			if (err.status === 404) {
				error = 'This view does not exist or is no longer shared publicly.';
			} else {
				error = err.message || 'Failed to load view';
			}
		} finally {
			loading = false;
		}
	}

	// Apply filters and sorts to records
	function getFilteredSortedRecords(records: Record[], fields: Field[], config: ViewConfig): Record[] {
		let result = [...records];

		// Apply filters
		if (config.filters && config.filters.length > 0) {
			result = result.filter(record => {
				return config.filters!.every(filter => {
					const value = record.values[filter.field_id];
					const filterValue = filter.value;

					switch (filter.operator) {
						case 'equals':
							return String(value) === filterValue;
						case 'not_equals':
							return String(value) !== filterValue;
						case 'contains':
							return String(value || '').toLowerCase().includes(filterValue.toLowerCase());
						case 'not_contains':
							return !String(value || '').toLowerCase().includes(filterValue.toLowerCase());
						case 'is_empty':
							return value === null || value === undefined || value === '';
						case 'is_not_empty':
							return value !== null && value !== undefined && value !== '';
						case 'greater_than':
							return Number(value) > Number(filterValue);
						case 'less_than':
							return Number(value) < Number(filterValue);
						default:
							return true;
					}
				});
			});
		}

		// Apply sorts
		if (config.sorts && config.sorts.length > 0) {
			result.sort((a, b) => {
				for (const sort of config.sorts!) {
					const aVal = a.values[sort.field_id];
					const bVal = b.values[sort.field_id];

					let comparison = 0;
					if (aVal === null || aVal === undefined) comparison = 1;
					else if (bVal === null || bVal === undefined) comparison = -1;
					else if (typeof aVal === 'number' && typeof bVal === 'number') {
						comparison = aVal - bVal;
					} else {
						comparison = String(aVal).localeCompare(String(bVal));
					}

					if (comparison !== 0) {
						return sort.direction === 'desc' ? -comparison : comparison;
					}
				}
				return 0;
			});
		}

		return result;
	}

	$: filteredRecords = publicView
		? getFilteredSortedRecords(publicView.records, publicView.fields, viewConfig)
		: [];
</script>

<svelte:head>
	<title>{publicView?.view?.name || 'Shared View'} | VibeTable</title>
</svelte:head>

<div class="view-page">
	{#if loading}
		<div class="loading-state">
			<div class="spinner"></div>
			<p>Loading view...</p>
		</div>
	{:else if error}
		<div class="error-state">
			<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#dc3545" stroke-width="2">
				<circle cx="12" cy="12" r="10"/>
				<path d="M12 8v4M12 16h.01"/>
			</svg>
			<h2>View Not Available</h2>
			<p>{error}</p>
		</div>
	{:else if publicView}
		<div class="view-container">
			<div class="view-header">
				<div class="view-info">
					<h1>{publicView.view.name}</h1>
					<span class="view-badge">{viewType}</span>
				</div>
				<div class="table-info">
					<span class="table-name">{publicView.table.name}</span>
					<span class="record-count">{filteredRecords.length} record{filteredRecords.length !== 1 ? 's' : ''}</span>
				</div>
			</div>

			<div class="view-content">
				{#if viewType === 'grid'}
					<Grid
						fields={publicView.fields}
						records={filteredRecords}
						readonly={true}
					/>
				{:else if viewType === 'kanban'}
					<Kanban
						fields={publicView.fields}
						records={filteredRecords}
						groupByFieldId={viewConfig.group_by_field_id || ''}
						readonly={true}
					/>
				{:else if viewType === 'calendar'}
					<Calendar
						fields={publicView.fields}
						records={filteredRecords}
						dateFieldId={viewConfig.date_field_id || ''}
						titleFieldId={viewConfig.title_field_id || ''}
						readonly={true}
					/>
				{:else if viewType === 'gallery'}
					<Gallery
						fields={publicView.fields}
						records={filteredRecords}
						titleFieldId={viewConfig.title_field_id || ''}
						coverFieldId={viewConfig.cover_field_id || ''}
						readonly={true}
					/>
				{:else}
					<div class="unsupported-view">
						<p>This view type is not supported for public viewing.</p>
					</div>
				{/if}
			</div>
		</div>
	{/if}

	<div class="powered-by">
		Powered by <a href="/" target="_blank">VibeTable</a>
	</div>
</div>

<style>
	.view-page {
		min-height: 100vh;
		background: #f5f7fa;
		display: flex;
		flex-direction: column;
	}

	.loading-state,
	.error-state {
		text-align: center;
		padding: 60px 40px;
		background: white;
		border-radius: 12px;
		box-shadow: 0 4px 24px rgba(0, 0, 0, 0.1);
		max-width: 400px;
		margin: 40px auto;
	}

	.loading-state p,
	.error-state p {
		color: #666;
		margin-top: 16px;
	}

	.loading-state h2,
	.error-state h2 {
		margin: 16px 0 0;
		font-size: 24px;
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 3px solid #e0e0e0;
		border-top-color: var(--primary-color, #2d7ff9);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
		margin: 0 auto;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.view-container {
		flex: 1;
		display: flex;
		flex-direction: column;
	}

	.view-header {
		background: white;
		border-bottom: 1px solid #e0e0e0;
		padding: 16px 24px;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.view-info {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.view-info h1 {
		margin: 0;
		font-size: 20px;
		color: #333;
	}

	.view-badge {
		padding: 4px 10px;
		background: #e8f0fe;
		color: var(--primary-color, #2d7ff9);
		border-radius: 12px;
		font-size: 12px;
		font-weight: 500;
		text-transform: capitalize;
	}

	.table-info {
		display: flex;
		align-items: center;
		gap: 16px;
		color: #666;
		font-size: 14px;
	}

	.table-name {
		font-weight: 500;
	}

	.record-count {
		padding: 4px 8px;
		background: #f0f0f0;
		border-radius: 4px;
	}

	.view-content {
		flex: 1;
		overflow: auto;
		padding: 16px;
	}

	.unsupported-view {
		text-align: center;
		padding: 60px;
		color: #666;
	}

	.powered-by {
		text-align: center;
		padding: 16px;
		font-size: 12px;
		color: #888;
		background: white;
		border-top: 1px solid #e0e0e0;
	}

	.powered-by a {
		color: var(--primary-color, #2d7ff9);
		text-decoration: none;
	}

	.powered-by a:hover {
		text-decoration: underline;
	}

	/* Ensure view components fill the space properly */
	.view-content :global(.grid-container),
	.view-content :global(.kanban-container),
	.view-content :global(.calendar-container),
	.view-content :global(.gallery-container) {
		height: 100%;
		min-height: calc(100vh - 200px);
	}
</style>
