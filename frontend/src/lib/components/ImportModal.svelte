<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { csv, type CSVPreviewResponse } from '$lib/api/client';
	import type { Field } from '$lib/types';

	export let tableId: string;
	export let fields: Field[] = [];

	const dispatch = createEventDispatcher();

	let step: 'upload' | 'mapping' | 'importing' | 'complete' = 'upload';
	let csvData = '';
	let preview: CSVPreviewResponse | null = null;
	let mappings: { [column: string]: string } = {};
	let importing = false;
	let result: { imported: number; skipped: number; errors: number } | null = null;
	let errorMessage = '';

	let dragActive = false;

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		dragActive = true;
	}

	function handleDragLeave() {
		dragActive = false;
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		dragActive = false;
		const file = e.dataTransfer?.files[0];
		if (file) {
			readFile(file);
		}
	}

	function handleFileSelect(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (file) {
			readFile(file);
		}
	}

	function readFile(file: File) {
		const reader = new FileReader();
		reader.onload = async (e) => {
			csvData = e.target?.result as string;
			await loadPreview();
		};
		reader.readAsText(file);
	}

	async function loadPreview() {
		errorMessage = '';
		try {
			preview = await csv.preview(tableId, csvData);
			// Initialize mappings with best guesses
			for (const col of preview.columns) {
				const matchingField = fields.find(
					f => f.name.toLowerCase() === col.toLowerCase()
				);
				mappings[col] = matchingField?.id || '';
			}
			step = 'mapping';
		} catch (e: any) {
			errorMessage = e.message || 'Failed to parse CSV';
		}
	}

	async function doImport() {
		importing = true;
		errorMessage = '';
		step = 'importing';

		try {
			result = await csv.import(tableId, csvData, mappings);
			step = 'complete';
		} catch (e: any) {
			errorMessage = e.message || 'Failed to import records';
			step = 'mapping';
		} finally {
			importing = false;
		}
	}

	function close() {
		dispatch('close', { imported: result?.imported || 0 });
	}

	function reset() {
		step = 'upload';
		csvData = '';
		preview = null;
		mappings = {};
		result = null;
		errorMessage = '';
	}
</script>

