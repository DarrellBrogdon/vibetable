<script lang="ts">
	import { onMount } from 'svelte';

	let apiStatus: 'loading' | 'connected' | 'error' = 'loading';
	let apiData: { name?: string; version?: string; status?: string } | null = null;
	let healthData: { status?: string; database?: { status: string } } | null = null;

	const API_URL = import.meta.env.VITE_PUBLIC_API_URL || 'http://localhost:8080';

	onMount(async () => {
		try {
			// Fetch API info
			const [rootRes, healthRes] = await Promise.all([
				fetch(`${API_URL}/`),
				fetch(`${API_URL}/health`)
			]);

			if (rootRes.ok && healthRes.ok) {
				apiData = await rootRes.json();
				healthData = await healthRes.json();
				apiStatus = 'connected';
			} else {
				apiStatus = 'error';
			}
		} catch (e) {
			console.error('Failed to connect to API:', e);
			apiStatus = 'error';
		}
	});
</script>

<div class="container">
	<header class="header">
		<div class="logo">
			<span class="logo-icon">📊</span>
			<h1>VibeTable</h1>
		</div>
		<p class="tagline">An Airtable clone, vibe coded with Go + SvelteKit</p>
	</header>

	<main class="main">
		<section class="status-card">
			<h2>System Status</h2>
			
			<div class="status-grid">
				<div class="status-item">
					<span class="status-label">Frontend</span>
					<span class="status-badge status-connected">Connected</span>
				</div>
				
				<div class="status-item">
					<span class="status-label">Backend API</span>
					{#if apiStatus === 'loading'}
						<span class="status-badge status-loading">Checking...</span>
					{:else if apiStatus === 'connected'}
						<span class="status-badge status-connected">Connected</span>
					{:else}
						<span class="status-badge status-error">Disconnected</span>
					{/if}
				</div>
				
				<div class="status-item">
					<span class="status-label">Database</span>
					{#if apiStatus === 'loading'}
						<span class="status-badge status-loading">Checking...</span>
					{:else if healthData?.database?.status === 'healthy'}
						<span class="status-badge status-connected">Healthy</span>
					{:else}
						<span class="status-badge status-error">Unhealthy</span>
					{/if}
				</div>
			</div>

			{#if apiData}
				<div class="api-info">
					<p><strong>API:</strong> {apiData.name} v{apiData.version}</p>
				</div>
			{/if}
		</section>

		<section class="features-card">
			<h2>Features</h2>
			<ul class="feature-list">
				<li class="feature-item done">
					<span class="checkbox">✓</span>
					Project setup with Docker
				</li>
				<li class="feature-item done">
					<span class="checkbox">✓</span>
					Magic link authentication
				</li>
				<li class="feature-item done">
					<span class="checkbox">✓</span>
					Bases & tables CRUD
				</li>
				<li class="feature-item done">
					<span class="checkbox">✓</span>
					Grid view with inline editing
				</li>
				<li class="feature-item done">
					<span class="checkbox">✓</span>
					Field types (text, number, checkbox, date, select)
				</li>
				<li class="feature-item done">
					<span class="checkbox">✓</span>
					Linked records
				</li>
				<li class="feature-item done">
					<span class="checkbox">✓</span>
					Filtering and sorting
				</li>
				<li class="feature-item done">
					<span class="checkbox">✓</span>
					Kanban view
				</li>
				<li class="feature-item done">
					<span class="checkbox">✓</span>
					Views (save/restore configurations)
				</li>
				<li class="feature-item done">
					<span class="checkbox">✓</span>
					Collaboration (share bases with roles)
				</li>
			</ul>
		</section>
	</main>

	<footer class="footer">
		<p>Built with 🎵 vibes, Go, SvelteKit, and AI assistance</p>
	</footer>
</div>

<style>
	.container {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		max-width: 800px;
		margin: 0 auto;
		padding: var(--spacing-lg);
	}

	.header {
		text-align: center;
		margin-bottom: var(--spacing-xl);
	}

	.logo {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--spacing-sm);
		margin-bottom: var(--spacing-sm);
	}

	.logo-icon {
		font-size: 2.5rem;
	}

	.logo h1 {
		font-size: var(--font-size-2xl);
		color: var(--color-primary);
	}

	.tagline {
		color: var(--color-text-muted);
		font-size: var(--font-size-lg);
	}

	.main {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: var(--spacing-lg);
	}

	.status-card,
	.features-card {
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		padding: var(--spacing-lg);
		box-shadow: var(--shadow-sm);
	}

	.status-card h2,
	.features-card h2 {
		font-size: var(--font-size-lg);
		margin-bottom: var(--spacing-md);
		color: var(--color-gray-700);
	}

	.status-grid {
		display: flex;
		flex-direction: column;
		gap: var(--spacing-sm);
	}

	.status-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--spacing-sm) 0;
		border-bottom: 1px solid var(--color-gray-100);
	}

	.status-item:last-child {
		border-bottom: none;
	}

	.status-label {
		font-weight: 500;
	}

	.status-badge {
		padding: var(--spacing-xs) var(--spacing-sm);
		border-radius: var(--radius-full);
		font-size: var(--font-size-sm);
		font-weight: 500;
	}

	.status-connected {
		background: #dcfce7;
		color: #166534;
	}

	.status-loading {
		background: var(--color-gray-100);
		color: var(--color-gray-600);
	}

	.status-error {
		background: #fee2e2;
		color: #991b1b;
	}

	.api-info {
		margin-top: var(--spacing-md);
		padding-top: var(--spacing-md);
		border-top: 1px solid var(--color-border);
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
	}

	.feature-list {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.feature-item {
		display: flex;
		align-items: center;
		gap: var(--spacing-sm);
		padding: var(--spacing-sm) 0;
		color: var(--color-gray-600);
	}

	.feature-item.done {
		color: var(--color-gray-400);
		text-decoration: line-through;
	}

	.checkbox {
		width: 20px;
		height: 20px;
		border: 2px solid var(--color-gray-300);
		border-radius: var(--radius-sm);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: var(--font-size-xs);
		flex-shrink: 0;
	}

	.feature-item.done .checkbox {
		background: var(--color-success);
		border-color: var(--color-success);
		color: white;
	}

	.footer {
		text-align: center;
		margin-top: var(--spacing-xl);
		padding-top: var(--spacing-lg);
		border-top: 1px solid var(--color-border);
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
	}
</style>
