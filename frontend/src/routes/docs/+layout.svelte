<script lang="ts">
	import { page } from '$app/stores';

	const navSections = [
		{
			title: 'Getting Started',
			items: [
				{ href: '/docs', label: 'Introduction', exact: true },
				{ href: '/docs/quickstart', label: 'Quick Start' },
			]
		},
		{
			title: 'Core Concepts',
			items: [
				{ href: '/docs/bases', label: 'Bases & Tables' },
				{ href: '/docs/fields', label: 'Fields & Field Types' },
				{ href: '/docs/records', label: 'Records & Data' },
			]
		},
		{
			title: 'Views',
			items: [
				{ href: '/docs/views/grid', label: 'Grid View' },
				{ href: '/docs/views/kanban', label: 'Kanban View' },
				{ href: '/docs/views/calendar', label: 'Calendar View' },
				{ href: '/docs/views/gallery', label: 'Gallery View' },
			]
		},
		{
			title: 'Collaboration',
			items: [
				{ href: '/docs/sharing', label: 'Sharing & Permissions' },
				{ href: '/docs/forms', label: 'Forms & Public Access' },
			]
		},
		{
			title: 'Reference',
			items: [
				{ href: '/docs/shortcuts', label: 'Keyboard Shortcuts' },
				{ href: '/docs/faq', label: 'FAQ' },
			]
		}
	];

	function isActive(href: string, exact: boolean = false): boolean {
		if (exact) {
			return $page.url.pathname === href;
		}
		return $page.url.pathname === href || $page.url.pathname.startsWith(href + '/');
	}
</script>

