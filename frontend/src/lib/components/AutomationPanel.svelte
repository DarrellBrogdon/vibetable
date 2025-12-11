<script lang="ts">
	import { automations } from '$lib/api/client';
	import type { Automation, Field, TriggerType, ActionType } from '$lib/types';

	export let tableId: string;
	export let fields: Field[] = [];

	let automationList: Automation[] = [];
	let loading = true;
	let error = '';
	let showCreateForm = false;

	// Form state
	let name = '';
	let triggerType: TriggerType = 'record_created';
	let actionType: ActionType = 'send_webhook';
	let webhookUrl = '';
	let creating = false;

	const triggerTypes: { value: TriggerType; label: string }[] = [
		{ value: 'record_created', label: 'When a record is created' },
		{ value: 'record_updated', label: 'When a record is updated' },
		{ value: 'record_deleted', label: 'When a record is deleted' },
		{ value: 'field_value_changed', label: 'When a field value changes' },
	];

	const actionTypes: { value: ActionType; label: string }[] = [
		{ value: 'send_webhook', label: 'Send a webhook' },
		{ value: 'update_record', label: 'Update the record' },
		{ value: 'create_record', label: 'Create a new record' },
		{ value: 'send_email', label: 'Send an email (logged)' },
	];

	async function loadAutomations() {
		try {
			loading = true;
			error = '';
			const response = await automations.list(tableId);
			automationList = response.automations;
		} catch (err: any) {
			error = err.message || 'Failed to load automations';
		} finally {
			loading = false;
		}
	}

	async function createAutomation() {
		if (!name.trim()) return;

		try {
			creating = true;
			error = '';

			const actionConfig: Record<string, any> = {};
			if (actionType === 'send_webhook' && webhookUrl) {
				actionConfig.url = webhookUrl;
				actionConfig.method = 'POST';
			}

			await automations.create(tableId, {
				name: name.trim(),
				triggerType,
				actionType,
				actionConfig,
				enabled: true,
			});

			name = '';
			webhookUrl = '';
			showCreateForm = false;
			await loadAutomations();
		} catch (err: any) {
			error = err.message || 'Failed to create automation';
		} finally {
			creating = false;
		}
	}

	async function toggleAutomation(automation: Automation) {
		try {
			await automations.toggle(automation.id, !automation.enabled);
			await loadAutomations();
		} catch (err: any) {
			error = err.message || 'Failed to toggle automation';
		}
	}

	async function deleteAutomation(automation: Automation) {
		if (!confirm(`Are you sure you want to delete "${automation.name}"?`)) return;

		try {
			await automations.delete(automation.id);
			await loadAutomations();
		} catch (err: any) {
			error = err.message || 'Failed to delete automation';
		}
	}

	function getTriggerLabel(type: TriggerType): string {
		return triggerTypes.find((t) => t.value === type)?.label || type;
	}

	function getActionLabel(type: ActionType): string {
		return actionTypes.find((a) => a.value === type)?.label || type;
	}

	$: tableId && loadAutomations();
</script>

<div class="automation-panel">
	<div class="panel-header">
		<h3>Automations</h3>
		<button class="btn-primary btn-sm" on:click={() => (showCreateForm = !showCreateForm)}>
			{showCreateForm ? 'Cancel' : '+ New Automation'}
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
					placeholder="e.g., Notify on new records"
				/>
			</div>

			<div class="form-group">
				<label for="trigger">When this happens...</label>
				<select id="trigger" bind:value={triggerType}>
					{#each triggerTypes as trigger}
						<option value={trigger.value}>{trigger.label}</option>
					{/each}
				</select>
			</div>

			<div class="form-group">
				<label for="action">Do this...</label>
				<select id="action" bind:value={actionType}>
					{#each actionTypes as action}
						<option value={action.value}>{action.label}</option>
					{/each}
				</select>
			</div>

			{#if actionType === 'send_webhook'}
				<div class="form-group">
					<label for="webhookUrl">Webhook URL</label>
					<input
						id="webhookUrl"
						type="url"
						bind:value={webhookUrl}
						placeholder="https://example.com/webhook"
					/>
				</div>
			{/if}

			<button class="btn-primary" on:click={createAutomation} disabled={creating || !name.trim()}>
				{creating ? 'Creating...' : 'Create Automation'}
			</button>
		</div>
	{/if}

	{#if loading}
		<div class="loading">Loading automations...</div>
	{:else if automationList.length === 0}
		<div class="empty-state">
			<p>No automations yet.</p>
			<p class="hint">Create an automation to trigger actions when records change.</p>
		</div>
	{:else}
		<div class="automation-list">
			{#each automationList as automation (automation.id)}
				<div class="automation-item" class:disabled={!automation.enabled}>
					<div class="automation-info">
						<div class="automation-name">{automation.name}</div>
						<div class="automation-details">
							<span class="trigger">{getTriggerLabel(automation.trigger_type)}</span>
							<span class="arrow">→</span>
							<span class="action">{getActionLabel(automation.action_type)}</span>
						</div>
						<div class="automation-stats">
							Runs: {automation.run_count}
							{#if automation.last_triggered_at}
								• Last: {new Date(automation.last_triggered_at).toLocaleDateString()}
							{/if}
						</div>
					</div>
					<div class="automation-actions">
						<label class="toggle">
							<input
								type="checkbox"
								checked={automation.enabled}
								on:change={() => toggleAutomation(automation)}
							/>
							<span class="toggle-slider"></span>
						</label>
						<button class="btn-icon" on:click={() => deleteAutomation(automation)} title="Delete">
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
	.automation-panel {
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
		color: #e74c3c;
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

	.form-group input,
	.form-group select {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid #ddd;
		border-radius: 4px;
		font-size: 14px;
	}

	.form-group input:focus,
	.form-group select:focus {
		outline: none;
		border-color: #2d7ff9;
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

	.automation-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.automation-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px;
		background: #f9f9f9;
		border-radius: 6px;
		border: 1px solid #eee;
	}

	.automation-item.disabled {
		opacity: 0.6;
	}

	.automation-info {
		flex: 1;
	}

	.automation-name {
		font-weight: 500;
		margin-bottom: 4px;
	}

	.automation-details {
		font-size: 13px;
		color: #666;
	}

	.automation-details .arrow {
		margin: 0 6px;
		color: #999;
	}

	.automation-stats {
		font-size: 12px;
		color: #999;
		margin-top: 4px;
	}

	.automation-actions {
		display: flex;
		align-items: center;
		gap: 8px;
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
</style>
