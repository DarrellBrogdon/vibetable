<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';

	const dispatch = createEventDispatcher();

	const isMac = typeof navigator !== 'undefined' && navigator.platform.toUpperCase().indexOf('MAC') >= 0;
	const cmdKey = isMac ? '⌘' : 'Ctrl';

	const shortcuts = [
		{ category: 'Navigation', items: [
			{ keys: ['↑', '↓', '←', '→'], description: 'Move between cells' },
			{ keys: ['Tab'], description: 'Move to next cell' },
			{ keys: ['Shift', 'Tab'], description: 'Move to previous cell' },
			{ keys: ['Enter'], description: 'Edit selected cell' },
			{ keys: ['Esc'], description: 'Cancel editing / Deselect' },
		]},
		{ category: 'Editing', items: [
			{ keys: [cmdKey, 'C'], description: 'Copy cell value' },
			{ keys: [cmdKey, 'V'], description: 'Paste cell value' },
			{ keys: ['Delete'], description: 'Clear cell content' },
			{ keys: [cmdKey, 'Z'], description: 'Undo last change' },
			{ keys: [cmdKey, 'Shift', 'Z'], description: 'Redo last change' },
		]},
		{ category: 'General', items: [
			{ keys: [cmdKey, 'F'], description: 'Search / Quick find' },
			{ keys: [cmdKey, '/'], description: 'Show this help' },
		]},
	];

	function close() {
		dispatch('close');
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			close();
		}
	}

	onMount(() => {
		document.addEventListener('keydown', handleKeydown);
		return () => document.removeEventListener('keydown', handleKeydown);
	});
</script>

<div class="modal-overlay" on:click|self={close}>
	<div class="modal">
		<div class="modal-header">
			<h2>Keyboard Shortcuts</h2>
			<button class="close-btn" on:click={close}>×</button>
		</div>

		<div class="modal-body">
			{#each shortcuts as section}
				<div class="shortcut-section">
					<h3>{section.category}</h3>
					<div class="shortcut-list">
						{#each section.items as shortcut}
							<div class="shortcut-row">
								<div class="keys">
									{#each shortcut.keys as key, i}
										{#if i > 0}<span class="key-separator">+</span>{/if}
										<kbd>{key}</kbd>
									{/each}
								</div>
								<div class="description">{shortcut.description}</div>
							</div>
						{/each}
					</div>
				</div>
			{/each}
		</div>

		<div class="modal-footer">
			<p class="hint">Press <kbd>Esc</kbd> to close</p>
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
		z-index: 300;
	}

	.modal {
		background: white;
		border-radius: var(--radius-lg);
		width: 100%;
		max-width: 480px;
		max-height: 80vh;
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

	.modal-header h2 {
		margin: 0;
		font-size: var(--font-size-lg);
	}

	.close-btn {
		width: 32px;
		height: 32px;
		border: none;
		background: var(--color-gray-100);
		border-radius: 50%;
		cursor: pointer;
		font-size: 1.2rem;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.close-btn:hover {
		background: var(--color-gray-200);
	}

	.modal-body {
		padding: var(--spacing-lg);
		overflow-y: auto;
		flex: 1;
	}

	.shortcut-section {
		margin-bottom: var(--spacing-lg);
	}

	.shortcut-section:last-child {
		margin-bottom: 0;
	}

	.shortcut-section h3 {
		margin: 0 0 var(--spacing-sm);
		font-size: var(--font-size-sm);
		font-weight: 600;
		color: var(--color-text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.shortcut-list {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-xs);
	}

	.shortcut-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--spacing-xs) 0;
	}

	.keys {
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.key-separator {
		color: var(--color-text-muted);
		font-size: var(--font-size-xs);
	}

	kbd {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 24px;
		padding: 2px 8px;
		background: var(--color-gray-100);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		font-family: inherit;
		font-size: var(--font-size-sm);
		font-weight: 500;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
	}

	.description {
		color: var(--color-text);
		font-size: var(--font-size-sm);
	}

	.modal-footer {
		padding: var(--spacing-sm) var(--spacing-lg);
		border-top: 1px solid var(--color-border);
		text-align: center;
	}

	.modal-footer .hint {
		margin: 0;
		font-size: var(--font-size-sm);
		color: var(--color-text-muted);
	}

	.modal-footer kbd {
		margin: 0 4px;
	}
</style>