<div class="modal-overlay" on:click|self={close}>
	<div class="modal" on:click|stopPropagation>
		<div class="modal-header">
			<h2>Import CSV</h2>
			<button class="close-btn" on:click={close}>×</button>
		</div>

		{#if step === 'upload'}
			<div class="modal-body">
				<div
					class="drop-zone"
					class:active={dragActive}
					on:dragover={handleDragOver}
					on:dragleave={handleDragLeave}
					on:drop={handleDrop}
				>
					<div class="drop-content">
						<span class="drop-icon">📄</span>
						<p>Drag and drop a CSV file here</p>
						<p class="or-text">or</p>
						<label class="file-input-label">
							<input
								type="file"
								accept=".csv,text/csv"
								on:change={handleFileSelect}
								hidden
							/>
							<span class="browse-btn">Browse files</span>
						</label>
					</div>
				</div>
				{#if errorMessage}
					<div class="error-message">{errorMessage}</div>
				{/if}
			</div>

		{:else if step === 'mapping'}
			<div class="modal-body">
				<p class="info-text">
					Map CSV columns to table fields. Found {preview?.total || 0} rows.
				</p>

				<div class="mapping-list">
					{#each preview?.columns || [] as column}
						<div class="mapping-row">
							<div class="column-name">
								<span class="csv-label">CSV:</span>
								{column}
							</div>
							<span class="arrow">→</span>
							<select bind:value={mappings[column]}>
								<option value="">Skip this column</option>
								{#each fields as field}
									<option value={field.id}>{field.name}</option>
								{/each}
							</select>
						</div>
					{/each}
				</div>

				{#if preview && preview.rows.length > 0}
					<div class="preview-section">
						<h4>Preview (first {preview.rows.length} rows)</h4>
						<div class="preview-table-wrapper">
							<table class="preview-table">
								<thead>
									<tr>
										{#each preview.columns as col}
											<th>{col}</th>
										{/each}
									</tr>
								</thead>
								<tbody>
									{#each preview.rows as row}
										<tr>
											{#each preview.columns as col}
												<td>{row[col] || ''}</td>
											{/each}
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					</div>
				{/if}

				{#if errorMessage}
					<div class="error-message">{errorMessage}</div>
				{/if}
			</div>

			<div class="modal-footer">
				<button class="secondary-btn" on:click={reset}>Back</button>
				<button
					class="primary-btn"
					on:click={doImport}
					disabled={importing || Object.values(mappings).filter(v => v).length === 0}
				>
					Import {preview?.total || 0} rows
				</button>
			</div>

		{:else if step === 'importing'}
			<div class="modal-body">
				<div class="importing-state">
					<div class="spinner"></div>
					<p>Importing records...</p>
				</div>
			</div>

		{:else if step === 'complete'}
			<div class="modal-body">
				<div class="success-state">
					<span class="success-icon">✓</span>
					<h3>Import Complete</h3>
					<div class="result-stats">
						<div class="stat">
							<span class="stat-value">{result?.imported || 0}</span>
							<span class="stat-label">Imported</span>
						</div>
						{#if (result?.skipped || 0) > 0}
							<div class="stat">
								<span class="stat-value">{result?.skipped}</span>
								<span class="stat-label">Skipped</span>
							</div>
						{/if}
						{#if (result?.errors || 0) > 0}
							<div class="stat error">
								<span class="stat-value">{result?.errors}</span>
								<span class="stat-label">Errors</span>
							</div>
						{/if}
					</div>
				</div>
			</div>

			<div class="modal-footer">
				<button class="primary-btn" on:click={close}>Done</button>
			</div>
		{/if}
	</div>
</div>

<style>
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
		width: 100%;
		max-width: 600px;
		max-height: 80vh;
		display: flex;
		flex-direction: column;
		box-shadow: var(--shadow-lg);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-md) var(--spacing-lg);
		border-bottom: 1px solid var(--color-border);
	}

	.modal-header h2 {
		margin: 0;
		font-size: var(--font-size-lg);
	}

	.close-btn {
		width: 32px;
		height: 32px;
		border: none;
		background: var(--color-gray-100);
		border-radius: 50%;
		cursor: pointer;
		font-size: 1.2rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.close-btn:hover {
		background: var(--color-gray-200);
	}

	.modal-body {
		padding: var(--spacing-lg);
		overflow-y: auto;
		flex: 1;
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--spacing-sm);
		padding: var(--spacing-md) var(--spacing-lg);
		border-top: 1px solid var(--color-border);
	}

	.drop-zone {
		border: 2px dashed var(--color-border);
		border-radius: var(--radius-lg);
		padding: var(--spacing-xl);
		text-align: center;
		transition: all 0.2s;
	}

	.drop-zone.active {
		border-color: var(--color-primary);
		background: var(--color-primary-light);
	}

	.drop-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--spacing-sm);
	}

	.drop-icon {
		font-size: 3rem;
	}

	.or-text {
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
	}

	.browse-btn {
		display: inline-block;
		padding: var(--spacing-sm) var(--spacing-md);
		background: var(--color-primary);
		color: white;
		border-radius: var(--radius-md);
		cursor: pointer;
		font-weight: 500;
	}

	.browse-btn:hover {
		background: var(--color-primary-hover);
	}

	.file-input-label {
		cursor: pointer;
	}

	.info-text {
		margin: 0 0 var(--spacing-md);
		color: var(--color-text-muted);
	}

	.mapping-list {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-sm);
		margin-bottom: var(--spacing-md);
	}

	.mapping-row {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
	}

	.column-name {
		flex: 1;
		padding: var(--spacing-xs) var(--spacing-sm);
		background: var(--color-gray-100);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
	}

	.csv-label {
		color: var(--color-text-muted);
		font-size: var(--font-size-xs);
		margin-right: var(--spacing-xs);
	}

	.arrow {
		color: var(--color-text-muted);
	}

	.mapping-row select {
		flex: 1;
		padding: var(--spacing-xs) var(--spacing-sm);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
	}

	.preview-section {
		margin-top: var(--spacing-md);
	}

	.preview-section h4 {
		margin: 0 0 var(--spacing-sm);
		font-size: var(--font-size-sm);
		color: var(--color-text-muted);
	}

	.preview-table-wrapper {
		overflow-x: auto;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
	}

	.preview-table {
		width: 100%;
		border-collapse: collapse;
		font-size: var(--font-size-sm);
	}

	.preview-table th,
	.preview-table td {
		padding: var(--spacing-xs) var(--spacing-sm);
		text-align: left;
		border-bottom: 1px solid var(--color-border);
		white-space: nowrap;
		max-width: 200px;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.preview-table th {
		background: var(--color-gray-50);
		font-weight: 500;
	}

	.preview-table tr:last-child td {
		border-bottom: none;
	}

	.error-message {
		margin-top: var(--spacing-md);
		padding: var(--spacing-sm);
		background: #fee2e2;
		color: var(--color-error);
		border-radius: var(--radius-sm);
		font-size: var(--font-size-sm);
	}

	.importing-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--spacing-md);
		padding: var(--spacing-xl);
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 3px solid var(--color-gray-200);
		border-top-color: var(--color-primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.success-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--spacing-md);
		padding: var(--spacing-xl);
	}

	.success-icon {
		width: 60px;
		height: 60px;
		background: #dcfce7;
		color: #16a34a;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 2rem;
	}

	.success-state h3 {
		margin: 0;
	}

	.result-stats {
		display: flex;
		gap: var(--spacing-lg);
	}

	.stat {
		display: flex;
		flex-direction: column;
		align-items: center;
	}

	.stat-value {
		font-size: var(--font-size-xl);
		font-weight: 600;
	}

	.stat-label {
		font-size: var(--font-size-sm);
		color: var(--color-text-muted);
	}

	.stat.error .stat-value {
		color: var(--color-error);
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
		opacity: 0.5;
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
</style>
