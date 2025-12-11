import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { get } from 'svelte/store';
import { realtime, MessageTypes, type RealtimeMessage, type UserPresence } from './realtime';

// Mock WebSocket
class MockWebSocket {
	static CONNECTING = 0;
	static OPEN = 1;
	static CLOSING = 2;
	static CLOSED = 3;

	url: string;
	readyState: number = MockWebSocket.CONNECTING;
	onopen: ((event: any) => void) | null = null;
	onclose: ((event: any) => void) | null = null;
	onmessage: ((event: any) => void) | null = null;
	onerror: ((event: any) => void) | null = null;

	constructor(url: string) {
		this.url = url;
		// Simulate async connection
		setTimeout(() => {
			this.readyState = MockWebSocket.OPEN;
			if (this.onopen) this.onopen({ type: 'open' });
		}, 0);
	}

	send = vi.fn();
	close = vi.fn((code?: number, reason?: string) => {
		this.readyState = MockWebSocket.CLOSED;
		if (this.onclose) {
			this.onclose({ code: code || 1000, reason });
		}
	});

	// Helper to simulate receiving a message
	simulateMessage(data: any) {
		if (this.onmessage) {
			this.onmessage({ data: JSON.stringify(data) });
		}
	}

	// Helper to simulate connection error
	simulateError() {
		if (this.onerror) {
			this.onerror({ type: 'error' });
		}
	}

	// Helper to simulate disconnect
	simulateClose(code: number = 1000, reason: string = '') {
		this.readyState = MockWebSocket.CLOSED;
		if (this.onclose) {
			this.onclose({ code, reason });
		}
	}
}

let mockWsInstance: MockWebSocket | null = null;

// Mock localStorage
const localStorageMock = (() => {
	let store: Record<string, string> = {};
	return {
		getItem: vi.fn((key: string) => store[key] || null),
		setItem: vi.fn((key: string, value: string) => { store[key] = value; }),
		removeItem: vi.fn((key: string) => { delete store[key]; }),
		clear: vi.fn(() => { store = {}; }),
	};
})();

