<script lang="ts">
	import { auth } from '$lib/api/client';

	let email = '';
	let loading = false;
	let sent = false;
	let error = '';

	async function handleSubmit() {
		if (!email) return;

		loading = true;
		error = '';

		try {
			await auth.forgotPassword(email);
			sent = true;
		} catch (e: any) {
			error = e.message || 'Failed to send reset link';
		} finally {
			loading = false;
		}
	}
</script>

<div class="forgot-container">
	<div class="forgot-card">
		<div class="logo">
			<span class="logo-icon">📊</span>
			<h1>VibeTable</h1>
		</div>

		{#if sent}
			<div class="success-message">
				<h2>Check your email</h2>
				<p>If an account exists for <strong>{email}</strong>, we've sent a password reset link.</p>
				<p class="hint">The link will expire in 1 hour.</p>
				<a href="/login" class="back-link">Back to sign in</a>
			</div>
		{:else}
			<form on:submit|preventDefault={handleSubmit}>
				<h2>Reset your password</h2>
				<p class="subtitle">Enter your email and we'll send you a link to reset your password</p>

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

				<button type="submit" class="primary" disabled={loading || !email}>
					{loading ? 'Sending...' : 'Send reset link'}
				</button>

				<a href="/login" class="back-link">Back to sign in</a>
			</form>
		{/if}
	</div>
</div>

<style>
	.forgot-container {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: var(--spacing-lg);
		background: var(--color-gray-50);
	}

	.forgot-card {
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

	.success-message {
		text-align: center;
	}

	.success-message p {
		margin: var(--spacing-sm) 0;
	}

	.success-message .hint {
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
	}

	.back-link {
		display: block;
		text-align: center;
		margin-top: var(--spacing-md);
		color: var(--color-primary);
		text-decoration: none;
		font-size: var(--font-size-sm);
	}

	.back-link:hover {
		text-decoration: underline;
	}
</style>
