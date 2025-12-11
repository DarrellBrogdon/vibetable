<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Form, FormField } from '$lib/types';
	import { forms, type FormFieldUpdate } from '$lib/api/client';

	export let form: Form;
	export let tableId: string;

	const dispatch = createEventDispatcher<{
		close: void;
		updated: Form;
		deleted: void;
	}>();

	let saving = false;
	let error = '';
	let copied = false;

	// Local form state
	let name = form.name;
	let description = form.description || '';
	let isActive = form.is_active;
	let successMessage = form.success_message;
	let redirectUrl = form.redirect_url || '';
	let submitButtonText = form.submit_button_text;
	let formFields = [...(form.fields || [])];

	$: publicUrl = form.public_token ? `${window.location.origin}/f/${form.public_token}` : '';

	async function saveForm() {
		saving = true;
		error = '';
		try {
			// Update form settings
			const updatedForm = await forms.update(form.id, {
				name,
				description: description || undefined,
				is_active: isActive,
				success_message: successMessage,
				redirect_url: redirectUrl || undefined,
				submit_button_text: submitButtonText
			});

			// Update form fields
			const fieldUpdates: FormFieldUpdate[] = formFields.map((f, i) => ({
				field_id: f.field_id,
				label: f.label || undefined,
				help_text: f.help_text || undefined,
				is_required: f.is_required,
				is_visible: f.is_visible,
				position: i
			}));
			await forms.updateFields(form.id, fieldUpdates);

			dispatch('updated', { ...updatedForm, fields: formFields });
		} catch (err: any) {
			error = err.message || 'Failed to save form';
		} finally {
			saving = false;
		}
	}

	async function deleteForm() {
		if (!confirm('Are you sure you want to delete this form?')) return;

		saving = true;
		try {
			await forms.delete(form.id);
			dispatch('deleted');
		} catch (err: any) {
			error = err.message || 'Failed to delete form';
			saving = false;
		}
	}

	function copyUrl() {
		navigator.clipboard.writeText(publicUrl);
		copied = true;
		setTimeout(() => copied = false, 2000);
	}

	function toggleFieldVisibility(index: number) {
		formFields[index].is_visible = !formFields[index].is_visible;
	}

	function toggleFieldRequired(index: number) {
		formFields[index].is_required = !formFields[index].is_required;
	}

	function updateFieldLabel(index: number, label: string) {
		formFields[index].label = label || undefined;
	}

	function updateFieldHelpText(index: number, helpText: string) {
		formFields[index].help_text = helpText || undefined;
	}

	function moveField(index: number, direction: 'up' | 'down') {
		const newIndex = direction === 'up' ? index - 1 : index + 1;
		if (newIndex < 0 || newIndex >= formFields.length) return;

		const temp = formFields[index];
		formFields[index] = formFields[newIndex];
		formFields[newIndex] = temp;
		formFields = formFields;
	}
</script>

