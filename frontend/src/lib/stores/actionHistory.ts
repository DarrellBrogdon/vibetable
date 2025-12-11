import { writable, get } from 'svelte/store';

export type ActionType =
	| 'record_create'
	| 'record_update'
	| 'record_delete'
	| 'field_create'
	| 'field_update'
	| 'field_delete';

export interface Action {
	type: ActionType;
	tableId: string;
	timestamp: number;
	// For record actions
	recordId?: string;
	// For field actions
	fieldId?: string;
	// Data needed to undo/redo
	previousData?: any;
	newData?: any;
}

interface ActionHistoryState {
	past: Action[];
	future: Action[];
	maxHistorySize: number;
}

function createActionHistory() {
	const { subscribe, set, update } = writable<ActionHistoryState>({
		past: [],
		future: [],
		maxHistorySize: 50
	});

	return {
		subscribe,

		/**
		 * Push a new action onto the history stack.
		 * Clears the future stack (can't redo after new actions).
		 */
		push(action: Action) {
			update(state => {
				const newPast = [...state.past, action];
				// Trim history if it exceeds max size
				if (newPast.length > state.maxHistorySize) {
					newPast.shift();
				}
				return {
					...state,
					past: newPast,
					future: [] // Clear redo stack on new action
				};
			});
		},

		/**
		 * Undo the last action. Returns the action to be undone, or null if nothing to undo.
		 */
		undo(): Action | null {
			const state = get({ subscribe });
			if (state.past.length === 0) return null;

			const action = state.past[state.past.length - 1];
			update(s => ({
				...s,
				past: s.past.slice(0, -1),
				future: [action, ...s.future]
			}));

			return action;
		},

		/**
		 * Redo the last undone action. Returns the action to be redone, or null if nothing to redo.
		 */
		redo(): Action | null {
			const state = get({ subscribe });
			if (state.future.length === 0) return null;

			const action = state.future[0];
			update(s => ({
				...s,
				past: [...s.past, action],
				future: s.future.slice(1)
			}));

			return action;
		},

		/**
		 * Check if undo is available
		 */
		canUndo(): boolean {
			return get({ subscribe }).past.length > 0;
		},

		/**
		 * Check if redo is available
		 */
		canRedo(): boolean {
			return get({ subscribe }).future.length > 0;
		},

		/**
		 * Clear all history for a specific table
		 */
		clearForTable(tableId: string) {
			update(state => ({
				...state,
				past: state.past.filter(a => a.tableId !== tableId),
				future: state.future.filter(a => a.tableId !== tableId)
			}));
		},

		/**
		 * Clear all history
		 */
		clear() {
			set({
				past: [],
				future: [],
				maxHistorySize: 50
			});
		}
	};
}

export const actionHistory = createActionHistory();