describe('realtime store', () => {
	beforeEach(() => {
		vi.useFakeTimers();
		localStorageMock.clear();
		Object.defineProperty(global, 'localStorage', { value: localStorageMock, configurable: true });

		// Mock WebSocket constructor
		(global as any).WebSocket = vi.fn((url: string) => {
			mockWsInstance = new MockWebSocket(url);
			return mockWsInstance;
		});
		(global as any).WebSocket.OPEN = MockWebSocket.OPEN;
		(global as any).WebSocket.CONNECTING = MockWebSocket.CONNECTING;
		(global as any).WebSocket.CLOSING = MockWebSocket.CLOSING;
		(global as any).WebSocket.CLOSED = MockWebSocket.CLOSED;

		realtime.disconnect();
	});

	afterEach(() => {
		vi.useRealTimers();
		vi.clearAllMocks();
		realtime.disconnect();
		mockWsInstance = null;
	});

	describe('initial state', () => {
		it('should start disconnected', () => {
			const state = get(realtime);
			expect(state.connected).toBe(false);
			expect(state.connecting).toBe(false);
			expect(state.baseId).toBeNull();
			expect(state.presence.size).toBe(0);
			expect(state.error).toBeNull();
		});
	});

	describe('connect', () => {
		it('should require auth token', () => {
			realtime.connect('base-1');

			const state = get(realtime);
			expect(state.error).toBe('No auth token');
			expect(state.connected).toBe(false);
		});

		it('should create WebSocket connection with token', async () => {
			localStorageMock.setItem('token', 'test-token');

			realtime.connect('base-1');

			expect((global as any).WebSocket).toHaveBeenCalledWith(
				expect.stringContaining('baseId=base-1')
			);
			expect((global as any).WebSocket).toHaveBeenCalledWith(
				expect.stringContaining('token=test-token')
			);
		});

		it('should update state to connecting', () => {
			localStorageMock.setItem('token', 'test-token');

			realtime.connect('base-1');

			const state = get(realtime);
			expect(state.connecting).toBe(true);
			expect(state.baseId).toBe('base-1');
		});

		it('should update state to connected on open', async () => {
			localStorageMock.setItem('token', 'test-token');

			realtime.connect('base-1');
			await vi.runAllTimersAsync();

			const state = get(realtime);
			expect(state.connected).toBe(true);
			expect(state.connecting).toBe(false);
			expect(state.error).toBeNull();
		});

		it('should handle connection error', async () => {
			localStorageMock.setItem('token', 'test-token');

			realtime.connect('base-1');
			mockWsInstance?.simulateError();

			const state = get(realtime);
			expect(state.error).toBe('Connection error');
		});
	});

	describe('disconnect', () => {
		it('should close WebSocket connection', async () => {
			localStorageMock.setItem('token', 'test-token');
			realtime.connect('base-1');
			await vi.runAllTimersAsync();

			realtime.disconnect();

			expect(mockWsInstance?.close).toHaveBeenCalledWith(1000, 'Client disconnecting');
		});

		it('should reset state on disconnect', async () => {
			localStorageMock.setItem('token', 'test-token');
			realtime.connect('base-1');
			await vi.runAllTimersAsync();

			realtime.disconnect();

			const state = get(realtime);
			expect(state.connected).toBe(false);
			expect(state.connecting).toBe(false);
			expect(state.baseId).toBeNull();
			expect(state.presence.size).toBe(0);
		});
	});

	describe('message handling', () => {
		beforeEach(async () => {
			localStorageMock.setItem('token', 'test-token');
			realtime.connect('base-1');
			await vi.runAllTimersAsync();
		});

		it('should handle presence list message', () => {
			const presenceData: UserPresence[] = [
				{ userId: 'user-1', email: 'a@test.com', joinedAt: '2024-01-01', updatedAt: '2024-01-01' },
				{ userId: 'user-2', email: 'b@test.com', joinedAt: '2024-01-01', updatedAt: '2024-01-01' }
			];

			mockWsInstance?.simulateMessage({
				type: MessageTypes.PRESENCE_LIST,
				baseId: 'base-1',
				userId: 'system',
				payload: presenceData,
				timestamp: new Date().toISOString()
			});

			const state = get(realtime);
			expect(state.presence.size).toBe(2);
			expect(state.presence.get('user-1')?.email).toBe('a@test.com');
		});

		it('should handle user joined message', () => {
			const newUser: UserPresence = {
				userId: 'user-3',
				email: 'c@test.com',
				name: 'User C',
				joinedAt: '2024-01-01',
				updatedAt: '2024-01-01'
			};

			mockWsInstance?.simulateMessage({
				type: MessageTypes.USER_JOINED,
				baseId: 'base-1',
				userId: 'user-3',
				payload: newUser,
				timestamp: new Date().toISOString()
			});

			const state = get(realtime);
			expect(state.presence.has('user-3')).toBe(true);
			expect(state.presence.get('user-3')?.name).toBe('User C');
		});

		it('should handle user left message', () => {
			// First add a user
			mockWsInstance?.simulateMessage({
				type: MessageTypes.PRESENCE_LIST,
				baseId: 'base-1',
				userId: 'system',
				payload: [{ userId: 'user-1', email: 'a@test.com', joinedAt: '2024-01-01', updatedAt: '2024-01-01' }],
				timestamp: new Date().toISOString()
			});

			// Then user leaves
			mockWsInstance?.simulateMessage({
				type: MessageTypes.USER_LEFT,
				baseId: 'base-1',
				userId: 'user-1',
				timestamp: new Date().toISOString()
			});

			const state = get(realtime);
			expect(state.presence.has('user-1')).toBe(false);
		});

		it('should handle presence update message', () => {
			// First add a user
			mockWsInstance?.simulateMessage({
				type: MessageTypes.PRESENCE_LIST,
				baseId: 'base-1',
				userId: 'system',
				payload: [{ userId: 'user-1', email: 'a@test.com', joinedAt: '2024-01-01', updatedAt: '2024-01-01' }],
				timestamp: new Date().toISOString()
			});

			// Then update presence
			mockWsInstance?.simulateMessage({
				type: MessageTypes.PRESENCE,
				baseId: 'base-1',
				userId: 'user-1',
				payload: {
					userId: 'user-1',
					email: 'a@test.com',
					tableId: 'table-1',
					cellRef: { recordId: 'rec-1', fieldId: 'field-1' },
					joinedAt: '2024-01-01',
					updatedAt: '2024-01-02'
				},
				timestamp: new Date().toISOString()
			});

			const state = get(realtime);
			const user = state.presence.get('user-1');
			expect(user?.tableId).toBe('table-1');
			expect(user?.cellRef?.recordId).toBe('rec-1');
		});

		it('should call message handlers for all messages', () => {
			const handler = vi.fn();
			const unsubscribe = realtime.onMessage(handler);

			const message: RealtimeMessage = {
				type: MessageTypes.RECORD_CREATED,
				baseId: 'base-1',
				tableId: 'table-1',
				recordId: 'rec-1',
				userId: 'user-1',
				payload: { values: { field1: 'value1' } },
				timestamp: new Date().toISOString()
			};

			mockWsInstance?.simulateMessage(message);

			expect(handler).toHaveBeenCalledWith(message);
			unsubscribe();
		});

		it('should handle invalid JSON gracefully', () => {
			// Should not throw
			expect(() => {
				if (mockWsInstance?.onmessage) {
					mockWsInstance.onmessage({ data: 'invalid json' });
				}
			}).not.toThrow();
		});
	});

	describe('sendCursor', () => {
		beforeEach(async () => {
			localStorageMock.setItem('token', 'test-token');
			realtime.connect('base-1');
			await vi.runAllTimersAsync();
		});

		it('should send cursor message when connected', () => {
			realtime.sendCursor('table-1', 'view-1', { recordId: 'rec-1', fieldId: 'field-1' });

			expect(mockWsInstance?.send).toHaveBeenCalledWith(
				expect.stringContaining('"type":"cursor"')
			);
		});

		it('should include cell reference in message', () => {
			realtime.sendCursor('table-1', 'view-1', { recordId: 'rec-1', fieldId: 'field-1' });

			const sentData = JSON.parse(mockWsInstance?.send.mock.calls[0][0]);
			expect(sentData.payload.cellRef).toEqual({ recordId: 'rec-1', fieldId: 'field-1' });
		});

		it('should handle null view and cellRef', () => {
			realtime.sendCursor('table-1', null, null);

			const sentData = JSON.parse(mockWsInstance?.send.mock.calls[0][0]);
			expect(sentData.payload.viewId).toBe('');
			expect(sentData.payload.cellRef).toBeNull();
		});
	});

	describe('onMessage', () => {
		beforeEach(async () => {
			localStorageMock.setItem('token', 'test-token');
			realtime.connect('base-1');
			await vi.runAllTimersAsync();
		});

		it('should add and remove message handlers', () => {
			const handler1 = vi.fn();
			const handler2 = vi.fn();

			const unsub1 = realtime.onMessage(handler1);
			const unsub2 = realtime.onMessage(handler2);

			mockWsInstance?.simulateMessage({
				type: MessageTypes.RECORD_UPDATED,
				baseId: 'base-1',
				userId: 'user-1',
				timestamp: new Date().toISOString()
			});

			expect(handler1).toHaveBeenCalledTimes(1);
			expect(handler2).toHaveBeenCalledTimes(1);

			unsub1();

			mockWsInstance?.simulateMessage({
				type: MessageTypes.RECORD_DELETED,
				baseId: 'base-1',
				userId: 'user-1',
				timestamp: new Date().toISOString()
			});

			expect(handler1).toHaveBeenCalledTimes(1); // Not called again
			expect(handler2).toHaveBeenCalledTimes(2);

			unsub2();
		});
	});

	describe('derived stores', () => {
		beforeEach(async () => {
			localStorageMock.setItem('token', 'test-token');
			realtime.connect('base-1');
			await vi.runAllTimersAsync();
		});

		it('presenceList should return array of users', () => {
			mockWsInstance?.simulateMessage({
				type: MessageTypes.PRESENCE_LIST,
				baseId: 'base-1',
				userId: 'system',
				payload: [
					{ userId: 'user-1', email: 'a@test.com', joinedAt: '2024-01-01', updatedAt: '2024-01-01' },
					{ userId: 'user-2', email: 'b@test.com', joinedAt: '2024-01-01', updatedAt: '2024-01-01' }
				],
				timestamp: new Date().toISOString()
			});

			const list = get(realtime.presenceList);
			expect(Array.isArray(list)).toBe(true);
			expect(list).toHaveLength(2);
		});

		it('activeUserCount should return number of users', () => {
			mockWsInstance?.simulateMessage({
				type: MessageTypes.PRESENCE_LIST,
				baseId: 'base-1',
				userId: 'system',
				payload: [
					{ userId: 'user-1', email: 'a@test.com', joinedAt: '2024-01-01', updatedAt: '2024-01-01' },
					{ userId: 'user-2', email: 'b@test.com', joinedAt: '2024-01-01', updatedAt: '2024-01-01' },
					{ userId: 'user-3', email: 'c@test.com', joinedAt: '2024-01-01', updatedAt: '2024-01-01' }
				],
				timestamp: new Date().toISOString()
			});

			const count = get(realtime.activeUserCount);
			expect(count).toBe(3);
		});
	});

	describe('MessageTypes', () => {
		it('should have all expected message types', () => {
			expect(MessageTypes.PRESENCE).toBe('presence');
			expect(MessageTypes.CURSOR).toBe('cursor');
			expect(MessageTypes.USER_JOINED).toBe('user_joined');
			expect(MessageTypes.USER_LEFT).toBe('user_left');
			expect(MessageTypes.PRESENCE_LIST).toBe('presence_list');
			expect(MessageTypes.RECORD_CREATED).toBe('record_created');
			expect(MessageTypes.RECORD_UPDATED).toBe('record_updated');
			expect(MessageTypes.RECORD_DELETED).toBe('record_deleted');
			expect(MessageTypes.FIELD_CREATED).toBe('field_created');
			expect(MessageTypes.FIELD_UPDATED).toBe('field_updated');
			expect(MessageTypes.FIELD_DELETED).toBe('field_deleted');
			expect(MessageTypes.TABLE_CREATED).toBe('table_created');
			expect(MessageTypes.TABLE_UPDATED).toBe('table_updated');
			expect(MessageTypes.TABLE_DELETED).toBe('table_deleted');
			expect(MessageTypes.VIEW_CREATED).toBe('view_created');
			expect(MessageTypes.VIEW_UPDATED).toBe('view_updated');
			expect(MessageTypes.VIEW_DELETED).toBe('view_deleted');
		});
	});
});
