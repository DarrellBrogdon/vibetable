<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { bases as basesApi } from '$lib/api/client';
	import { authStore } from '$lib/stores/auth';
	import type { Base } from '$lib/types';
	import HelpButton from '$lib/components/HelpButton.svelte';

	let bases: Base[] = [];
	let loading = true;
	let showNewBaseModal = false;
	let newBaseName = '';
	let creating = false;

	onMount(async () => {
		await loadBases();
	});

	async function loadBases() {
		try {
			const result = await basesApi.list();
			bases = result.bases;
		} catch (e) {
			console.error('Failed to load bases:', e);
		} finally {
			loading = false;
		}
	}

	async function createBase() {
		if (!newBaseName.trim()) return;

		creating = true;
		try {
			const base = await basesApi.create(newBaseName.trim());
			bases = [base, ...bases];
			showNewBaseModal = false;
			newBaseName = '';
			goto(`/bases/${base.id}`);
		} catch (e) {
			console.error('Failed to create base:', e);
		} finally {
			creating = false;
		}
	}

	function formatDate(dateStr: string) {
		return new Date(dateStr).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}
</script>

<div class="page">
	<header class="header">
		<div class="header-left">
			<span class="logo-icon">📊</span>
			<h1>VibeTable</h1>
		</div>
		<div class="header-right">
			<HelpButton />
			<span class="user-email">{$authStore.user?.email}</span>
			<button class="logout-btn" on:click={() => authStore.logout()}>
				Logout
			</button>
		</div>
	</header>

	<main class="main">
		<div class="page-header">
			<h2>Your Bases</h2>
			<button class="primary-btn" on:click={() => showNewBaseModal = true}>
				+ New Base
			</button>
		</div>

		{#if loading}
			<div class="loading">Loading...</div>
		{:else if bases.length === 0}
			<div class="empty-state">
				<div class="empty-icon">📁</div>
				<h3>No bases yet</h3>
				<p>Create your first base to get started</p>
				<button class="primary-btn" on:click={() => showNewBaseModal = true}>
					Create a base
				</button>
			</div>
		{:else}
			<div class="bases-grid">
				{#each bases as base}
					<a href="/bases/{base.id}" class="base-card">
						<div class="base-icon">📊</div>
						<div class="base-info">
							<h3>{base.name}</h3>
							<p class="base-meta">
								{base.role} · Updated {formatDate(base.updated_at)}
							</p>
						</div>
					</a>
				{/each}
			</div>
		{/if}
	</main>
</div>

{#if showNewBaseModal}
	<div class="modal-overlay" on:click={() => showNewBaseModal = false}>
		<div class="modal" on:click|stopPropagation>
			<h3>Create new base</h3>
			<form on:submit|preventDefault={createBase}>
				<input
					type="text"
					placeholder="Base name"
					bind:value={newBaseName}
					disabled={creating}
					autofocus
				/>
				<div class="modal-actions">
					<button type="button" class="secondary-btn" on:click={() => showNewBaseModal = false}>
						Cancel
					</button>
					<button type="submit" class="primary-btn" disabled={creating || !newBaseName.trim()}>
						{creating ? 'Creating...' : 'Create'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<style>
	.page {
		min-height: 100vh;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-md) var(--spacing-lg);
		background: white;
		border-bottom: 1px solid var(--color-border);
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
	}

	.logo-icon {
		font-size: 1.5rem;
	}

	.header h1 {
		font-size: var(--font-size-lg);
		color: var(--color-primary);
		margin: 0;
	}

	.header-right {
		display: flex;
		align-items: center;
		gap: var(--spacing-md);
	}

	.user-email {
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
	}

	.logout-btn {
		padding: var(--spacing-xs) var(--spacing-sm);
		background: none;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		color: var(--color-text-muted);
		cursor: pointer;
		font-size: var(--font-size-sm);
	}

	.logout-btn:hover {
		background: var(--color-gray-50);
	}

	.main {
		max-width: 1200px;
		margin: 0 auto;
		padding: var(--spacing-xl);
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--spacing-lg);
	}

	.page-header h2 {
		margin: 0;
		font-size: var(--font-size-xl);
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
		opacity: 0.6;
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

	.loading {
		text-align: center;
		padding: var(--spacing-xl);
		color: var(--color-text-muted);
	}

	.empty-state {
		text-align: center;
		padding: var(--spacing-xl) * 2;
		background: white;
		border-radius: var(--radius-lg);
		border: 1px solid var(--color-border);
	}

	.empty-icon {
		font-size: 3rem;
		margin-bottom: var(--spacing-md);
	}

	.empty-state h3 {
		margin: 0 0 var(--spacing-xs);
	}

	.empty-state p {
		color: var(--color-text-muted);
		margin: 0 0 var(--spacing-lg);
	}

	.bases-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: var(--spacing-md);
	}

	.base-card {
		display: flex;
		align-items: center;
		gap: var(--spacing-md);
		padding: var(--spacing-md);
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		text-decoration: none;
		color: inherit;
		transition: box-shadow 0.15s, border-color 0.15s;
	}

	.base-card:hover {
		border-color: var(--color-primary);
		box-shadow: var(--shadow-md);
	}

	.base-icon {
		font-size: 2rem;
		width: 50px;
		height: 50px;
		background: var(--color-primary-light);
		border-radius: var(--radius-md);
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.base-info h3 {
		margin: 0 0 var(--spacing-xs);
		font-size: var(--font-size-base);
	}

	.base-meta {
		margin: 0;
		font-size: var(--font-size-sm);
		color: var(--color-text-muted);
	}

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
		padding: var(--spacing-lg);
		width: 100%;
		max-width: 400px;
		box-shadow: var(--shadow-lg);
	}

	.modal h3 {
		margin: 0 0 var(--spacing-md);
	}

	.modal input {
		width: 100%;
		padding: var(--spacing-sm) var(--spacing-md);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: var(--font-size-base);
		margin-bottom: var(--spacing-md);
	}

	.modal input:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: 0 0 0 3px var(--color-primary-light);
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: var(--spacing-sm);
	}
</style>
