<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { publicForms } from '$lib/api/client';
	import type { PublicForm, PublicFormField, FieldType } from '$lib/types';

	let form: PublicForm | null = null;
	let loading = true;
	let error = '';
	let submitting = false;
	let submitted = false;
	let validationErrors: { [fieldId: string]: string } = {};

	// Form values
	let values: { [fieldId: string]: any } = {};

	$: token = $page.params.token;

	onMount(async () => {
		await loadForm();
	});

	async function loadForm() {
		loading = true;
		error = '';
		try {
			form = await publicForms.get(token);
			// Initialize values
			form.fields.forEach(field => {
				if (field.field_type === 'checkbox') {
					values[field.field_id] = false;
				} else if (field.field_type === 'multi_select') {
					values[field.field_id] = [];
				} else {
					values[field.field_id] = '';
				}
			});
		} catch (err: any) {
			if (err.status === 404) {
				error = 'This form does not exist or is no longer active.';
			} else {
				error = err.message || 'Failed to load form';
			}
		} finally {
			loading = false;
		}
	}

	function validateForm(): boolean {
		validationErrors = {};
		let isValid = true;

		if (!form) return false;

		for (const field of form.fields) {
			if (field.is_required) {
				const value = values[field.field_id];
				const isEmpty = value === '' || value === null || value === undefined ||
					(Array.isArray(value) && value.length === 0);

				if (isEmpty) {
					validationErrors[field.field_id] = `${field.label} is required`;
					isValid = false;
				}
			}
		}

		return isValid;
	}

	async function handleSubmit() {
		if (!validateForm()) return;

		submitting = true;
		error = '';

		try {
			await publicForms.submit(token, values);
			submitted = true;

			// Handle redirect if specified
			if (form?.redirect_url) {
				window.location.href = form.redirect_url;
			}
		} catch (err: any) {
			if (err.code === 'form_inactive') {
				error = 'This form is no longer accepting submissions.';
			} else if (err.code === 'validation_error') {
				error = err.message;
			} else {
				error = err.message || 'Failed to submit form. Please try again.';
			}
		} finally {
			submitting = false;
		}
	}

	function getFieldComponent(field: PublicFormField): string {
		switch (field.field_type) {
			case 'text':
			case 'number':
				return 'input';
			case 'checkbox':
				return 'checkbox';
			case 'date':
				return 'date';
			case 'single_select':
				return 'select';
			case 'multi_select':
				return 'multi_select';
			default:
				return 'input';
		}
	}

	function handleMultiSelectToggle(fieldId: string, optionId: string) {
		const current = values[fieldId] || [];
		if (current.includes(optionId)) {
			values[fieldId] = current.filter((id: string) => id !== optionId);
		} else {
			values[fieldId] = [...current, optionId];
		}
	}
</script>

<svelte:head>
	<title>{form?.name || 'Form'} | VibeTable</title>
</svelte:head>

