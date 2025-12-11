<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import type { Comment, User } from '$lib/types';
	import { comments as commentsApi } from '$lib/api/client';

	export let recordId: string;
	export let currentUser: User | null = null;
	export let readonly: boolean = false;

	const dispatch = createEventDispatcher<{
		countChange: { count: number };
	}>();

	let commentsList: Comment[] = [];
	let loading = true;
	let error = '';
	let newCommentText = '';
	let submitting = false;
	let replyingTo: string | null = null;
	let replyText = '';
	let editingId: string | null = null;
	let editText = '';

	onMount(() => {
		loadComments();
	});

	async function loadComments() {
		loading = true;
		error = '';
		try {
			const result = await commentsApi.list(recordId);
			commentsList = result.comments;
			dispatch('countChange', { count: countAllComments(commentsList) });
		} catch (e: any) {
			error = e.message || 'Failed to load comments';
		} finally {
			loading = false;
		}
	}

	function countAllComments(list: Comment[]): number {
		let count = list.length;
		for (const c of list) {
			if (c.replies) {
				count += countAllComments(c.replies);
			}
		}
		return count;
	}

	async function submitComment() {
		if (!newCommentText.trim() || submitting) return;

		submitting = true;
		try {
			await commentsApi.create(recordId, newCommentText.trim());
			newCommentText = '';
			await loadComments();
		} catch (e: any) {
			error = e.message || 'Failed to post comment';
		} finally {
			submitting = false;
		}
	}

	async function submitReply(parentId: string) {
		if (!replyText.trim() || submitting) return;

		submitting = true;
		try {
			await commentsApi.create(recordId, replyText.trim(), parentId);
			replyText = '';
			replyingTo = null;
			await loadComments();
		} catch (e: any) {
			error = e.message || 'Failed to post reply';
		} finally {
			submitting = false;
		}
	}

	async function updateComment(commentId: string) {
		if (!editText.trim() || submitting) return;

		submitting = true;
		try {
			await commentsApi.update(commentId, editText.trim());
			editingId = null;
			editText = '';
			await loadComments();
		} catch (e: any) {
			error = e.message || 'Failed to update comment';
		} finally {
			submitting = false;
		}
	}

	async function deleteComment(commentId: string) {
		if (!confirm('Delete this comment?')) return;

		try {
			await commentsApi.delete(commentId);
			await loadComments();
		} catch (e: any) {
			error = e.message || 'Failed to delete comment';
		}
	}

	async function toggleResolved(comment: Comment) {
		try {
			await commentsApi.resolve(comment.id, !comment.is_resolved);
			await loadComments();
		} catch (e: any) {
			error = e.message || 'Failed to update comment';
		}
	}

	function startEdit(comment: Comment) {
		editingId = comment.id;
		editText = comment.content;
	}

	function cancelEdit() {
		editingId = null;
		editText = '';
	}

	function startReply(commentId: string) {
		replyingTo = commentId;
		replyText = '';
	}

	function cancelReply() {
		replyingTo = null;
		replyText = '';
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const minutes = Math.floor(diff / 60000);
		const hours = Math.floor(minutes / 60);
		const days = Math.floor(hours / 24);

		if (minutes < 1) return 'just now';
		if (minutes < 60) return `${minutes}m ago`;
		if (hours < 24) return `${hours}h ago`;
		if (days < 7) return `${days}d ago`;
		return date.toLocaleDateString();
	}

	function getUserInitials(user?: User): string {
		if (!user) return '?';
		if (user.name) {
			return user.name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2);
		}
		return user.email[0].toUpperCase();
	}

	function getUserName(user?: User): string {
		if (!user) return 'Unknown';
		return user.name || user.email.split('@')[0];
	}
</script>

