<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';

	let ready = false;

	onMount(async () => {
		await authStore.init();
		ready = true;
	});

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
</style>
