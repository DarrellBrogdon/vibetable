<script lang="ts">
	import { apiKeys } from '$lib/api/client';
	import type { APIKey, APIKeyWithToken } from '$lib/types';

	let keyList: APIKey[] = [];
	let loading = true;
	let error = '';
	let showCreateForm = false;
	let newKeyToken: string | null = null;

	// Form state
	let name = '';
	let creating = false;

	async function loadAPIKeys() {
		try {
			loading = true;
			error = '';
			const response = await apiKeys.list();
			keyList = response.api_keys || [];
		} catch (err: any) {
			error = err.message || 'Failed to load API keys';
		} finally {
			loading = false;
		}
	}

	async function createAPIKey() {
		if (!name.trim()) return;

		try {
			creating = true;
			error = '';

			const result = await apiKeys.create(name.trim());
			newKeyToken = result.token;
			name = '';
			await loadAPIKeys();
		} catch (err: any) {
			error = err.message || 'Failed to create API key';
		} finally {
			creating = false;
		}
	}

	async function deleteAPIKey(key: APIKey) {
		if (!confirm(`Are you sure you want to delete "${key.name}"? This action cannot be undone.`)) return;

		try {
			await apiKeys.delete(key.id);
			await loadAPIKeys();
		} catch (err: any) {
			error = err.message || 'Failed to delete API key';
		}
	}

	function copyToClipboard(text: string) {
		navigator.clipboard.writeText(text);
	}

	function closeTokenModal() {
		newKeyToken = null;
		showCreateForm = false;
	}

	// Load on mount
	loadAPIKeys();
</script>

<div class="api-keys-panel">
	<div class="panel-header">
		<h3>API Keys</h3>
		<button class="btn-primary btn-sm" on:click={() => (showCreateForm = !showCreateForm)}>
			{showCreateForm ? 'Cancel' : '+ New API Key'}
		</button>
	</div>

	{#if error}
		<div class="error-message">{error}</div>
	{/if}

	{#if showCreateForm && !newKeyToken}
		<div class="create-form">
			<div class="form-group">
				<label for="name">Key Name</label>
				<input
					id="name"
					type="text"
					bind:value={name}
					placeholder="e.g., Integration Key"
				/>
			</div>

			<button class="btn-primary" on:click={createAPIKey} disabled={creating || !name.trim()}>
				{creating ? 'Creating...' : 'Create API Key'}
			</button>
		</div>
	{/if}

	{#if newKeyToken}
		<div class="token-modal">
			<div class="token-content">
				<h4>API Key Created</h4>
				<p class="warning">Copy this key now. You won't be able to see it again!</p>
				<div class="token-display">
					<code>{newKeyToken}</code>
					<button class="btn-copy" on:click={() => copyToClipboard(newKeyToken || '')}>
						Copy
					</button>
				</div>
				<button class="btn-primary" on:click={closeTokenModal}>Done</button>
			</div>
		</div>
	{/if}

	{#if loading}
		<div class="loading">Loading API keys...</div>
	{:else if keyList.length === 0}
		<div class="empty-state">
			<p>No API keys yet.</p>
			<p class="hint">Create an API key to access the VibeTable API programmatically.</p>
		</div>
	{:else}
		<div class="key-list">
			{#each keyList as key (key.id)}
				<div class="key-item">
					<div class="key-info">
						<div class="key-name">{key.name}</div>
						<div class="key-details">
							<span class="key-prefix">{key.key_prefix}...</span>
							<span class="key-date">Created {new Date(key.created_at).toLocaleDateString()}</span>
							{#if key.last_used_at}
								<span class="key-used">Last used {new Date(key.last_used_at).toLocaleDateString()}</span>
							{/if}
						</div>
					</div>
					<div class="key-actions">
						<button class="btn-icon btn-danger" on:click={() => deleteAPIKey(key)} title="Delete">
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2" />
							</svg>
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.api-keys-panel {
		padding: 16px;
		background: white;
		border-radius: 8px;
		border: 1px solid #e0e0e0;
	}

	.panel-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16px;
	}

	.panel-header h3 {
		margin: 0;
		font-size: 16px;
		font-weight: 600;
	}

	.btn-primary {
		background: #2d7ff9;
		color: white;
		border: none;
		padding: 8px 16px;
		border-radius: 4px;
		cursor: pointer;
		font-size: 14px;
	}

	.btn-primary:hover {
		background: #1a6fe8;
	}

	.btn-primary:disabled {
		background: #ccc;
		cursor: not-allowed;
	}

	.btn-sm {
		padding: 6px 12px;
		font-size: 13px;
	}

	.btn-icon {
		background: transparent;
		border: none;
		padding: 6px;
		cursor: pointer;
		color: #666;
		border-radius: 4px;
	}

	.btn-icon:hover {
		background: #f0f0f0;
	}

	.btn-icon.btn-danger:hover {
		color: #e74c3c;
	}

	.btn-copy {
		background: #f0f0f0;
		border: none;
		padding: 4px 8px;
		border-radius: 4px;
		cursor: pointer;
		font-size: 12px;
	}

	.btn-copy:hover {
		background: #e0e0e0;
	}

	.error-message {
		background: #ffebee;
		color: #c62828;
		padding: 12px;
		border-radius: 4px;
		margin-bottom: 16px;
	}

	.create-form {
		background: #f9f9f9;
		padding: 16px;
		border-radius: 8px;
		margin-bottom: 16px;
	}

	.form-group {
		margin-bottom: 12px;
	}

	.form-group label {
		display: block;
		font-size: 13px;
		font-weight: 500;
		margin-bottom: 4px;
		color: #333;
	}

	.form-group input {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid #ddd;
		border-radius: 4px;
		font-size: 14px;
	}

	.form-group input:focus {
		outline: none;
		border-color: #2d7ff9;
	}

	.token-modal {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.token-content {
		background: white;
		padding: 24px;
		border-radius: 8px;
		max-width: 500px;
		width: 90%;
	}

	.token-content h4 {
		margin: 0 0 12px 0;
	}

	.warning {
		color: #e67e22;
		font-size: 14px;
		margin-bottom: 16px;
	}

	.token-display {
		display: flex;
		gap: 8px;
		align-items: center;
		background: #f5f5f5;
		padding: 12px;
		border-radius: 4px;
		margin-bottom: 16px;
	}

	.token-display code {
		flex: 1;
		font-family: monospace;
		font-size: 12px;
		word-break: break-all;
	}

	.loading,
	.empty-state {
		text-align: center;
		padding: 24px;
		color: #666;
	}

	.empty-state .hint {
		font-size: 13px;
		color: #999;
	}

	.key-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.key-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px;
		background: #f9f9f9;
		border-radius: 6px;
		border: 1px solid #eee;
	}

	.key-info {
		flex: 1;
	}

	.key-name {
		font-weight: 500;
		margin-bottom: 4px;
	}

	.key-details {
		font-size: 12px;
		color: #666;
		display: flex;
		gap: 12px;
	}

	.key-prefix {
		font-family: monospace;
		background: #e0e0e0;
		padding: 2px 6px;
		border-radius: 3px;
	}

	.key-actions {
		display: flex;
		align-items: center;
		gap: 8px;
	}
</style>
