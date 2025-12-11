<script lang="ts">
	import { onMount } from 'svelte';

	export let topic: string;
	export let position: 'top' | 'bottom' | 'left' | 'right' = 'top';

	const topicLinks: Record<string, { url: string; title: string }> = {
		'fields': { url: '/docs/fields', title: 'Learn about field types' },
		'single-select': { url: '/docs/fields#single-select', title: 'Learn about Single Select fields' },
		'multi-select': { url: '/docs/fields#multi-select', title: 'Learn about Multi Select fields' },
		'linked-record': { url: '/docs/fields#linked-record', title: 'Learn about Linked Records' },
		'views': { url: '/docs/views/grid', title: 'Learn about views' },
		'grid': { url: '/docs/views/grid', title: 'Learn about Grid view' },
		'kanban': { url: '/docs/views/kanban', title: 'Learn about Kanban view' },
		'calendar': { url: '/docs/views/calendar', title: 'Learn about Calendar view' },
		'gallery': { url: '/docs/views/gallery', title: 'Learn about Gallery view' },
		'sharing': { url: '/docs/sharing', title: 'Learn about sharing' },
		'permissions': { url: '/docs/sharing#permission-roles', title: 'Learn about permissions' },
		'forms': { url: '/docs/forms', title: 'Learn about forms' },
		'shortcuts': { url: '/docs/shortcuts', title: 'View keyboard shortcuts' },
		'filters': { url: '/docs/records#filtering-records', title: 'Learn about filtering' },
		'sorting': { url: '/docs/records#sorting-records', title: 'Learn about sorting' },
	};

	let showTooltip = false;
	let buttonEl: HTMLButtonElement;

	const link = topicLinks[topic] || { url: '/docs', title: 'View documentation' };

	function openHelp() {
		window.open(link.url, '_blank');
	}
</script>

<button
	class="contextual-help"
	class:tooltip-visible={showTooltip}
	bind:this={buttonEl}
	on:mouseenter={() => showTooltip = true}
	on:mouseleave={() => showTooltip = false}
	on:focus={() => showTooltip = true}
	on:blur={() => showTooltip = false}
	on:click={openHelp}
	title={link.title}
>
	<svg
		xmlns="http://www.w3.org/2000/svg"
		width="14"
		height="14"
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

	{#if showTooltip}
		<div class="tooltip" class:top={position === 'top'} class:bottom={position === 'bottom'} class:left={position === 'left'} class:right={position === 'right'}>
			{link.title}
		</div>
	{/if}
</button>

<style>
	.contextual-help {
		position: relative;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 18px;
		height: 18px;
		padding: 0;
		background: none;
		border: none;
		border-radius: 50%;
		color: var(--color-text-muted);
		cursor: pointer;
		transition: color 0.15s, background-color 0.15s;
	}

	.contextual-help:hover {
		color: var(--color-primary);
		background: var(--color-primary-light);
	}

	.tooltip {
		position: absolute;
		padding: 6px 10px;
		background: var(--color-gray-800);
		color: white;
		font-size: var(--font-size-xs);
		font-weight: 500;
		border-radius: var(--radius-sm);
		white-space: nowrap;
		z-index: 1000;
		pointer-events: none;
	}

	.tooltip::after {
		content: '';
		position: absolute;
		border: 5px solid transparent;
	}

	.tooltip.top {
		bottom: calc(100% + 8px);
		left: 50%;
		transform: translateX(-50%);
	}

	.tooltip.top::after {
		top: 100%;
		left: 50%;
		transform: translateX(-50%);
		border-top-color: var(--color-gray-800);
	}

	.tooltip.bottom {
		top: calc(100% + 8px);
		left: 50%;
		transform: translateX(-50%);
	}

	.tooltip.bottom::after {
		bottom: 100%;
		left: 50%;
		transform: translateX(-50%);
		border-bottom-color: var(--color-gray-800);
	}

	.tooltip.left {
		right: calc(100% + 8px);
		top: 50%;
		transform: translateY(-50%);
	}

	.tooltip.left::after {
		left: 100%;
		top: 50%;
		transform: translateY(-50%);
		border-left-color: var(--color-gray-800);
	}

	.tooltip.right {
		left: calc(100% + 8px);
		top: 50%;
		transform: translateY(-50%);
	}

	.tooltip.right::after {
		right: 100%;
		top: 50%;
		transform: translateY(-50%);
		border-right-color: var(--color-gray-800);
	}
</style>