<div class="docs-layout">
	<aside class="docs-sidebar">
		<a href="/" class="docs-logo">
			<span class="logo-icon">📊</span>
			<span class="logo-text">VibeTable</span>
		</a>

		<nav class="docs-nav">
			{#each navSections as section}
				<div class="nav-section">
					<h3 class="nav-section-title">{section.title}</h3>
					<ul class="nav-list">
						{#each section.items as item}
							<li>
								<a
									href={item.href}
									class="nav-link"
									class:active={isActive(item.href, item.exact)}
								>
									{item.label}
								</a>
							</li>
						{/each}
					</ul>
				</div>
			{/each}
		</nav>

		<div class="sidebar-footer">
			<a href="/login" class="back-to-app">
				<span class="icon">←</span>
				Back to App
			</a>
		</div>
	</aside>

	<main class="docs-main">
		<div class="docs-content">
			<slot />
		</div>

		<footer class="docs-footer">
			<p>VibeTable Documentation</p>
		</footer>
	</main>
</div>

<style>
	.docs-layout {
		display: flex;
		min-height: 100vh;
	}

	.docs-sidebar {
		width: 280px;
		background: var(--color-gray-50);
		border-right: 1px solid var(--color-border);
		display: flex;
		flex-direction: column;
		position: fixed;
		top: 0;
		left: 0;
		bottom: 0;
		overflow-y: auto;
	}

	.docs-logo {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		padding: var(--spacing-lg);
		text-decoration: none;
		border-bottom: 1px solid var(--color-border);
	}

	.logo-icon {
		font-size: 1.5rem;
	}

	.logo-text {
		font-size: var(--font-size-lg);
		font-weight: 600;
		color: var(--color-primary);
	}

	.docs-nav {
		flex: 1;
		padding: var(--spacing-md);
		overflow-y: auto;
	}

	.nav-section {
		margin-bottom: var(--spacing-lg);
	}

	.nav-section:last-child {
		margin-bottom: 0;
	}

	.nav-section-title {
		font-size: var(--font-size-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--color-text-muted);
		margin: 0 0 var(--spacing-sm);
		padding: 0 var(--spacing-sm);
	}

	.nav-list {
		list-style: none;
		margin: 0;
		padding: 0;
	}

	.nav-link {
		display: block;
		padding: var(--spacing-sm) var(--spacing-sm);
		border-radius: var(--radius-md);
		color: var(--color-text);
		text-decoration: none;
		font-size: var(--font-size-sm);
		transition: background-color 0.15s, color 0.15s;
	}

	.nav-link:hover {
		background: var(--color-gray-200);
		text-decoration: none;
	}

	.nav-link.active {
		background: var(--color-primary-light);
		color: var(--color-primary);
		font-weight: 500;
	}

	.sidebar-footer {
		padding: var(--spacing-md);
		border-top: 1px solid var(--color-border);
	}

	.back-to-app {
		display: flex;
		align-items: center;
		gap: var(--spacing-xs);
		padding: var(--spacing-sm);
		color: var(--color-text-muted);
		text-decoration: none;
		font-size: var(--font-size-sm);
		border-radius: var(--radius-md);
		transition: background-color 0.15s, color 0.15s;
	}

	.back-to-app:hover {
		background: var(--color-gray-200);
		color: var(--color-text);
		text-decoration: none;
	}

	.back-to-app .icon {
		font-size: var(--font-size-base);
	}

	.docs-main {
		flex: 1;
		margin-left: 280px;
		display: flex;
		flex-direction: column;
		min-height: 100vh;
	}

	.docs-content {
		flex: 1;
		padding: var(--spacing-xl) var(--spacing-xl) var(--spacing-xl) calc(var(--spacing-xl) * 2);
		max-width: 900px;
	}

	.docs-footer {
		padding: var(--spacing-lg) var(--spacing-xl);
		border-top: 1px solid var(--color-border);
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
	}

	.docs-footer p {
		margin: 0;
	}

	/* Global documentation content styles */
	:global(.docs-content h1) {
		font-size: var(--font-size-2xl);
		margin: 0 0 var(--spacing-lg);
		padding-bottom: var(--spacing-md);
		border-bottom: 1px solid var(--color-border);
	}

	:global(.docs-content h2) {
		font-size: var(--font-size-xl);
		margin: var(--spacing-xl) 0 var(--spacing-md);
	}

	:global(.docs-content h3) {
		font-size: var(--font-size-lg);
		margin: var(--spacing-lg) 0 var(--spacing-sm);
	}

	:global(.docs-content p) {
		margin: 0 0 var(--spacing-md);
		line-height: 1.7;
	}

	:global(.docs-content ul),
	:global(.docs-content ol) {
		margin: 0 0 var(--spacing-md);
		padding-left: var(--spacing-lg);
	}

	:global(.docs-content li) {
		margin-bottom: var(--spacing-xs);
		line-height: 1.7;
	}

	:global(.docs-content code) {
		background: var(--color-gray-100);
		padding: 2px 6px;
		border-radius: var(--radius-sm);
		font-family: 'SF Mono', Monaco, 'Courier New', monospace;
		font-size: 0.9em;
	}

	:global(.docs-content pre) {
		background: var(--color-gray-100);
		padding: var(--spacing-md);
		border-radius: var(--radius-md);
		overflow-x: auto;
		margin: 0 0 var(--spacing-md);
	}

	:global(.docs-content pre code) {
		background: none;
		padding: 0;
	}

	:global(.docs-content blockquote) {
		margin: 0 0 var(--spacing-md);
		padding: var(--spacing-md);
		background: var(--color-primary-light);
		border-left: 4px solid var(--color-primary);
		border-radius: 0 var(--radius-md) var(--radius-md) 0;
	}

	:global(.docs-content blockquote p:last-child) {
		margin-bottom: 0;
	}

	:global(.docs-content table) {
		width: 100%;
		border-collapse: collapse;
		margin: 0 0 var(--spacing-md);
	}

	:global(.docs-content th),
	:global(.docs-content td) {
		padding: var(--spacing-sm) var(--spacing-md);
		text-align: left;
		border: 1px solid var(--color-border);
	}

	:global(.docs-content th) {
		background: var(--color-gray-50);
		font-weight: 600;
	}

	:global(.docs-content .tip) {
		background: #e8f5e9;
		border-left: 4px solid var(--color-success);
		padding: var(--spacing-md);
		border-radius: 0 var(--radius-md) var(--radius-md) 0;
		margin: 0 0 var(--spacing-md);
	}

	:global(.docs-content .warning) {
		background: #fff8e1;
		border-left: 4px solid var(--color-warning);
		padding: var(--spacing-md);
		border-radius: 0 var(--radius-md) var(--radius-md) 0;
		margin: 0 0 var(--spacing-md);
	}

	:global(.docs-content .note) {
		background: var(--color-gray-100);
		border-left: 4px solid var(--color-gray-400);
		padding: var(--spacing-md);
		border-radius: 0 var(--radius-md) var(--radius-md) 0;
		margin: 0 0 var(--spacing-md);
	}
</style>
