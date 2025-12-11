<script lang="ts">
	export let size: 'sm' | 'md' = 'md';
</script>

<a
	href="/docs"
	class="help-button"
	class:small={size === 'sm'}
	title="Help & Documentation"
	target="_blank"
>
	<svg
		xmlns="http://www.w3.org/2000/svg"
		width={size === 'sm' ? '16' : '20'}
		height={size === 'sm' ? '16' : '20'}
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="2"
		stroke-linecap="round"
		stroke-linejoin="round"
	>
		<circle cx="12" cy="12" r="10"></circle>
		<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"></path>
		<line x1="12" y1="17" x2="12.01" y2="17"></line>
	</svg>
	<span class="help-text">Help</span>
</a>

<style>
	.help-button {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
		padding: var(--spacing-xs) var(--spacing-sm);
		background: none;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		color: var(--color-text-muted);
		cursor: pointer;
		font-size: var(--font-size-sm);
		text-decoration: none;
		transition: background-color 0.15s, color 0.15s, border-color 0.15s;
	}

	.help-button:hover {
		background: var(--color-gray-50);
		color: var(--color-text);
		text-decoration: none;
	}

	.help-button.small {
		padding: 4px 8px;
	}

	.help-button.small .help-text {
		display: none;
	}

	.help-text {
		font-weight: 500;
	}

	@media (max-width: 768px) {
		.help-text {
			display: none;
		}
	}
</style>
