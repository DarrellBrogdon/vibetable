<script lang="ts">
	import { onMount } from 'svelte';
	import type { Activity, User, Field } from '$lib/types';
	import { activity as activityApi, type ActivityFilters } from '$lib/api/client';

	export let baseId: string = '';
	export let recordId: string = '';
	export let fields: Field[] = [];
	export let limit: number = 50;

	let activities: Activity[] = [];
	let loading = true;
	let error = '';
	let hasMore = false;
	let offset = 0;

	onMount(() => {
		loadActivities();
	});

	async function loadActivities() {
		loading = true;
		error = '';
		try {
			if (recordId) {
				const result = await activityApi.listForRecord(recordId, limit);
				activities = result.activities;
			} else if (baseId) {
				const result = await activityApi.listForBase(baseId, { limit, offset });
				activities = result.activities;
				hasMore = result.activities.length === limit;
			}
		} catch (e: any) {
			error = e.message || 'Failed to load activity';
		} finally {
			loading = false;
		}
	}

	async function loadMore() {
		if (!baseId || loading) return;

		offset += limit;
		loading = true;
		try {
			const result = await activityApi.listForBase(baseId, { limit, offset });
			activities = [...activities, ...result.activities];
			hasMore = result.activities.length === limit;
		} catch (e: any) {
			error = e.message || 'Failed to load more activity';
		} finally {
			loading = false;
		}
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const minutes = Math.floor(diff / 60000);
		const hours = Math.floor(minutes / 60);
		const days = Math.floor(hours / 24);

		if (minutes < 1) return 'just now';
		if (minutes < 60) return `${minutes}m ago`;
		if (hours < 24) return `${hours}h ago`;
		if (days < 7) return `${days}d ago`;
		return date.toLocaleDateString();
	}

	function getUserInitials(user?: User): string {
		if (!user) return '?';
		if (user.name) {
			return user.name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2);
		}
		return user.email[0].toUpperCase();
	}

	function getUserName(user?: User): string {
		if (!user) return 'Unknown';
		return user.name || user.email.split('@')[0];
	}

	function getActionVerb(action: string): string {
		switch (action) {
			case 'create': return 'created';
			case 'update': return 'updated';
			case 'delete': return 'deleted';
			default: return action;
		}
	}

	function getEntityLabel(activity: Activity): string {
		if (activity.entity_name) {
			return `"${activity.entity_name}"`;
		}
		return `a ${activity.entity_type}`;
	}

	function getFieldName(fieldId: string): string {
		const field = fields.find(f => f.id === fieldId);
		return field?.name || 'Unknown field';
	}

	function formatChangeValue(value: any): string {
		if (value === null || value === undefined) return 'empty';
		if (typeof value === 'boolean') return value ? 'checked' : 'unchecked';
		if (Array.isArray(value)) return value.length > 0 ? `${value.length} items` : 'empty';
		return String(value).slice(0, 50) + (String(value).length > 50 ? '...' : '');
	}

	function getActionIcon(action: string): string {
		switch (action) {
			case 'create': return '+';
			case 'update': return '~';
			case 'delete': return '−';
			default: return '•';
		}
	}

	function getActionColor(action: string): string {
		switch (action) {
			case 'create': return '#059669';
			case 'update': return '#2563eb';
			case 'delete': return '#dc2626';
			default: return '#6b7280';
		}
	}
</script>

<div class="activity-feed">
	{#if loading && activities.length === 0}
		<div class="loading">Loading activity...</div>
	{:else if error}
		<div class="error">{error}</div>
	{:else if activities.length === 0}
		<div class="empty">No activity yet</div>
	{:else}
		<div class="activities-list">
			{#each activities as activity (activity.id)}
				<div class="activity-item">
					<div class="activity-icon" style="background: {getActionColor(activity.action)}20; color: {getActionColor(activity.action)}">
						{getActionIcon(activity.action)}
					</div>
					<div class="activity-content">
						<div class="activity-summary">
							<span class="activity-user">{getUserName(activity.user)}</span>
							<span class="activity-action">{getActionVerb(activity.action)}</span>
							<span class="activity-entity">{getEntityLabel(activity)}</span>
						</div>

						{#if activity.changes && Array.isArray(activity.changes) && activity.changes.length > 0}
							<div class="activity-changes">
								{#each activity.changes as change}
									<div class="change-item">
										<span class="change-field">{change.field_name || getFieldName(change.field_id || '')}</span>
										{#if change.old_value !== undefined && change.new_value !== undefined}
											<span class="change-values">
												<span class="old-value">{formatChangeValue(change.old_value)}</span>
												<span class="arrow">→</span>
												<span class="new-value">{formatChangeValue(change.new_value)}</span>
											</span>
										{/if}
									</div>
								{/each}
							</div>
						{/if}

						<div class="activity-time">{formatDate(activity.created_at)}</div>
					</div>
				</div>
			{/each}
		</div>

		{#if hasMore && !loading}
			<button class="load-more" on:click={loadMore}>
				Load more
			</button>
		{/if}

		{#if loading && activities.length > 0}
			<div class="loading-more">Loading...</div>
		{/if}
	{/if}
</div>

<style>
	.activity-feed {
		display: flex;
		flex-direction: column;
	}

	.loading, .error, .empty {
		padding: 24px;
		text-align: center;
		color: var(--color-text-muted);
		font-size: 14px;
	}

	.error {
		color: #dc2626;
	}

	.activities-list {
		display: flex;
		flex-direction: column;
	}

	.activity-item {
		display: flex;
		gap: 12px;
		padding: 12px 16px;
		border-bottom: 1px solid var(--color-border);
	}

	.activity-item:last-child {
		border-bottom: none;
	}

	.activity-icon {
		width: 28px;
		height: 28px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 14px;
		font-weight: 600;
		flex-shrink: 0;
	}

	.activity-content {
		flex: 1;
		min-width: 0;
	}

	.activity-summary {
		font-size: 13px;
		line-height: 1.4;
	}

	.activity-user {
		font-weight: 500;
	}

	.activity-action {
		color: var(--color-text-muted);
	}

	.activity-entity {
		color: var(--color-text);
	}

	.activity-changes {
		margin-top: 6px;
		padding: 8px;
		background: var(--color-gray-50);
		border-radius: var(--radius-md);
		font-size: 12px;
	}

	.change-item {
		display: flex;
		align-items: baseline;
		gap: 8px;
		margin-bottom: 4px;
	}

	.change-item:last-child {
		margin-bottom: 0;
	}

	.change-field {
		font-weight: 500;
		color: var(--color-text-muted);
	}

	.change-values {
		display: flex;
		align-items: baseline;
		gap: 6px;
	}

	.old-value {
		text-decoration: line-through;
		color: var(--color-text-muted);
	}

	.arrow {
		color: var(--color-text-muted);
	}

	.new-value {
		color: var(--color-text);
	}

	.activity-time {
		font-size: 11px;
		color: var(--color-text-muted);
		margin-top: 4px;
	}

	.load-more {
		display: block;
		width: 100%;
		padding: 12px;
		background: none;
		border: none;
		color: var(--color-primary);
		font-size: 13px;
		cursor: pointer;
		text-align: center;
	}

	.load-more:hover {
		background: var(--color-gray-50);
	}

	.loading-more {
		padding: 12px;
		text-align: center;
		color: var(--color-text-muted);
		font-size: 13px;
	}
</style>
