<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';

	let ready = false;
	let showWarning = true;

	onMount(async () => {
		await authStore.init();
		ready = true;
	});

	function dismissWarning() {
		showWarning = false;
	}

	$: if (ready && $authStore.initialized && !$authStore.user) {
		goto('/login');
	}
</script>

{#if !ready || $authStore.loading}
	<div class="loading-screen">
		<div class="spinner"></div>
	</div>
{:else if $authStore.user}
	<div class="app-container">
		{#if showWarning}
			<div class="experiment-warning">
				<div class="warning-content">
					<span class="warning-icon">⚠️</span>
					<p>
						<strong>Experimental Project:</strong> VibeTable is a coding experiment.
						There is no guarantee of data preservation or availability.
						This site may be taken down at any time without notice.
						Do not store important data here.
					</p>
				</div>
				<button class="dismiss-btn" on:click={dismissWarning} aria-label="Dismiss warning">
					✕
				</button>
			</div>
		{/if}
		<slot />
	</div>
{/if}

<style>
	.loading-screen {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--color-gray-50);
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 3px solid var(--color-gray-200);
		border-top-color: var(--color-primary);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.app-container {
		min-height: 100vh;
		background: var(--color-gray-50);
	}

	.experiment-warning {
		background: #fef3c7;
		border-bottom: 1px solid #f59e0b;
		padding: var(--spacing-sm) var(--spacing-md);
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--spacing-md);
	}

	.warning-content {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		flex: 1;
	}

	.warning-icon {
		font-size: var(--font-size-lg);
		flex-shrink: 0;
	}

	.warning-content p {
		margin: 0;
		font-size: var(--font-size-sm);
		color: #92400e;
		line-height: 1.4;
	}

	.warning-content strong {
		color: #78350f;
	}

	.dismiss-btn {
		background: none;
		border: none;
		color: #92400e;
		font-size: var(--font-size-lg);
		padding: var(--spacing-xs);
		cursor: pointer;
		opacity: 0.7;
		transition: opacity 0.15s;
		flex-shrink: 0;
	}

	.dismiss-btn:hover {
		opacity: 1;
	}
</style>
