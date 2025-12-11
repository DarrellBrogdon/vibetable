<script lang="ts">
	import { webhooks } from '$lib/api/client';
	import type { Webhook, WebhookEvent, WebhookDelivery } from '$lib/types';

	export let baseId: string;

	let webhookList: Webhook[] = [];
	let loading = true;
	let error = '';
	let showCreateForm = false;
	let selectedWebhook: Webhook | null = null;
	let deliveries: WebhookDelivery[] = [];

	// Form state
	let name = '';
	let url = '';
	let secret = '';
	let events: WebhookEvent[] = ['record.created', 'record.updated', 'record.deleted'];
	let creating = false;

	const eventOptions: { value: WebhookEvent; label: string }[] = [
		{ value: 'record.created', label: 'Record Created' },
		{ value: 'record.updated', label: 'Record Updated' },
		{ value: 'record.deleted', label: 'Record Deleted' },
	];

	async function loadWebhooks() {
		try {
			loading = true;
			error = '';
			const response = await webhooks.list(baseId);
			webhookList = response.webhooks || [];
		} catch (err: any) {
			error = err.message || 'Failed to load webhooks';
		} finally {
			loading = false;
		}
	}

	async function createWebhook() {
		if (!name.trim() || !url.trim()) return;

		try {
			creating = true;
			error = '';

			await webhooks.create(baseId, name.trim(), url.trim(), events, secret || undefined);
			name = '';
			url = '';
			secret = '';
			events = ['record.created', 'record.updated', 'record.deleted'];
			showCreateForm = false;
			await loadWebhooks();
		} catch (err: any) {
			error = err.message || 'Failed to create webhook';
		} finally {
			creating = false;
		}
	}

	async function toggleWebhook(webhook: Webhook) {
		try {
			await webhooks.update(webhook.id, { is_active: !webhook.is_active });
			await loadWebhooks();
		} catch (err: any) {
			error = err.message || 'Failed to update webhook';
		}
	}

	async function deleteWebhook(webhook: Webhook) {
		if (!confirm(`Are you sure you want to delete "${webhook.name}"?`)) return;

		try {
			await webhooks.delete(webhook.id);
			await loadWebhooks();
		} catch (err: any) {
			error = err.message || 'Failed to delete webhook';
		}
	}

	async function viewDeliveries(webhook: Webhook) {
		try {
			selectedWebhook = webhook;
			const response = await webhooks.listDeliveries(webhook.id);
			deliveries = response.deliveries || [];
		} catch (err: any) {
			error = err.message || 'Failed to load deliveries';
		}
	}

	function closeDeliveries() {
		selectedWebhook = null;
		deliveries = [];
	}

	function toggleEvent(event: WebhookEvent) {
		if (events.includes(event)) {
			events = events.filter((e) => e !== event);
		} else {
			events = [...events, event];
		}
	}

	$: baseId && loadWebhooks();
</script>