<div class="form-page">
	{#if loading}
		<div class="loading-state">
			<div class="spinner"></div>
			<p>Loading form...</p>
		</div>
	{:else if error && !form}
		<div class="error-state">
			<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="#dc3545" stroke-width="2">
				<circle cx="12" cy="12" r="10"/>
				<path d="M12 8v4M12 16h.01"/>
			</svg>
			<h2>Form Not Available</h2>
			<p>{error}</p>
		</div>
	{:else if submitted}
		<div class="success-state">
			<svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="#28a745" stroke-width="2">
				<circle cx="12" cy="12" r="10"/>
				<path d="M9 12l2 2 4-4"/>
			</svg>
			<h2>Thank You!</h2>
			<p>{form?.success_message}</p>
		</div>
	{:else if form}
		<div class="form-container">
			<div class="form-header">
				<h1>{form.name}</h1>
				{#if form.description}
					<p class="description">{form.description}</p>
				{/if}
			</div>

			{#if error}
				<div class="error-message">{error}</div>
			{/if}

			<form on:submit|preventDefault={handleSubmit}>
				{#each form.fields as field (field.field_id)}
					<div class="form-field" class:has-error={validationErrors[field.field_id]}>
						{#if field.field_type === 'checkbox'}
							<label class="checkbox-field">
								<input
									type="checkbox"
									bind:checked={values[field.field_id]}
									disabled={submitting}
								/>
								<span>{field.label}</span>
								{#if field.is_required}<span class="required">*</span>{/if}
							</label>
						{:else}
							<label>
								<span class="label-text">
									{field.label}
									{#if field.is_required}<span class="required">*</span>{/if}
								</span>

								{#if field.field_type === 'text'}
									<input
										type="text"
										bind:value={values[field.field_id]}
										disabled={submitting}
										placeholder=""
									/>
								{:else if field.field_type === 'number'}
									<input
										type="number"
										bind:value={values[field.field_id]}
										disabled={submitting}
										step="any"
									/>
								{:else if field.field_type === 'date'}
									<input
										type="date"
										bind:value={values[field.field_id]}
										disabled={submitting}
									/>
								{:else if field.field_type === 'single_select'}
									<select bind:value={values[field.field_id]} disabled={submitting}>
										<option value="">Select an option...</option>
										{#each (field.field_options?.options || []) as option}
											<option value={option.id}>{option.name}</option>
										{/each}
									</select>
								{:else if field.field_type === 'multi_select'}
									<div class="multi-select-options">
										{#each (field.field_options?.options || []) as option}
											<label class="option-checkbox">
												<input
													type="checkbox"
													checked={(values[field.field_id] || []).includes(option.id)}
													on:change={() => handleMultiSelectToggle(field.field_id, option.id)}
													disabled={submitting}
												/>
												<span class="option-label" style:background-color={option.color || '#e0e0e0'}>
													{option.name}
												</span>
											</label>
										{/each}
									</div>
								{:else}
									<input
										type="text"
										bind:value={values[field.field_id]}
										disabled={submitting}
									/>
								{/if}
							</label>
						{/if}

						{#if field.help_text}
							<p class="help-text">{field.help_text}</p>
						{/if}

						{#if validationErrors[field.field_id]}
							<p class="error-text">{validationErrors[field.field_id]}</p>
						{/if}
					</div>
				{/each}

				<button type="submit" class="submit-btn" disabled={submitting}>
					{#if submitting}
						<span class="button-spinner"></span>
						Submitting...
					{:else}
						{form.submit_button_text}
					{/if}
				</button>
			</form>
		</div>
	{/if}

	<div class="powered-by">
		Powered by <a href="/" target="_blank">VibeTable</a>
	</div>
</div>

<style>
	.form-page {
		min-height: 100vh;
		background: linear-gradient(135deg, #f5f7fa 0%, #e4e8ec 100%);
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 40px 20px;
	}

	.loading-state,
	.error-state,
	.success-state {
		text-align: center;
		padding: 60px 40px;
		background: white;
		border-radius: 12px;
		box-shadow: 0 4px 24px rgba(0, 0, 0, 0.1);
		max-width: 400px;
	}

	.loading-state p,
	.error-state p,
	.success-state p {
		color: #666;
		margin-top: 16px;
	}

	.loading-state h2,
	.error-state h2,
	.success-state h2 {
		margin: 16px 0 0;
		font-size: 24px;
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 3px solid #e0e0e0;
		border-top-color: var(--primary-color, #2d7ff9);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
		margin: 0 auto;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	.form-container {
		background: white;
		border-radius: 12px;
		box-shadow: 0 4px 24px rgba(0, 0, 0, 0.1);
		width: 100%;
		max-width: 560px;
		padding: 32px;
	}

	.form-header {
		margin-bottom: 24px;
		padding-bottom: 20px;
		border-bottom: 1px solid #e0e0e0;
	}

	.form-header h1 {
		margin: 0;
		font-size: 28px;
		color: #333;
	}

	.form-header .description {
		margin: 12px 0 0;
		color: #666;
		font-size: 15px;
		line-height: 1.5;
	}

	.error-message {
		background: #fee;
		color: #c00;
		padding: 12px 16px;
		border-radius: 6px;
		margin-bottom: 20px;
		font-size: 14px;
	}

	.form-field {
		margin-bottom: 20px;
	}

	.form-field.has-error input,
	.form-field.has-error select {
		border-color: #dc3545;
	}

	.form-field label {
		display: block;
	}

	.label-text {
		display: block;
		font-weight: 500;
		color: #333;
		margin-bottom: 6px;
		font-size: 14px;
	}

	.required {
		color: #dc3545;
		margin-left: 2px;
	}

	.form-field input[type="text"],
	.form-field input[type="number"],
	.form-field input[type="date"],
	.form-field select {
		width: 100%;
		padding: 10px 14px;
		border: 1px solid #ddd;
		border-radius: 6px;
		font-size: 15px;
		transition: border-color 0.2s, box-shadow 0.2s;
	}

	.form-field input:focus,
	.form-field select:focus {
		outline: none;
		border-color: var(--primary-color, #2d7ff9);
		box-shadow: 0 0 0 3px rgba(45, 127, 249, 0.1);
	}

	.checkbox-field {
		display: flex;
		align-items: center;
		gap: 10px;
		cursor: pointer;
	}

	.checkbox-field input {
		width: 18px;
		height: 18px;
		cursor: pointer;
	}

	.checkbox-field span {
		font-weight: 500;
		color: #333;
	}

	.multi-select-options {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.option-checkbox {
		display: flex;
		align-items: center;
		cursor: pointer;
	}

	.option-checkbox input {
		display: none;
	}

	.option-label {
		padding: 6px 12px;
		border-radius: 16px;
		font-size: 13px;
		border: 2px solid transparent;
		transition: all 0.2s;
	}

	.option-checkbox input:checked + .option-label {
		border-color: var(--primary-color, #2d7ff9);
		box-shadow: 0 0 0 1px var(--primary-color, #2d7ff9);
	}

	.help-text {
		margin: 6px 0 0;
		font-size: 12px;
		color: #888;
	}

	.error-text {
		margin: 6px 0 0;
		font-size: 12px;
		color: #dc3545;
	}

	.submit-btn {
		width: 100%;
		padding: 14px 24px;
		background: var(--primary-color, #2d7ff9);
		color: white;
		border: none;
		border-radius: 6px;
		font-size: 16px;
		font-weight: 500;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		transition: background 0.2s;
		margin-top: 8px;
	}

	.submit-btn:hover:not(:disabled) {
		background: #1a6fe8;
	}

	.submit-btn:disabled {
		opacity: 0.7;
		cursor: not-allowed;
	}

	.button-spinner {
		width: 18px;
		height: 18px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top-color: white;
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	.powered-by {
		margin-top: 24px;
		font-size: 12px;
		color: #888;
	}

	.powered-by a {
		color: var(--primary-color, #2d7ff9);
		text-decoration: none;
	}

	.powered-by a:hover {
		text-decoration: underline;
	}
</style>
