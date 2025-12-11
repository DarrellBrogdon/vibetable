<script lang="ts">
	const features = [
		{
			icon: '📊',
			title: 'Bases & Tables',
			description: 'Organize your data into bases (like workbooks) containing multiple tables.',
			link: '/docs/bases'
		},
		{
			icon: '📝',
			title: 'Rich Field Types',
			description: 'Text, numbers, dates, checkboxes, select options, linked records, and more.',
			link: '/docs/fields'
		},
		{
			icon: '📋',
			title: 'Multiple Views',
			description: 'Grid, Kanban, Calendar, and Gallery views to see your data your way.',
			link: '/docs/views/grid'
		},
		{
			icon: '👥',
			title: 'Collaboration',
			description: 'Share bases with your team and work together in real-time.',
			link: '/docs/sharing'
		}
	];
</script>

<svelte:head>
	<title>Documentation - VibeTable</title>
</svelte:head>

<h1>Welcome to VibeTable</h1>

<p class="lead">
	VibeTable is a powerful, flexible database tool that helps you organize anything.
	Whether you're tracking projects, managing inventory, or planning events,
	VibeTable adapts to your workflow.
</p>

<div class="feature-grid">
	{#each features as feature}
		<a href={feature.link} class="feature-card">
			<span class="feature-icon">{feature.icon}</span>
			<h3>{feature.title}</h3>
			<p>{feature.description}</p>
		</a>
	{/each}
</div>

<h2>What is VibeTable?</h2>

<p>
	VibeTable combines the simplicity of a spreadsheet with the power of a database.
	Unlike traditional spreadsheets, VibeTable lets you:
</p>

<ul>
	<li><strong>Define field types</strong> - Ensure data consistency with typed columns</li>
	<li><strong>Link records</strong> - Create relationships between tables</li>
	<li><strong>View data multiple ways</strong> - Switch between Grid, Kanban, Calendar, and Gallery views</li>
	<li><strong>Collaborate in real-time</strong> - Work with your team simultaneously</li>
	<li><strong>Share publicly</strong> - Create public views and forms for data collection</li>
</ul>

<h2>Quick Links</h2>

<div class="quick-links">
	<a href="/docs/quickstart" class="quick-link">
		<span class="icon">🚀</span>
		<div>
			<strong>Quick Start Guide</strong>
			<span>Get up and running in 5 minutes</span>
		</div>
	</a>
	<a href="/docs/shortcuts" class="quick-link">
		<span class="icon">⌨️</span>
		<div>
			<strong>Keyboard Shortcuts</strong>
			<span>Master navigation and editing</span>
		</div>
	</a>
	<a href="/docs/faq" class="quick-link">
		<span class="icon">❓</span>
		<div>
			<strong>FAQ</strong>
			<span>Common questions answered</span>
		</div>
	</a>
</div>

<style>
	.lead {
		font-size: var(--font-size-lg);
		color: var(--color-text-muted);
		line-height: 1.6;
		margin-bottom: var(--spacing-xl);
	}

	.feature-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: var(--spacing-md);
		margin-bottom: var(--spacing-xl);
	}

	.feature-card {
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: var(--spacing-lg);
		text-decoration: none;
		color: inherit;
		transition: border-color 0.15s, box-shadow 0.15s;
	}

	.feature-card:hover {
		border-color: var(--color-primary);
		box-shadow: var(--shadow-md);
		text-decoration: none;
	}

	.feature-icon {
		font-size: 2rem;
		display: block;
		margin-bottom: var(--spacing-sm);
	}

	.feature-card h3 {
		margin: 0 0 var(--spacing-xs);
		font-size: var(--font-size-base);
	}

	.feature-card p {
		margin: 0;
		font-size: var(--font-size-sm);
		color: var(--color-text-muted);
		line-height: 1.5;
	}

	.quick-links {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-sm);
	}

	.quick-link {
		display: flex;
		align-items: center;
		gap: var(--spacing-md);
		padding: var(--spacing-md);
		background: var(--color-gray-50);
		border-radius: var(--radius-md);
		text-decoration: none;
		color: inherit;
		transition: background-color 0.15s;
	}

	.quick-link:hover {
		background: var(--color-gray-100);
		text-decoration: none;
	}

	.quick-link .icon {
		font-size: 1.5rem;
	}

	.quick-link div {
		display: flex;
		flex-direction: column;
	}

	.quick-link strong {
		font-weight: 500;
	}

	.quick-link span {
		font-size: var(--font-size-sm);
		color: var(--color-text-muted);
	}
</style>