<div class="webhooks-panel">
	<div class="panel-header">
		<h3>Webhooks</h3>
		<button class="btn-primary btn-sm" on:click={() => (showCreateForm = !showCreateForm)}>
			{showCreateForm ? 'Cancel' : '+ New Webhook'}
		</button>
	</div>

	{#if error}
		<div class="error-message">{error}</div>
	{/if}

	{#if showCreateForm}
		<div class="create-form">
			<div class="form-group">
				<label for="name">Name</label>
				<input
					id="name"
					type="text"
					bind:value={name}
					placeholder="e.g., Slack Notification"
				/>
			</div>

			<div class="form-group">
				<label for="url">Endpoint URL</label>
				<input
					id="url"
					type="url"
					bind:value={url}
					placeholder="https://example.com/webhook"
				/>
			</div>

			<div class="form-group">
				<label>Events</label>
				<div class="event-checkboxes">
					{#each eventOptions as option}
						<label class="checkbox-label">
							<input
								type="checkbox"
								checked={events.includes(option.value)}
								on:change={() => toggleEvent(option.value)}
							/>
							{option.label}
						</label>
					{/each}
				</div>
			</div>

			<div class="form-group">
				<label for="secret">Secret (optional)</label>
				<input
					id="secret"
					type="text"
					bind:value={secret}
					placeholder="For HMAC signature verification"
				/>
				<p class="hint">If set, requests will include an X-Webhook-Signature header.</p>
			</div>

			<button class="btn-primary" on:click={createWebhook} disabled={creating || !name.trim() || !url.trim()}>
				{creating ? 'Creating...' : 'Create Webhook'}
			</button>
		</div>
	{/if}

	{#if loading}
		<div class="loading">Loading webhooks...</div>
	{:else if webhookList.length === 0}
		<div class="empty-state">
			<p>No webhooks yet.</p>
			<p class="hint">Create a webhook to receive notifications when records change.</p>
		</div>
	{:else}
		<div class="webhook-list">
			{#each webhookList as webhook (webhook.id)}
				<div class="webhook-item" class:disabled={!webhook.is_active}>
					<div class="webhook-info">
						<div class="webhook-name">{webhook.name}</div>
						<div class="webhook-url">{webhook.url}</div>
						<div class="webhook-events">
							{#each webhook.events as event}
								<span class="event-badge">{event}</span>
							{/each}
						</div>
					</div>
					<div class="webhook-actions">
						<button class="btn-link" on:click={() => viewDeliveries(webhook)}>
							View Logs
						</button>
						<label class="toggle">
							<input
								type="checkbox"
								checked={webhook.is_active}
								on:change={() => toggleWebhook(webhook)}
							/>
							<span class="toggle-slider"></span>
						</label>
						<button class="btn-icon" on:click={() => deleteWebhook(webhook)} title="Delete">
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2" />
							</svg>
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	{#if selectedWebhook}
		<div class="deliveries-modal">
			<div class="deliveries-content">
				<div class="deliveries-header">
					<h4>Delivery Log: {selectedWebhook.name}</h4>
					<button class="btn-close" on:click={closeDeliveries}>&times;</button>
				</div>
				{#if deliveries.length === 0}
					<div class="empty-state">No deliveries yet.</div>
				{:else}
					<div class="deliveries-list">
						{#each deliveries as delivery (delivery.id)}
							<div class="delivery-item" class:success={delivery.response_status && delivery.response_status >= 200 && delivery.response_status < 300} class:error={delivery.error || (delivery.response_status && delivery.response_status >= 400)}>
								<div class="delivery-info">
									<span class="delivery-event">{delivery.event_type}</span>
									<span class="delivery-time">{new Date(delivery.delivered_at).toLocaleString()}</span>
								</div>
								<div class="delivery-status">
									{#if delivery.response_status}
										<span class="status-code">{delivery.response_status}</span>
									{/if}
									{#if delivery.duration_ms}
										<span class="duration">{delivery.duration_ms}ms</span>
									{/if}
									{#if delivery.error}
										<span class="error-text">{delivery.error}</span>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	.webhooks-panel {
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

	.btn-link {
		background: none;
		border: none;
		color: #2d7ff9;
		cursor: pointer;
		font-size: 13px;
	}

	.btn-link:hover {
		text-decoration: underline;
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
		color: #e74c3c;
	}

	.btn-close {
		background: none;
		border: none;
		font-size: 24px;
		cursor: pointer;
		color: #666;
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

	.form-group input[type="text"],
	.form-group input[type="url"] {
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

	.event-checkboxes {
		display: flex;
		flex-wrap: wrap;
		gap: 12px;
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 14px;
		cursor: pointer;
	}

	.hint {
		font-size: 12px;
		color: #999;
		margin-top: 4px;
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

	.webhook-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.webhook-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px;
		background: #f9f9f9;
		border-radius: 6px;
		border: 1px solid #eee;
	}

	.webhook-item.disabled {
		opacity: 0.6;
	}

	.webhook-info {
		flex: 1;
	}

	.webhook-name {
		font-weight: 500;
		margin-bottom: 4px;
	}

	.webhook-url {
		font-size: 12px;
		color: #666;
		font-family: monospace;
		margin-bottom: 4px;
	}

	.webhook-events {
		display: flex;
		gap: 6px;
	}

	.event-badge {
		background: #e3f2fd;
		color: #1976d2;
		padding: 2px 8px;
		border-radius: 12px;
		font-size: 11px;
	}

	.webhook-actions {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	/* Toggle switch */
	.toggle {
		position: relative;
		display: inline-block;
		width: 40px;
		height: 22px;
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
		background-color: #ccc;
		transition: 0.3s;
		border-radius: 22px;
	}

	.toggle-slider:before {
		position: absolute;
		content: '';
		height: 16px;
		width: 16px;
		left: 3px;
		bottom: 3px;
		background-color: white;
		transition: 0.3s;
		border-radius: 50%;
	}

	.toggle input:checked + .toggle-slider {
		background-color: #4caf50;
	}

	.toggle input:checked + .toggle-slider:before {
		transform: translateX(18px);
	}

	/* Deliveries modal */
	.deliveries-modal {
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

	.deliveries-content {
		background: white;
		border-radius: 8px;
		max-width: 600px;
		width: 90%;
		max-height: 80vh;
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.deliveries-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 16px;
		border-bottom: 1px solid #e0e0e0;
	}

	.deliveries-header h4 {
		margin: 0;
	}

	.deliveries-list {
		overflow-y: auto;
		padding: 16px;
	}

	.delivery-item {
		padding: 12px;
		border: 1px solid #e0e0e0;
		border-radius: 6px;
		margin-bottom: 8px;
	}

	.delivery-item.success {
		border-left: 3px solid #4caf50;
	}

	.delivery-item.error {
		border-left: 3px solid #e74c3c;
	}

	.delivery-info {
		display: flex;
		justify-content: space-between;
		margin-bottom: 8px;
	}

	.delivery-event {
		font-weight: 500;
	}

	.delivery-time {
		font-size: 12px;
		color: #666;
	}

	.delivery-status {
		display: flex;
		gap: 12px;
		font-size: 13px;
	}

	.status-code {
		background: #f0f0f0;
		padding: 2px 8px;
		border-radius: 4px;
	}

	.duration {
		color: #666;
	}

	.error-text {
		color: #e74c3c;
	}
</style>
