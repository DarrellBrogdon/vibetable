import { describe, it, expect } from 'vitest';
import { isComputedField, type FieldType } from './index';

describe('types', () => {
	describe('isComputedField', () => {
		it('should return true for formula field', () => {
			expect(isComputedField('formula')).toBe(true);
		});

		it('should return true for rollup field', () => {
			expect(isComputedField('rollup')).toBe(true);
		});

		it('should return true for lookup field', () => {
			expect(isComputedField('lookup')).toBe(true);
		});

		it('should return false for text field', () => {
			expect(isComputedField('text')).toBe(false);
		});

		it('should return false for number field', () => {
			expect(isComputedField('number')).toBe(false);
		});

		it('should return false for checkbox field', () => {
			expect(isComputedField('checkbox')).toBe(false);
		});

		it('should return false for date field', () => {
			expect(isComputedField('date')).toBe(false);
		});

		it('should return false for single_select field', () => {
			expect(isComputedField('single_select')).toBe(false);
		});

		it('should return false for multi_select field', () => {
			expect(isComputedField('multi_select')).toBe(false);
		});

		it('should return false for linked_record field', () => {
			expect(isComputedField('linked_record')).toBe(false);
		});

		it('should return false for attachment field', () => {
			expect(isComputedField('attachment')).toBe(false);
		});

		it('should correctly categorize all field types', () => {
			const computedTypes: FieldType[] = ['formula', 'rollup', 'lookup'];
			const editableTypes: FieldType[] = ['text', 'number', 'checkbox', 'date', 'single_select', 'multi_select', 'linked_record', 'attachment'];

			computedTypes.forEach(type => {
				expect(isComputedField(type)).toBe(true);
			});

			editableTypes.forEach(type => {
				expect(isComputedField(type)).toBe(false);
			});
		});
	});
});
