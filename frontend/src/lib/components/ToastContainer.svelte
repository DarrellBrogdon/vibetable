<script lang="ts">
	import { toastStore, type Toast } from '$lib/stores/toast';
	import { fly } from 'svelte/transition';

	let toasts: Toast[] = [];
	toastStore.subscribe(value => toasts = value);
</script>

<div class="toast-container">
	{#each toasts as toast (toast.id)}
		<div
			class="toast toast-{toast.type}"
			transition:fly={{ y: -20, duration: 200 }}
		>
			<span class="toast-icon">
				{#if toast.type === 'success'}
					&#10003;
				{:else if toast.type === 'error'}
					&#10007;
				{:else}
					&#9432;
				{/if}
			</span>
			<span class="toast-message">{toast.message}</span>
			<button class="toast-close" on:click={() => toastStore.remove(toast.id)}>
				&times;
			</button>
		</div>
	{/each}
</div>

<style>
	.toast-container {
		position: fixed;
		top: var(--spacing-md);
		right: var(--spacing-md);
		z-index: 1000;
		display: flex;
		flex-direction: column;
		gap: var(--spacing-sm);
		max-width: 360px;
	}

	.toast {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		padding: var(--spacing-sm) var(--spacing-md);
		border-radius: var(--radius-md);
		background: white;
		box-shadow: var(--shadow-lg);
		border-left: 4px solid;
	}

	.toast-success {
		border-left-color: var(--color-success, #10b981);
	}

	.toast-error {
		border-left-color: var(--color-error);
	}

	.toast-info {
		border-left-color: var(--color-primary);
	}

	.toast-icon {
		font-size: 16px;
		line-height: 1;
	}

	.toast-success .toast-icon {
		color: var(--color-success, #10b981);
	}

	.toast-error .toast-icon {
		color: var(--color-error);
	}

	.toast-info .toast-icon {
		color: var(--color-primary);
	}

	.toast-message {
		flex: 1;
		font-size: var(--font-size-sm);
		color: var(--color-text);
	}

	.toast-close {
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		font-size: 18px;
		color: var(--color-text-muted);
		line-height: 1;
	}

	.toast-close:hover {
		color: var(--color-text);
	}
</style>