<div class="form-builder-overlay" on:click|self={() => dispatch('close')}>
	<div class="form-builder">
		<div class="header">
			<h2>Edit Form</h2>
			<button class="close-btn" on:click={() => dispatch('close')}>
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M18 6L6 18M6 6l12 12"/>
				</svg>
			</button>
		</div>

		{#if error}
			<div class="error-message">{error}</div>
		{/if}

		<div class="form-content">
			<div class="section">
				<h3>Form Settings</h3>

				<div class="field-group">
					<label for="name">Form Name</label>
					<input id="name" type="text" bind:value={name} placeholder="Form name" />
				</div>

				<div class="field-group">
					<label for="description">Description (optional)</label>
					<textarea id="description" bind:value={description} placeholder="Describe what this form is for..." rows="2"></textarea>
				</div>

				<div class="field-group">
					<label class="checkbox-label">
						<input type="checkbox" bind:checked={isActive} />
						<span>Form is active (accepting submissions)</span>
					</label>
				</div>

				<div class="field-group">
					<label for="success">Success Message</label>
					<input id="success" type="text" bind:value={successMessage} placeholder="Thank you for your submission!" />
				</div>

				<div class="field-group">
					<label for="redirect">Redirect URL (optional)</label>
					<input id="redirect" type="url" bind:value={redirectUrl} placeholder="https://example.com/thank-you" />
				</div>

				<div class="field-group">
					<label for="button">Submit Button Text</label>
					<input id="button" type="text" bind:value={submitButtonText} placeholder="Submit" />
				</div>
			</div>

			<div class="section">
				<h3>Public Link</h3>
				{#if publicUrl}
					<div class="url-box">
						<input type="text" value={publicUrl} readonly />
						<button class="copy-btn" on:click={copyUrl}>
							{copied ? 'Copied!' : 'Copy'}
						</button>
					</div>
					<p class="url-hint">Share this link to collect form submissions</p>
				{:else}
					<p class="url-hint">Save the form to generate a public link</p>
				{/if}
			</div>

			<div class="section">
				<h3>Form Fields</h3>
				<p class="field-hint">Configure which fields appear on the form and in what order</p>

				<div class="fields-list">
					{#each formFields as field, i}
						<div class="field-config" class:hidden={!field.is_visible}>
							<div class="field-header">
								<div class="field-controls">
									<button
										class="move-btn"
										disabled={i === 0}
										on:click={() => moveField(i, 'up')}
										title="Move up"
									>
										<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M18 15l-6-6-6 6"/>
										</svg>
									</button>
									<button
										class="move-btn"
										disabled={i === formFields.length - 1}
										on:click={() => moveField(i, 'down')}
										title="Move down"
									>
										<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M6 9l6 6 6-6"/>
										</svg>
									</button>
								</div>
								<div class="field-info">
									<span class="field-name">{field.field_name || 'Field'}</span>
									<span class="field-type">{field.field_type}</span>
								</div>
								<div class="field-toggles">
									<label class="toggle-label" title="Show field on form">
										<input type="checkbox" checked={field.is_visible} on:change={() => toggleFieldVisibility(i)} />
										<span>Visible</span>
									</label>
									<label class="toggle-label" title="Require this field">
										<input type="checkbox" checked={field.is_required} on:change={() => toggleFieldRequired(i)} />
										<span>Required</span>
									</label>
								</div>
							</div>
							{#if field.is_visible}
								<div class="field-details">
									<div class="detail-row">
										<label>Label</label>
										<input
											type="text"
											value={field.label || field.field_name || ''}
											on:input={(e) => updateFieldLabel(i, e.currentTarget.value)}
											placeholder={field.field_name}
										/>
									</div>
									<div class="detail-row">
										<label>Help text</label>
										<input
											type="text"
											value={field.help_text || ''}
											on:input={(e) => updateFieldHelpText(i, e.currentTarget.value)}
											placeholder="Optional help text..."
										/>
									</div>
								</div>
							{/if}
						</div>
					{/each}
				</div>
			</div>
		</div>

		<div class="footer">
			<button class="delete-btn" on:click={deleteForm} disabled={saving}>
				Delete Form
			</button>
			<div class="footer-right">
				<button class="cancel-btn" on:click={() => dispatch('close')} disabled={saving}>
					Cancel
				</button>
				<button class="save-btn" on:click={saveForm} disabled={saving}>
					{saving ? 'Saving...' : 'Save Changes'}
				</button>
			</div>
		</div>
	</div>
</div>

<style>
	.form-builder-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}

	.form-builder {
		background: white;
		border-radius: 8px;
		width: 600px;
		max-width: 95vw;
		max-height: 90vh;
		display: flex;
		flex-direction: column;
		box-shadow: 0 4px 24px rgba(0, 0, 0, 0.2);
	}

	.header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 20px;
		border-bottom: 1px solid #e0e0e0;
	}

	.header h2 {
		margin: 0;
		font-size: 18px;
		font-weight: 600;
	}

	.close-btn {
		background: none;
		border: none;
		cursor: pointer;
		padding: 4px;
		color: #666;
	}

	.close-btn:hover {
		color: #333;
	}

	.error-message {
		background: #fee;
		color: #c00;
		padding: 12px 20px;
		font-size: 14px;
	}

	.form-content {
		flex: 1;
		overflow-y: auto;
		padding: 20px;
	}

	.section {
		margin-bottom: 24px;
	}

	.section h3 {
		font-size: 14px;
		font-weight: 600;
		color: #333;
		margin: 0 0 12px 0;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.field-group {
		margin-bottom: 16px;
	}

	.field-group label {
		display: block;
		font-size: 13px;
		font-weight: 500;
		color: #555;
		margin-bottom: 4px;
	}

	.field-group input[type="text"],
	.field-group input[type="url"],
	.field-group textarea {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid #ddd;
		border-radius: 4px;
		font-size: 14px;
	}

	.field-group input:focus,
	.field-group textarea:focus {
		outline: none;
		border-color: var(--primary-color, #2d7ff9);
		box-shadow: 0 0 0 2px rgba(45, 127, 249, 0.1);
	}

	.checkbox-label {
		display: flex !important;
		align-items: center;
		gap: 8px;
		cursor: pointer;
	}

	.checkbox-label input {
		width: auto;
	}

	.url-box {
		display: flex;
		gap: 8px;
	}

	.url-box input {
		flex: 1;
		padding: 8px 12px;
		border: 1px solid #ddd;
		border-radius: 4px;
		font-size: 13px;
		background: #f5f5f5;
	}

	.copy-btn {
		padding: 8px 16px;
		background: var(--primary-color, #2d7ff9);
		color: white;
		border: none;
		border-radius: 4px;
		font-size: 13px;
		cursor: pointer;
		white-space: nowrap;
	}

	.copy-btn:hover {
		background: #1a6fe8;
	}

	.url-hint, .field-hint {
		font-size: 12px;
		color: #888;
		margin-top: 8px;
	}

	.fields-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.field-config {
		background: #f9f9f9;
		border: 1px solid #e0e0e0;
		border-radius: 6px;
		overflow: hidden;
	}

	.field-config.hidden {
		opacity: 0.5;
	}

	.field-header {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 12px;
	}

	.field-controls {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.move-btn {
		background: none;
		border: none;
		padding: 2px;
		cursor: pointer;
		color: #888;
		line-height: 1;
	}

	.move-btn:hover:not(:disabled) {
		color: #333;
	}

	.move-btn:disabled {
		opacity: 0.3;
		cursor: not-allowed;
	}

	.field-info {
		flex: 1;
	}

	.field-name {
		font-weight: 500;
		color: #333;
	}

	.field-type {
		font-size: 12px;
		color: #888;
		margin-left: 8px;
	}

	.field-toggles {
		display: flex;
		gap: 12px;
	}

	.toggle-label {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 12px;
		color: #666;
		cursor: pointer;
	}

	.toggle-label input {
		margin: 0;
	}

	.field-details {
		background: white;
		padding: 12px;
		border-top: 1px solid #e0e0e0;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.detail-row {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.detail-row label {
		font-size: 12px;
		color: #666;
		width: 70px;
		flex-shrink: 0;
	}

	.detail-row input {
		flex: 1;
		padding: 6px 10px;
		border: 1px solid #ddd;
		border-radius: 4px;
		font-size: 13px;
	}

	.footer {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 20px;
		border-top: 1px solid #e0e0e0;
	}

	.footer-right {
		display: flex;
		gap: 8px;
	}

	.delete-btn {
		background: none;
		border: 1px solid #dc3545;
		color: #dc3545;
		padding: 8px 16px;
		border-radius: 4px;
		font-size: 14px;
		cursor: pointer;
	}

	.delete-btn:hover:not(:disabled) {
		background: #dc3545;
		color: white;
	}

	.cancel-btn {
		background: none;
		border: 1px solid #ddd;
		padding: 8px 16px;
		border-radius: 4px;
		font-size: 14px;
		cursor: pointer;
	}

	.cancel-btn:hover:not(:disabled) {
		background: #f5f5f5;
	}

	.save-btn {
		background: var(--primary-color, #2d7ff9);
		color: white;
		border: none;
		padding: 8px 20px;
		border-radius: 4px;
		font-size: 14px;
		cursor: pointer;
	}

	.save-btn:hover:not(:disabled) {
		background: #1a6fe8;
	}

	button:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}
</style>
