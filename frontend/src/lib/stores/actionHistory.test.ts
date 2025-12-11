import { describe, it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import { actionHistory, type Action } from './actionHistory';

describe('actionHistory store', () => {
	beforeEach(() => {
		actionHistory.clear();
	});

	describe('initial state', () => {
		it('should start with empty past and future', () => {
			const state = get(actionHistory);
			expect(state.past).toEqual([]);
			expect(state.future).toEqual([]);
			expect(state.maxHistorySize).toBe(50);
		});

		it('should not be able to undo or redo initially', () => {
			expect(actionHistory.canUndo()).toBe(false);
			expect(actionHistory.canRedo()).toBe(false);
		});
	});

	describe('push', () => {
		it('should add action to past', () => {
			const action: Action = {
				type: 'record_create',
				tableId: 'table-1',
				timestamp: Date.now(),
				recordId: 'rec-1',
				newData: { field1: 'value1' }
			};

			actionHistory.push(action);

			const state = get(actionHistory);
			expect(state.past).toHaveLength(1);
			expect(state.past[0]).toEqual(action);
			expect(actionHistory.canUndo()).toBe(true);
		});

		it('should clear future stack on new action', () => {
			const action1: Action = {
				type: 'record_create',
				tableId: 'table-1',
				timestamp: Date.now(),
				recordId: 'rec-1'
			};
			const action2: Action = {
				type: 'record_update',
				tableId: 'table-1',
				timestamp: Date.now(),
				recordId: 'rec-1'
			};

			actionHistory.push(action1);
			actionHistory.undo(); // Moves action1 to future
			expect(get(actionHistory).future).toHaveLength(1);

			actionHistory.push(action2);
			expect(get(actionHistory).future).toHaveLength(0);
			expect(get(actionHistory).past).toHaveLength(1);
			expect(get(actionHistory).past[0]).toEqual(action2);
		});

		it('should trim history when exceeding max size', () => {
			// Push more than maxHistorySize actions
			for (let i = 0; i < 55; i++) {
				actionHistory.push({
					type: 'record_update',
					tableId: 'table-1',
					timestamp: Date.now(),
					recordId: `rec-${i}`
				});
			}

			const state = get(actionHistory);
			expect(state.past.length).toBe(50);
			// First 5 actions should be trimmed, so first remaining is rec-5
			expect(state.past[0].recordId).toBe('rec-5');
		});
	});

	describe('undo', () => {
		it('should return null when nothing to undo', () => {
			const result = actionHistory.undo();
			expect(result).toBeNull();
		});

		it('should move action from past to future', () => {
			const action: Action = {
				type: 'record_delete',
				tableId: 'table-1',
				timestamp: Date.now(),
				recordId: 'rec-1',
				previousData: { field1: 'old-value' }
			};

			actionHistory.push(action);
			const undoneAction = actionHistory.undo();

			expect(undoneAction).toEqual(action);
			expect(get(actionHistory).past).toHaveLength(0);
			expect(get(actionHistory).future).toHaveLength(1);
			expect(get(actionHistory).future[0]).toEqual(action);
		});

		it('should undo multiple actions in correct order', () => {
			const actions: Action[] = [
				{ type: 'record_create', tableId: 't1', timestamp: 1, recordId: 'r1' },
				{ type: 'record_update', tableId: 't1', timestamp: 2, recordId: 'r1' },
				{ type: 'record_delete', tableId: 't1', timestamp: 3, recordId: 'r1' }
			];

			actions.forEach(a => actionHistory.push(a));

			expect(actionHistory.undo()?.type).toBe('record_delete');
			expect(actionHistory.undo()?.type).toBe('record_update');
			expect(actionHistory.undo()?.type).toBe('record_create');
			expect(actionHistory.undo()).toBeNull();
		});
	});

	describe('redo', () => {
		it('should return null when nothing to redo', () => {
			const result = actionHistory.redo();
			expect(result).toBeNull();
		});

		it('should move action from future to past', () => {
			const action: Action = {
				type: 'field_create',
				tableId: 'table-1',
				timestamp: Date.now(),
				fieldId: 'field-1',
				newData: { name: 'Status', type: 'single_select' }
			};

			actionHistory.push(action);
			actionHistory.undo();
			const redoneAction = actionHistory.redo();

			expect(redoneAction).toEqual(action);
			expect(get(actionHistory).past).toHaveLength(1);
			expect(get(actionHistory).future).toHaveLength(0);
		});

		it('should redo multiple actions in correct order', () => {
			const actions: Action[] = [
				{ type: 'field_create', tableId: 't1', timestamp: 1, fieldId: 'f1' },
				{ type: 'field_update', tableId: 't1', timestamp: 2, fieldId: 'f1' },
				{ type: 'field_delete', tableId: 't1', timestamp: 3, fieldId: 'f1' }
			];

			actions.forEach(a => actionHistory.push(a));
			actions.forEach(() => actionHistory.undo());

			expect(actionHistory.redo()?.type).toBe('field_create');
			expect(actionHistory.redo()?.type).toBe('field_update');
			expect(actionHistory.redo()?.type).toBe('field_delete');
			expect(actionHistory.redo()).toBeNull();
		});
	});

	describe('canUndo / canRedo', () => {
		it('should track undo availability', () => {
			expect(actionHistory.canUndo()).toBe(false);

			actionHistory.push({ type: 'record_create', tableId: 't1', timestamp: 1 });
			expect(actionHistory.canUndo()).toBe(true);

			actionHistory.undo();
			expect(actionHistory.canUndo()).toBe(false);
		});

		it('should track redo availability', () => {
			expect(actionHistory.canRedo()).toBe(false);

			actionHistory.push({ type: 'record_create', tableId: 't1', timestamp: 1 });
			expect(actionHistory.canRedo()).toBe(false);

			actionHistory.undo();
			expect(actionHistory.canRedo()).toBe(true);

			actionHistory.redo();
			expect(actionHistory.canRedo()).toBe(false);
		});
	});

	describe('clearForTable', () => {
		it('should remove actions for specific table only', () => {
			actionHistory.push({ type: 'record_create', tableId: 'table-1', timestamp: 1 });
			actionHistory.push({ type: 'record_create', tableId: 'table-2', timestamp: 2 });
			actionHistory.push({ type: 'record_update', tableId: 'table-1', timestamp: 3 });
			actionHistory.undo(); // Move last action to future

			actionHistory.clearForTable('table-1');

			const state = get(actionHistory);
			expect(state.past).toHaveLength(1);
			expect(state.past[0].tableId).toBe('table-2');
			expect(state.future).toHaveLength(0);
		});

		it('should do nothing if table has no actions', () => {
			actionHistory.push({ type: 'record_create', tableId: 'table-1', timestamp: 1 });

			actionHistory.clearForTable('table-999');

			expect(get(actionHistory).past).toHaveLength(1);
		});
	});

	describe('clear', () => {
		it('should reset all history', () => {
			actionHistory.push({ type: 'record_create', tableId: 't1', timestamp: 1 });
			actionHistory.push({ type: 'record_update', tableId: 't1', timestamp: 2 });
			actionHistory.undo();

			actionHistory.clear();

			const state = get(actionHistory);
			expect(state.past).toEqual([]);
			expect(state.future).toEqual([]);
			expect(state.maxHistorySize).toBe(50);
		});
	});

	describe('action types', () => {
		it('should handle all action types', () => {
			const actionTypes: Action['type'][] = [
				'record_create',
				'record_update',
				'record_delete',
				'field_create',
				'field_update',
				'field_delete'
			];

			actionTypes.forEach((type, i) => {
				actionHistory.push({
					type,
					tableId: 'table-1',
					timestamp: i
				});
			});

			expect(get(actionHistory).past).toHaveLength(6);
		});

		it('should store previousData and newData', () => {
			const action: Action = {
				type: 'record_update',
				tableId: 'table-1',
				timestamp: Date.now(),
				recordId: 'rec-1',
				previousData: { name: 'Old Name', status: 'draft' },
				newData: { name: 'New Name', status: 'published' }
			};

			actionHistory.push(action);

			const undone = actionHistory.undo();
			expect(undone?.previousData).toEqual({ name: 'Old Name', status: 'draft' });
			expect(undone?.newData).toEqual({ name: 'New Name', status: 'published' });
		});
	});
});