<div class="comment-thread">
	{#if loading}
		<div class="loading">Loading comments...</div>
	{:else if error}
		<div class="error">{error}</div>
	{:else}
		{#if commentsList.length === 0}
			<div class="empty">No comments yet</div>
		{:else}
			<div class="comments-list">
				{#each commentsList as comment (comment.id)}
					<div class="comment" class:resolved={comment.is_resolved}>
						<div class="comment-avatar">
							{getUserInitials(comment.user)}
						</div>
						<div class="comment-content">
							<div class="comment-header">
								<span class="comment-author">{getUserName(comment.user)}</span>
								<span class="comment-time">{formatDate(comment.created_at)}</span>
								{#if comment.is_resolved}
									<span class="resolved-badge">Resolved</span>
								{/if}
							</div>

							{#if editingId === comment.id}
								<div class="edit-form">
									<textarea bind:value={editText} rows="2"></textarea>
									<div class="edit-actions">
										<button class="btn-cancel" on:click={cancelEdit}>Cancel</button>
										<button class="btn-save" on:click={() => updateComment(comment.id)} disabled={submitting}>
											Save
										</button>
									</div>
								</div>
							{:else}
								<div class="comment-body">{comment.content}</div>

								{#if !readonly}
									<div class="comment-actions">
										<button class="action-btn" on:click={() => startReply(comment.id)}>Reply</button>
										<button class="action-btn" on:click={() => toggleResolved(comment)}>
											{comment.is_resolved ? 'Unresolve' : 'Resolve'}
										</button>
										{#if currentUser && comment.user_id === currentUser.id}
											<button class="action-btn" on:click={() => startEdit(comment)}>Edit</button>
											<button class="action-btn danger" on:click={() => deleteComment(comment.id)}>Delete</button>
										{/if}
									</div>
								{/if}
							{/if}

							{#if replyingTo === comment.id}
								<div class="reply-form">
									<textarea
										bind:value={replyText}
										placeholder="Write a reply..."
										rows="2"
									></textarea>
									<div class="reply-actions">
										<button class="btn-cancel" on:click={cancelReply}>Cancel</button>
										<button class="btn-submit" on:click={() => submitReply(comment.id)} disabled={submitting || !replyText.trim()}>
											Reply
										</button>
									</div>
								</div>
							{/if}

							{#if comment.replies && comment.replies.length > 0}
								<div class="replies">
									{#each comment.replies as reply (reply.id)}
										<div class="comment reply">
											<div class="comment-avatar small">
												{getUserInitials(reply.user)}
											</div>
											<div class="comment-content">
												<div class="comment-header">
													<span class="comment-author">{getUserName(reply.user)}</span>
													<span class="comment-time">{formatDate(reply.created_at)}</span>
												</div>
												<div class="comment-body">{reply.content}</div>
												{#if !readonly && currentUser && reply.user_id === currentUser.id}
													<div class="comment-actions">
														<button class="action-btn danger" on:click={() => deleteComment(reply.id)}>Delete</button>
													</div>
												{/if}
											</div>
										</div>
									{/each}
								</div>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}

		{#if !readonly}
			<div class="new-comment">
				<div class="comment-avatar">
					{getUserInitials(currentUser || undefined)}
				</div>
				<div class="new-comment-form">
					<textarea
						bind:value={newCommentText}
						placeholder="Write a comment..."
						rows="2"
						on:keydown={(e) => {
							if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
								submitComment();
							}
						}}
					></textarea>
					<div class="new-comment-actions">
						<span class="hint">Cmd+Enter to submit</span>
						<button
							class="btn-submit"
							on:click={submitComment}
							disabled={submitting || !newCommentText.trim()}
						>
							{submitting ? 'Posting...' : 'Comment'}
						</button>
					</div>
				</div>
			</div>
		{/if}
	{/if}
</div>

<style>
	.comment-thread {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.loading, .error, .empty {
		padding: 20px;
		text-align: center;
		color: var(--color-text-muted);
	}

	.error {
		color: #dc2626;
	}

	.comments-list {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.comment {
		display: flex;
		gap: 12px;
	}

	.comment.resolved {
		opacity: 0.6;
	}

	.comment-avatar {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background: var(--color-primary);
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 12px;
		font-weight: 600;
		flex-shrink: 0;
	}

	.comment-avatar.small {
		width: 24px;
		height: 24px;
		font-size: 10px;
	}

	.comment-content {
		flex: 1;
		min-width: 0;
	}

	.comment-header {
		display: flex;
		align-items: center;
		gap: 8px;
		margin-bottom: 4px;
	}

	.comment-author {
		font-weight: 500;
		font-size: 13px;
	}

	.comment-time {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	.resolved-badge {
		font-size: 11px;
		padding: 2px 6px;
		background: #d1fae5;
		color: #059669;
		border-radius: 4px;
	}

	.comment-body {
		font-size: 14px;
		line-height: 1.5;
		white-space: pre-wrap;
		word-break: break-word;
	}

	.comment-actions {
		display: flex;
		gap: 12px;
		margin-top: 6px;
	}

	.action-btn {
		background: none;
		border: none;
		padding: 0;
		font-size: 12px;
		color: var(--color-text-muted);
		cursor: pointer;
	}

	.action-btn:hover {
		color: var(--color-text);
	}

	.action-btn.danger:hover {
		color: #dc2626;
	}

	.replies {
		margin-top: 12px;
		padding-left: 12px;
		border-left: 2px solid var(--color-border);
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.reply-form, .edit-form {
		margin-top: 8px;
	}

	.reply-form textarea, .edit-form textarea, .new-comment-form textarea {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: 14px;
		resize: vertical;
		min-height: 60px;
	}

	.reply-form textarea:focus, .edit-form textarea:focus, .new-comment-form textarea:focus {
		outline: none;
		border-color: var(--color-primary);
	}

	.reply-actions, .edit-actions, .new-comment-actions {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
		margin-top: 8px;
	}

	.new-comment-actions {
		justify-content: space-between;
		align-items: center;
	}

	.hint {
		font-size: 12px;
		color: var(--color-text-muted);
	}

	.btn-cancel {
		padding: 6px 12px;
		background: white;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		font-size: 13px;
		cursor: pointer;
	}

	.btn-cancel:hover {
		background: var(--color-gray-50);
	}

	.btn-save, .btn-submit {
		padding: 6px 12px;
		background: var(--color-primary);
		color: white;
		border: none;
		border-radius: var(--radius-md);
		font-size: 13px;
		cursor: pointer;
	}

	.btn-save:hover:not(:disabled), .btn-submit:hover:not(:disabled) {
		background: var(--color-primary-hover);
	}

	.btn-save:disabled, .btn-submit:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.new-comment {
		display: flex;
		gap: 12px;
		padding-top: 16px;
		border-top: 1px solid var(--color-border);
	}

	.new-comment-form {
		flex: 1;
	}
</style>
