<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	// Magic links are no longer used - redirect to login
	onMount(() => {
		goto('/login');
	});
</script>

<div class="redirect-container">
	<p>Redirecting to login...</p>
</div>

<style>
	.redirect-container {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--color-text-muted);
	}
</style>
