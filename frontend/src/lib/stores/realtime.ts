import { writable, derived, get } from 'svelte/store';
import { websocket as wsApi } from '$lib/api/client';

// Message types from backend
export const MessageTypes = {
	// Presence
	PRESENCE: 'presence',
	CURSOR: 'cursor',
	USER_JOINED: 'user_joined',
	USER_LEFT: 'user_left',
	PRESENCE_LIST: 'presence_list',

	// Records
	RECORD_CREATED: 'record_created',
	RECORD_UPDATED: 'record_updated',
	RECORD_DELETED: 'record_deleted',

	// Fields
	FIELD_CREATED: 'field_created',
	FIELD_UPDATED: 'field_updated',
	FIELD_DELETED: 'field_deleted',

	// Tables
	TABLE_CREATED: 'table_created',
	TABLE_UPDATED: 'table_updated',
	TABLE_DELETED: 'table_deleted',

	// Views
	VIEW_CREATED: 'view_created',
	VIEW_UPDATED: 'view_updated',
	VIEW_DELETED: 'view_deleted',
} as const;

export interface UserPresence {
	userId: string;
	email: string;
	name?: string;
	tableId?: string;
	viewId?: string;
	cellRef?: {
		recordId: string;
		fieldId: string;
	};
	joinedAt: string;
	updatedAt: string;
}

export interface RealtimeMessage {
	type: string;
	baseId: string;
	tableId?: string;
	recordId?: string;
	fieldId?: string;
	viewId?: string;
	userId: string;
	payload?: any;
	timestamp: string;
}

interface RealtimeState {
	connected: boolean;
	connecting: boolean;
	baseId: string | null;
	presence: Map<string, UserPresence>;
	error: string | null;
}

type MessageHandler = (message: RealtimeMessage) => void;

function createRealtimeStore() {
	let ws: WebSocket | null = null;
	let reconnectTimeout: number | null = null;
	let reconnectAttempts = 0;
	const maxReconnectAttempts = 5;
	const messageHandlers = new Set<MessageHandler>();

	const { subscribe, set, update } = writable<RealtimeState>({
		connected: false,
		connecting: false,
		baseId: null,
		presence: new Map(),
		error: null,
	});

	async function connect(baseId: string) {
		const token = typeof localStorage !== 'undefined' ? localStorage.getItem('token') : null;
		if (!token) {
			update((s) => ({ ...s, error: 'No auth token' }));
			return;
		}

		// Disconnect existing connection
		disconnect();

		update((s) => ({ ...s, connecting: true, baseId, error: null }));

		try {
			// Get a short-lived ticket for WebSocket authentication
			const wsUrl = await wsApi.getUrl(baseId);
			ws = new WebSocket(wsUrl);
		} catch (err) {
			console.error('[Realtime] Failed to get WebSocket ticket:', err);
			update((s) => ({ ...s, connecting: false, error: 'Failed to get WebSocket ticket' }));
			return;
		}

		ws.onopen = () => {
			reconnectAttempts = 0;
			update((s) => ({
				...s,
				connected: true,
				connecting: false,
				error: null,
			}));
			console.log('[Realtime] Connected to base:', baseId);
		};

		ws.onclose = (event) => {
			console.log('[Realtime] Disconnected:', event.code, event.reason);
			update((s) => ({
				...s,
				connected: false,
				connecting: false,
				presence: new Map(),
			}));

			// Attempt reconnect if not intentional
			if (event.code !== 1000 && reconnectAttempts < maxReconnectAttempts) {
				const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
				reconnectAttempts++;
				console.log(`[Realtime] Reconnecting in ${delay}ms (attempt ${reconnectAttempts})`);
				reconnectTimeout = window.setTimeout(() => {
					const state = get({ subscribe });
					if (state.baseId) {
						connect(state.baseId);
					}
				}, delay);
			}
		};

		ws.onerror = (error) => {
			console.error('[Realtime] WebSocket error:', error);
			update((s) => ({ ...s, error: 'Connection error' }));
		};

		ws.onmessage = (event) => {
			try {
				const message: RealtimeMessage = JSON.parse(event.data);
				handleMessage(message);
			} catch (err) {
				console.error('[Realtime] Failed to parse message:', err);
			}
		};
	}

	function disconnect() {
		if (reconnectTimeout) {
			clearTimeout(reconnectTimeout);
			reconnectTimeout = null;
		}
		reconnectAttempts = maxReconnectAttempts; // Prevent auto-reconnect

		if (ws) {
			ws.close(1000, 'Client disconnecting');
			ws = null;
		}

		set({
			connected: false,
			connecting: false,
			baseId: null,
			presence: new Map(),
			error: null,
		});
	}

	function handleMessage(message: RealtimeMessage) {
		switch (message.type) {
			case MessageTypes.PRESENCE_LIST:
				update((s) => {
					const presence = new Map<string, UserPresence>();
					if (Array.isArray(message.payload)) {
						for (const p of message.payload) {
							presence.set(p.userId, p);
						}
					}
					return { ...s, presence };
				});
				break;

			case MessageTypes.USER_JOINED:
				update((s) => {
					const presence = new Map(s.presence);
					if (message.payload) {
						presence.set(message.payload.userId, message.payload);
					}
					return { ...s, presence };
				});
				break;

			case MessageTypes.USER_LEFT:
				update((s) => {
					const presence = new Map(s.presence);
					presence.delete(message.userId);
					return { ...s, presence };
				});
				break;

			case MessageTypes.PRESENCE:
				update((s) => {
					const presence = new Map(s.presence);
					if (message.payload) {
						presence.set(message.payload.userId, message.payload);
					}
					return { ...s, presence };
				});
				break;

			default:
				// Forward all other messages to handlers
				break;
		}

		// Notify all handlers
		messageHandlers.forEach((handler) => handler(message));
	}

	function sendCursor(tableId: string, viewId: string | null, cellRef: { recordId: string; fieldId: string } | null) {
		if (ws && ws.readyState === WebSocket.OPEN) {
			ws.send(
				JSON.stringify({
					type: MessageTypes.CURSOR,
					payload: {
						tableId,
						viewId: viewId || '',
						cellRef,
					},
				})
			);
		}
	}

	function onMessage(handler: MessageHandler) {
		messageHandlers.add(handler);
		return () => {
			messageHandlers.delete(handler);
		};
	}

	// Derived store for presence list
	const presenceList = derived({ subscribe }, ($state) => Array.from($state.presence.values()));

	// Derived store for active user count
	const activeUserCount = derived({ subscribe }, ($state) => $state.presence.size);

	return {
		subscribe,
		connect,
		disconnect,
		sendCursor,
		onMessage,
		presenceList,
		activeUserCount,
	};
}

export const realtime = createRealtimeStore();
