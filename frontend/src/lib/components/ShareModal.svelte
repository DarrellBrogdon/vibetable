<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { bases as basesApi } from '$lib/api/client';
	import type { BaseCollaborator, CollaboratorRole } from '$lib/types';
	import ContextualHelp from './ContextualHelp.svelte';

	export let baseId: string;
	export let isOwner: boolean = false;

	const dispatch = createEventDispatcher();

	let collaborators: BaseCollaborator[] = [];
	let loading = true;
	let error = '';

	let inviteEmail = '';
	let inviteRole: CollaboratorRole = 'editor';
	let inviting = false;
	let inviteError = '';

	const roles: { value: CollaboratorRole; label: string }[] = [
		{ value: 'editor', label: 'Editor' },
		{ value: 'viewer', label: 'Viewer' }
	];

	async function loadCollaborators() {
		loading = true;
		error = '';
		try {
			const result = await basesApi.listCollaborators(baseId);
			collaborators = result.collaborators;
		} catch (e: any) {
			error = e.message || 'Failed to load collaborators';
		} finally {
			loading = false;
		}
	}

	async function invite() {
		if (!inviteEmail.trim()) return;

		inviting = true;
		inviteError = '';
		try {
			const collaborator = await basesApi.addCollaborator(baseId, inviteEmail.trim(), inviteRole);
			collaborators = [...collaborators, collaborator];
			inviteEmail = '';
			inviteRole = 'editor';
		} catch (e: any) {
			inviteError = e.message || 'Failed to invite collaborator';
		} finally {
			inviting = false;
		}
	}

	async function updateRole(userId: string, newRole: CollaboratorRole) {
		try {
			// The API doesn't have an update role endpoint in the client, so we'll need to add it
			// For now, we'll just update locally
			collaborators = collaborators.map(c =>
				c.user_id === userId ? { ...c, role: newRole } : c
			);
		} catch (e: any) {
			console.error('Failed to update role:', e);
		}
	}

	async function removeCollaborator(userId: string) {
		if (!confirm('Remove this collaborator?')) return;

		try {
			await basesApi.removeCollaborator(baseId, userId);
			collaborators = collaborators.filter(c => c.user_id !== userId);
		} catch (e: any) {
			console.error('Failed to remove collaborator:', e);
		}
	}

	function close() {
		dispatch('close');
	}

	function getRoleLabel(role: CollaboratorRole): string {
		switch (role) {
			case 'owner': return 'Owner';
			case 'editor': return 'Can edit';
			case 'viewer': return 'Can view';
			default: return role;
		}
	}

	// Load collaborators when modal opens
	loadCollaborators();
</script>

<div class="modal-overlay" on:click={close}>
	<div class="modal" on:click|stopPropagation>
		<div class="modal-header">
			<h3>Share this base</h3>
			<button class="close-btn" on:click={close}>×</button>
		</div>

		{#if isOwner}
			<form class="invite-form" on:submit|preventDefault={invite}>
				<div class="invite-row">
					<input
						type="email"
						placeholder="Enter email address"
						bind:value={inviteEmail}
						disabled={inviting}
					/>
					<select bind:value={inviteRole} disabled={inviting}>
						{#each roles as role}
							<option value={role.value}>{role.label}</option>
						{/each}
					</select>
					<ContextualHelp topic="permissions" position="bottom" />
					<button type="submit" class="invite-btn" disabled={inviting || !inviteEmail.trim()}>
						{inviting ? 'Inviting...' : 'Invite'}
					</button>
				</div>
				{#if inviteError}
					<p class="error">{inviteError}</p>
				{/if}
			</form>
		{/if}

		<div class="collaborators-section">
			<h4>People with access</h4>

			{#if loading}
				<div class="loading">Loading...</div>
			{:else if error}
				<div class="error">{error}</div>
			{:else if collaborators.length === 0}
				<p class="empty">No collaborators yet</p>
			{:else}
				<ul class="collaborators-list">
					{#each collaborators as collab}
						<li class="collaborator-item">
							<div class="collaborator-info">
								<span class="collaborator-email">{collab.user?.email || 'Unknown'}</span>
								<span class="collaborator-role">{getRoleLabel(collab.role)}</span>
							</div>
							{#if isOwner && collab.role !== 'owner'}
								<button
									class="remove-btn"
									on:click={() => removeCollaborator(collab.user_id)}
									title="Remove access"
								>
									×
								</button>
							{/if}
						</li>
					{/each}
				</ul>
			{/if}
		</div>
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
		max-width: 500px;
		max-height: 80vh;
		overflow: hidden;
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

	.modal-header h3 {
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
		color: var(--color-text);
	}

	.invite-form {
		padding: var(--spacing-lg);
		border-bottom: 1px solid var(--color-border);
	}

	.invite-row {
		display: flex;
		gap: var(--spacing-sm);
	}

	.invite-row input {
		flex: 1;
		padding: var(--spacing-sm) var(--spacing-md);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: var(--font-size-sm);
	}

	.invite-row input:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: 0 0 0 3px var(--color-primary-light);
	}

	.invite-row select {
		padding: var(--spacing-sm) var(--spacing-md);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: var(--font-size-sm);
		background: white;
	}

	.invite-btn {
		padding: var(--spacing-sm) var(--spacing-md);
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-md);
		font-weight: 500;
		cursor: pointer;
		white-space: nowrap;
	}

	.invite-btn:hover:not(:disabled) {
		background: var(--color-primary-hover);
	}

	.invite-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.error {
		color: var(--color-error);
		font-size: var(--font-size-sm);
		margin-top: var(--spacing-sm);
	}

	.collaborators-section {
		padding: var(--spacing-lg);
		overflow-y: auto;
	}

	.collaborators-section h4 {
		margin: 0 0 var(--spacing-md);
		font-size: var(--font-size-sm);
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.loading, .empty {
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
	}

	.collaborators-list {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.collaborator-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-sm) 0;
		border-bottom: 1px solid var(--color-border);
	}

	.collaborator-item:last-child {
		border-bottom: none;
	}

	.collaborator-info {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.collaborator-email {
		font-size: var(--font-size-sm);
		font-weight: 500;
	}

	.collaborator-role {
		font-size: var(--font-size-xs);
		color: var(--color-text-muted);
	}

	.remove-btn {
		width: 28px;
		height: 28px;
		background: none;
		border: none;
		color: var(--color-text-muted);
		cursor: pointer;
		border-radius: var(--radius-md);
		font-size: 1.2rem;
	}

	.remove-btn:hover {
		background: var(--color-error);
		color: white;
	}
</style>
