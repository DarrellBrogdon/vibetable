<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/api/client';

	let password = '';
	let confirmPassword = '';
	let loading = false;
	let success = false;
	let error = '';

	$: token = $page.url.searchParams.get('token') || '';

	async function handleSubmit() {
		if (!password || !confirmPassword) return;

		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			return;
		}

		if (password.length < 8) {
			error = 'Password must be at least 8 characters';
			return;
		}

		loading = true;
		error = '';

		try {
			await auth.resetPassword(token, password);
			success = true;
		} catch (e: any) {
			if (e.code === 'invalid_token') {
				error = 'This reset link is invalid or has expired';
			} else if (e.code === 'expired_token') {
				error = 'This reset link has expired. Please request a new one.';
			} else if (e.code === 'used_token') {
				error = 'This reset link has already been used. Please request a new one.';
			} else {
				error = e.message || 'Failed to reset password';
			}
		} finally {
			loading = false;
		}
	}
</script>

<div class="reset-container">
	<div class="reset-card">
		<div class="logo">
			<span class="logo-icon">📊</span>
			<h1>VibeTable</h1>
		</div>

		{#if !token}
			<div class="error-state">
				<h2>Invalid Link</h2>
				<p>This password reset link is invalid.</p>
				<a href="/auth/forgot-password" class="back-link">Request a new reset link</a>
			</div>
		{:else if success}
			<div class="success-message">
				<h2>Password Reset</h2>
				<p>Your password has been successfully reset.</p>
				<button class="primary" on:click={() => goto('/login')}>
					Sign in with new password
				</button>
			</div>
		{:else}
			<form on:submit|preventDefault={handleSubmit}>
				<h2>Set new password</h2>
				<p class="subtitle">Enter your new password below</p>

				{#if error}
					<div class="error-message">{error}</div>
				{/if}

				<div class="form-group">
					<label for="password">New password</label>
					<input
						type="password"
						id="password"
						bind:value={password}
						placeholder="Enter new password"
						required
						disabled={loading}
						minlength="8"
					/>
				</div>

				<div class="form-group">
					<label for="confirmPassword">Confirm password</label>
					<input
						type="password"
						id="confirmPassword"
						bind:value={confirmPassword}
						placeholder="Confirm new password"
						required
						disabled={loading}
						minlength="8"
					/>
				</div>

				<p class="password-hint">Password must be at least 8 characters</p>

				<button type="submit" class="primary" disabled={loading || !password || !confirmPassword}>
					{loading ? 'Resetting...' : 'Reset password'}
				</button>

				<a href="/login" class="back-link">Back to sign in</a>
			</form>
		{/if}
	</div>
</div>

<style>
	.reset-container {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: var(--spacing-lg);
		background: var(--color-gray-50);
	}

	.reset-card {
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

	.error-state {
		text-align: center;
	}

	.error-state p {
		color: var(--color-text-muted);
		margin: var(--spacing-sm) 0 var(--spacing-md);
	}

	.success-message {
		text-align: center;
	}

	.success-message p {
		margin: var(--spacing-sm) 0 var(--spacing-lg);
		color: var(--color-text-muted);
	}

	.password-hint {
		font-size: var(--font-size-sm);
		color: var(--color-text-muted);
		margin-bottom: var(--spacing-md);
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
