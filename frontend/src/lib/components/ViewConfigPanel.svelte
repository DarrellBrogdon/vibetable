<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Field, ViewConfig, ViewType } from '$lib/types';

	export let viewType: ViewType;
	export let config: ViewConfig;
	export let fields: Field[] = [];

	const dispatch = createEventDispatcher<{
		save: { config: ViewConfig };
		close: void;
	}>();

	// Local copy of config for editing with proper defaults
	let localConfig: ViewConfig = {
		...config,
		// Ensure string fields default to empty string, not undefined
		date_field_id: config.date_field_id || '',
		title_field_id: config.title_field_id || '',
		cover_field_id: config.cover_field_id || '',
		group_by_field_id: config.group_by_field_id || ''
	};

	// Field options for different types
	$: dateFields = fields.filter(f => f.field_type === 'date');
	$: textFields = fields.filter(f => f.field_type === 'text');
	$: singleSelectFields = fields.filter(f => f.field_type === 'single_select');

	function save() {
		dispatch('save', { config: localConfig });
	}

	function close() {
		dispatch('close');
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
		if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) save();
	}
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="config-overlay" on:click|self={close}>
	<div class="config-panel">
		<div class="panel-header">
			<h3>Configure {viewType} View</h3>
			<button class="close-btn" on:click={close}>×</button>
		</div>

		<div class="panel-content">
			{#if viewType === 'calendar'}
				<div class="config-section">
					<label for="date-field">Date Field</label>
					<p class="config-hint">Select which date field to use for positioning records on the calendar</p>
					<select id="date-field" bind:value={localConfig.date_field_id}>
						<option value="">Select a date field...</option>
						{#each dateFields as field}
							<option value={field.id}>{field.name}</option>
						{/each}
					</select>
					{#if dateFields.length === 0}
						<p class="warning">No date fields found. Create a date field first.</p>
					{/if}
				</div>

				<div class="config-section">
					<label for="title-field">Title Field</label>
					<p class="config-hint">Select which field to display as the event title</p>
					<select id="title-field" bind:value={localConfig.title_field_id}>
						<option value="">Use first text field</option>
						{#each textFields as field}
							<option value={field.id}>{field.name}</option>
						{/each}
					</select>
				</div>

			{:else if viewType === 'gallery'}
				<div class="config-section">
					<label for="cover-field">Cover Image Field</label>
					<p class="config-hint">Select a text field containing image URLs to display as card covers</p>
					<select id="cover-field" bind:value={localConfig.cover_field_id}>
						<option value="">No cover image</option>
						{#each textFields as field}
							<option value={field.id}>{field.name}</option>
						{/each}
					</select>
				</div>

				<div class="config-section">
					<label for="gallery-title-field">Title Field</label>
					<p class="config-hint">Select which field to display as the card title</p>
					<select id="gallery-title-field" bind:value={localConfig.title_field_id}>
						<option value="">Use first text field</option>
						{#each textFields as field}
							<option value={field.id}>{field.name}</option>
						{/each}
					</select>
				</div>

			{:else if viewType === 'kanban'}
				<div class="config-section">
					<label for="group-field">Group By Field</label>
					<p class="config-hint">Select a single-select field to group cards into columns</p>
					<select id="group-field" bind:value={localConfig.group_by_field_id}>
						<option value="">Select a field...</option>
						{#each singleSelectFields as field}
							<option value={field.id}>{field.name}</option>
						{/each}
					</select>
					{#if singleSelectFields.length === 0}
						<p class="warning">No single-select fields found. Create one to use Kanban view.</p>
					{/if}
				</div>

			{:else if viewType === 'grid'}
				<div class="config-section">
					<p class="empty-config">Grid view has no additional configuration options.</p>
				</div>
			{/if}
		</div>

		<div class="panel-footer">
			<button class="cancel-btn" on:click={close}>Cancel</button>
			<button class="save-btn" on:click={save}>Save</button>
		</div>
	</div>
</div>

<style>
	.config-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.4);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 200;
	}

	.config-panel {
		background: white;
		border-radius: var(--radius-lg);
		width: 100%;
		max-width: 420px;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
	}

	.panel-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 16px 20px;
		border-bottom: 1px solid var(--color-border);
	}

	.panel-header h3 {
		margin: 0;
		font-size: 16px;
		font-weight: 600;
		text-transform: capitalize;
	}

	.close-btn {
		width: 28px;
		height: 28px;
		display: flex;
		align-items: center;
		justify-content: center;
		background: none;
		border: none;
		border-radius: var(--radius-md);
		cursor: pointer;
		font-size: 18px;
		color: var(--color-text-muted);
	}

	.close-btn:hover {
		background: var(--color-gray-100);
	}

	.panel-content {
		padding: 20px;
	}

	.config-section {
		margin-bottom: 20px;
	}

	.config-section:last-child {
		margin-bottom: 0;
	}

	.config-section label {
		display: block;
		font-size: 14px;
		font-weight: 500;
		margin-bottom: 4px;
	}

	.config-hint {
		font-size: 12px;
		color: var(--color-text-muted);
		margin: 0 0 8px 0;
	}

	.config-section select {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: 14px;
		background: white;
	}

	.config-section select:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: 0 0 0 3px var(--color-primary-light);
	}

	.warning {
		margin: 8px 0 0;
		padding: 8px 12px;
		background: #fef3c7;
		color: #92400e;
		border-radius: var(--radius-sm);
		font-size: 12px;
	}

	.empty-config {
		text-align: center;
		color: var(--color-text-muted);
		padding: 20px;
	}

	.panel-footer {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
		padding: 16px 20px;
		border-top: 1px solid var(--color-border);
		background: var(--color-gray-50);
		border-radius: 0 0 var(--radius-lg) var(--radius-lg);
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

	.save-btn:hover {
		background: var(--color-primary-hover);
	}
</style>
