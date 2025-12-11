<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/api/client';
	import { authStore } from '$lib/stores/auth';

	let email = '';
	let password = '';
	let loading = false;
	let error = '';

	async function handleSubmit() {
		if (!email || !password) return;

		loading = true;
		error = '';

		try {
			const response = await auth.login(email, password);
			// Update auth store (this also stores the token)
			await authStore.login(response.token, response.user);
			// Redirect to bases
			goto('/bases');
		} catch (e: any) {
			if (e.code === 'invalid_credentials') {
				error = 'Invalid email or password';
			} else if (e.code === 'password_too_short') {
				error = 'Password must be at least 8 characters';
			} else {
				error = e.message || 'Failed to sign in';
			}
		} finally {
			loading = false;
		}
	}
</script>

<div class="login-container">
	<div class="login-card">
		<div class="logo">
			<span class="logo-icon">📊</span>
			<h1>VibeTable</h1>
		</div>

		<div class="experiment-warning">
			<span class="warning-icon">⚠️</span>
			<p>
				<strong>Experimental Project:</strong> This is a coding experiment with no guarantee of data preservation.
				The site may be taken down at any time.
			</p>
		</div>

		<form on:submit|preventDefault={handleSubmit}>
			<h2>Sign in to your account</h2>
			<p class="subtitle">Enter your email and password to continue</p>

			{#if error}
				<div class="error-message">{error}</div>
			{/if}

			<div class="form-group">
				<label for="email">Email address</label>
				<input
					type="email"
					id="email"
					bind:value={email}
					placeholder="you@example.com"
					required
					disabled={loading}
				/>
			</div>

			<div class="form-group">
				<label for="password">Password</label>
				<input
					type="password"
					id="password"
					bind:value={password}
					placeholder="Enter your password"
					required
					disabled={loading}
					minlength="8"
				/>
			</div>

			<button type="submit" class="primary" disabled={loading || !email || !password}>
				{loading ? 'Signing in...' : 'Sign in'}
			</button>

			<div class="forgot-password">
				<a href="/auth/forgot-password">Forgot your password?</a>
			</div>

			<p class="new-account-hint">
				Don't have an account? Just enter your email and password to create one.
			</p>
		</form>
	</div>
</div>

<style>
	.login-container {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: var(--spacing-lg);
		background: var(--color-gray-50);
	}

	.login-card {
		background: var(--color-surface);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-lg);
		padding: var(--spacing-xl);
		width: 100%;
		max-width: 400px;
	}

	.logo {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--spacing-sm);
		margin-bottom: var(--spacing-xl);
	}

	.logo-icon {
		font-size: 2rem;
	}

	.logo h1 {
		font-size: var(--font-size-xl);
		color: var(--color-primary);
		margin: 0;
	}

	h2 {
		font-size: var(--font-size-lg);
		margin: 0 0 var(--spacing-xs);
		text-align: center;
	}

	.subtitle {
		color: var(--color-text-muted);
		text-align: center;
		margin-bottom: var(--spacing-lg);
	}

	.form-group {
		margin-bottom: var(--spacing-md);
	}

	label {
		display: block;
		font-size: var(--font-size-sm);
		font-weight: 500;
		margin-bottom: var(--spacing-xs);
	}

	input {
		width: 100%;
		padding: var(--spacing-sm) var(--spacing-md);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: var(--font-size-base);
		transition: border-color 0.15s, box-shadow 0.15s;
	}

	input:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: 0 0 0 3px var(--color-primary-light);
	}

	input:disabled {
		background: var(--color-gray-100);
	}

	button {
		width: 100%;
		padding: var(--spacing-sm) var(--spacing-md);
		border: none;
		border-radius: var(--radius-md);
		font-size: var(--font-size-base);
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.15s;
	}

	button.primary {
		background: var(--color-primary);
		color: white;
	}

	button.primary:hover:not(:disabled) {
		background: var(--color-primary-hover);
	}

	button.primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.error-message {
		background: #fee2e2;
		color: #991b1b;
		padding: var(--spacing-sm) var(--spacing-md);
		border-radius: var(--radius-md);
		margin-bottom: var(--spacing-md);
		font-size: var(--font-size-sm);
	}

	.forgot-password {
		text-align: center;
		margin-top: var(--spacing-md);
	}

	.forgot-password a {
		color: var(--color-primary);
		text-decoration: none;
		font-size: var(--font-size-sm);
	}

	.forgot-password a:hover {
		text-decoration: underline;
	}

	.new-account-hint {
		text-align: center;
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
		margin-top: var(--spacing-lg);
		padding-top: var(--spacing-md);
		border-top: 1px solid var(--color-border);
	}

	.experiment-warning {
		background: #fef3c7;
		border: 1px solid #f59e0b;
		border-radius: var(--radius-md);
		padding: var(--spacing-sm) var(--spacing-md);
		margin-bottom: var(--spacing-lg);
		display: flex;
		align-items: flex-start;
		gap: var(--spacing-sm);
	}

	.warning-icon {
		font-size: var(--font-size-base);
		flex-shrink: 0;
		line-height: 1.4;
	}

	.experiment-warning p {
		margin: 0;
		font-size: var(--font-size-sm);
		color: #92400e;
		line-height: 1.4;
	}

	.experiment-warning strong {
		color: #78350f;
	}
</style>
