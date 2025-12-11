<script lang="ts">
	import { realtime, type UserPresence } from '$lib/stores/realtime';

	// User colors for presence indicators
	const userColors = [
		'#E91E63', // Pink
		'#9C27B0', // Purple
		'#673AB7', // Deep Purple
		'#3F51B5', // Indigo
		'#2196F3', // Blue
		'#00BCD4', // Cyan
		'#009688', // Teal
		'#4CAF50', // Green
		'#FF9800', // Orange
		'#FF5722', // Deep Orange
	];

	function getUserColor(userId: string): string {
		// Generate a consistent color based on user ID
		let hash = 0;
		for (let i = 0; i < userId.length; i++) {
			hash = userId.charCodeAt(i) + ((hash << 5) - hash);
		}
		return userColors[Math.abs(hash) % userColors.length];
	}

	function getInitials(presence: UserPresence): string {
		if (presence.name) {
			const parts = presence.name.split(' ');
			if (parts.length >= 2) {
				return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
			}
			return presence.name.substring(0, 2).toUpperCase();
		}
		return presence.email.substring(0, 2).toUpperCase();
	}

	function getDisplayName(presence: UserPresence): string {
		return presence.name || presence.email.split('@')[0];
	}

	$: connected = $realtime.connected;
	$: presenceArray = Array.from($realtime.presence.values());
	$: visibleUsers = presenceArray.slice(0, 5);
	$: overflowCount = Math.max(0, presenceArray.length - 5);
</script>

<div class="presence-indicator">
	{#if connected}
		<div class="presence-avatars">
			{#each visibleUsers as presence (presence.userId)}
				<div
					class="avatar"
					style="background-color: {getUserColor(presence.userId)}"
					title="{getDisplayName(presence)} ({presence.email})"
				>
					{getInitials(presence)}
				</div>
			{/each}
			{#if overflowCount > 0}
				<div class="avatar overflow" title="{overflowCount} more users online">
					+{overflowCount}
				</div>
			{/if}
		</div>
		<div class="connection-status connected" title="Connected - Real-time updates active">
			<span class="status-dot"></span>
		</div>
	{:else if $realtime.connecting}
		<div class="connection-status connecting" title="Connecting...">
			<span class="status-dot"></span>
		</div>
	{:else}
		<div class="connection-status disconnected" title="Disconnected - Changes may not sync">
			<span class="status-dot"></span>
		</div>
	{/if}
</div>

<style>
	.presence-indicator {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.presence-avatars {
		display: flex;
		flex-direction: row-reverse;
	}

	.avatar {
		width: 28px;
		height: 28px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 10px;
		font-weight: 600;
		color: white;
		border: 2px solid white;
		margin-left: -8px;
		cursor: default;
		transition: transform 0.15s ease;
	}

	.avatar:hover {
		transform: scale(1.1);
		z-index: 10;
	}

	.avatar:last-child {
		margin-left: 0;
	}

	.avatar.overflow {
		background-color: #607d8b;
		font-size: 9px;
	}

	.connection-status {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 12px;
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}

	.connected .status-dot {
		background-color: #4caf50;
		box-shadow: 0 0 4px rgba(76, 175, 80, 0.5);
	}

	.connecting .status-dot {
		background-color: #ff9800;
		animation: pulse 1s infinite;
	}

	.disconnected .status-dot {
		background-color: #9e9e9e;
	}

	@keyframes pulse {
		0%,
		100% {
			opacity: 1;
		}
		50% {
			opacity: 0.5;
		}
	}
</style>
