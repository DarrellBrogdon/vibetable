<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { Field, Record } from '$lib/types';

	export let fields: Field[] = [];
	export let records: Record[] = [];
	export let dateFieldId: string = '';
	export let titleFieldId: string = '';
	export let readonly: boolean = false;

	const dispatch = createEventDispatcher<{
		addRecord: { date: string };
		selectRecord: { id: string };
	}>();

	// Date state
	let currentDate = new Date();
	$: currentYear = currentDate.getFullYear();
	$: currentMonth = currentDate.getMonth();

	// Get date field if not specified
	$: effectiveDateFieldId = dateFieldId || fields.find(f => f.field_type === 'date')?.id || '';

	// Get title field if not specified (first text field)
	$: effectiveTitleFieldId = titleFieldId || fields.find(f => f.field_type === 'text')?.id || '';

	// Calendar computations
	$: firstDayOfMonth = new Date(currentYear, currentMonth, 1);
	$: lastDayOfMonth = new Date(currentYear, currentMonth + 1, 0);
	$: daysInMonth = lastDayOfMonth.getDate();
	$: startingDayOfWeek = firstDayOfMonth.getDay();

	$: monthName = new Intl.DateTimeFormat('en-US', { month: 'long', year: 'numeric' }).format(currentDate);

	// Build calendar grid
	$: calendarDays = buildCalendarDays(currentYear, currentMonth, daysInMonth, startingDayOfWeek);

	function buildCalendarDays(year: number, month: number, days: number, startDay: number): (number | null)[] {
		const result: (number | null)[] = [];
		// Add empty cells for days before the first day
		for (let i = 0; i < startDay; i++) {
			result.push(null);
		}
		// Add the days of the month
		for (let day = 1; day <= days; day++) {
			result.push(day);
		}
		// Add empty cells to fill the last week
		while (result.length % 7 !== 0) {
			result.push(null);
		}
		return result;
	}

	// Group records by date
	$: recordsByDate = groupRecordsByDate(records, effectiveDateFieldId, currentYear, currentMonth);

	function groupRecordsByDate(recs: Record[], fieldId: string, year: number, month: number): Map<number, Record[]> {
		const map = new Map<number, Record[]>();
		if (!fieldId) return map;

		for (const rec of recs) {
			const dateValue = rec.values[fieldId];
			if (!dateValue) continue;

			// Parse date string manually to avoid timezone issues
			// Date strings like "2024-01-15" should be treated as local dates, not UTC
			const parsed = parseDateString(String(dateValue));
			if (!parsed) continue;

			if (parsed.year === year && parsed.month === month) {
				const day = parsed.day;
				if (!map.has(day)) {
					map.set(day, []);
				}
				map.get(day)!.push(rec);
			}
		}
		return map;
	}

	// Parse a date string (YYYY-MM-DD or ISO format) without timezone conversion
	function parseDateString(dateStr: string): { year: number; month: number; day: number } | null {
		if (!dateStr) return null;
		// Handle YYYY-MM-DD format
		const match = dateStr.match(/^(\d{4})-(\d{2})-(\d{2})/);
		if (match) {
			return {
				year: parseInt(match[1], 10),
				month: parseInt(match[2], 10) - 1, // Convert to 0-indexed month
				day: parseInt(match[3], 10)
			};
		}
		return null;
	}

	function getRecordTitle(record: Record): string {
		if (effectiveTitleFieldId && record.values[effectiveTitleFieldId]) {
			return String(record.values[effectiveTitleFieldId]);
		}
		// Fall back to first non-empty text value
		for (const field of fields) {
			if (field.field_type === 'text' && record.values[field.id]) {
				return String(record.values[field.id]);
			}
		}
		return 'Untitled';
	}

	function previousMonth() {
		currentDate = new Date(currentYear, currentMonth - 1, 1);
	}

	function nextMonth() {
		currentDate = new Date(currentYear, currentMonth + 1, 1);
	}

	function goToToday() {
		currentDate = new Date();
	}

	function handleDayClick(day: number | null) {
		if (!day || readonly || !effectiveDateFieldId) return;
		const dateStr = formatDateForApi(currentYear, currentMonth, day);
		dispatch('addRecord', { date: dateStr });
	}

	function handleRecordClick(e: MouseEvent, record: Record) {
		e.stopPropagation();
		dispatch('selectRecord', { id: record.id });
	}

	function formatDateForApi(year: number, month: number, day: number): string {
		const m = String(month + 1).padStart(2, '0');
		const d = String(day).padStart(2, '0');
		return `${year}-${m}-${d}`;
	}

	function isToday(day: number | null): boolean {
		if (!day) return false;
		const today = new Date();
		return today.getFullYear() === currentYear &&
			today.getMonth() === currentMonth &&
			today.getDate() === day;
	}

	const weekDays = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
</script>

<div class="calendar-view">
	{#if !effectiveDateFieldId}
		<div class="no-date-field">
			<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
				<rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
				<line x1="16" y1="2" x2="16" y2="6"/>
				<line x1="8" y1="2" x2="8" y2="6"/>
				<line x1="3" y1="10" x2="21" y2="10"/>
			</svg>
			<p>Add a date field to use Calendar view</p>
		</div>
	{:else}
		<div class="calendar-header">
			<div class="nav-buttons">
				<button class="nav-btn" on:click={previousMonth} title="Previous month">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M15 18l-6-6 6-6"/>
					</svg>
				</button>
				<button class="today-btn" on:click={goToToday}>Today</button>
				<button class="nav-btn" on:click={nextMonth} title="Next month">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M9 18l6-6-6-6"/>
					</svg>
				</button>
			</div>
			<h2 class="month-title">{monthName}</h2>
		</div>

		<div class="calendar-grid">
			<div class="weekday-header">
				{#each weekDays as day}
					<div class="weekday">{day}</div>
				{/each}
			</div>

			<div class="days-grid">
				{#each calendarDays as day, i}
					<div
						class="day-cell"
						class:empty={!day}
						class:today={isToday(day)}
						class:has-records={day && recordsByDate.has(day)}
						on:click={() => handleDayClick(day)}
						on:keydown={(e) => e.key === 'Enter' && handleDayClick(day)}
						role="button"
						tabindex={day ? 0 : -1}
					>
						{#if day}
							<div class="day-number">{day}</div>
							<div class="day-records">
								{#each (recordsByDate.get(day) || []).slice(0, 3) as record}
									<button
										class="record-chip"
										class:colored={record.color}
										style:background-color={record.color ? `var(--color-${record.color}-light, #e5e5e5)` : undefined}
										on:click={(e) => handleRecordClick(e, record)}
									>
										{getRecordTitle(record)}
									</button>
								{/each}
								{#if (recordsByDate.get(day) || []).length > 3}
									<span class="more-records">+{(recordsByDate.get(day) || []).length - 3} more</span>
								{/if}
							</div>
						{/if}
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

<style>
	.calendar-view {
		height: 100%;
		display: flex;
		flex-direction: column;
		background: white;
		overflow: hidden;
	}

	.no-date-field {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		color: #888;
		gap: 16px;
	}

	.no-date-field p {
		font-size: 16px;
		margin: 0;
	}

	.calendar-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 20px;
		border-bottom: 1px solid #e0e0e0;
	}

	.nav-buttons {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.nav-btn {
		background: none;
		border: 1px solid #ddd;
		border-radius: 4px;
		padding: 6px;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #666;
	}

	.nav-btn:hover {
		background: #f5f5f5;
		color: #333;
	}

	.today-btn {
		background: none;
		border: 1px solid #ddd;
		border-radius: 4px;
		padding: 6px 12px;
		font-size: 13px;
		cursor: pointer;
		color: #666;
	}

	.today-btn:hover {
		background: #f5f5f5;
		color: #333;
	}

	.month-title {
		font-size: 20px;
		font-weight: 600;
		color: #333;
		margin: 0;
	}

	.calendar-grid {
		flex: 1;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}

	.weekday-header {
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		border-bottom: 1px solid #e0e0e0;
		background: #f9f9f9;
	}

	.weekday {
		padding: 10px;
		text-align: center;
		font-size: 12px;
		font-weight: 600;
		color: #666;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.days-grid {
		flex: 1;
		display: grid;
		grid-template-columns: repeat(7, 1fr);
		grid-auto-rows: 1fr;
		overflow-y: auto;
	}

	.day-cell {
		border-right: 1px solid #e0e0e0;
		border-bottom: 1px solid #e0e0e0;
		padding: 4px;
		min-height: 100px;
		cursor: pointer;
		transition: background 0.1s;
	}

	.day-cell:nth-child(7n) {
		border-right: none;
	}

	.day-cell.empty {
		background: #f9f9f9;
		cursor: default;
	}

	.day-cell:not(.empty):hover {
		background: #f0f7ff;
	}

	.day-cell.today {
		background: #e6f0ff;
	}

	.day-number {
		font-size: 14px;
		font-weight: 500;
		color: #333;
		padding: 4px 8px;
	}

	.day-cell.today .day-number {
		background: var(--primary-color, #2d7ff9);
		color: white;
		border-radius: 50%;
		width: 28px;
		height: 28px;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0;
	}

	.day-records {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 0 4px;
		overflow: hidden;
	}

	.record-chip {
		background: #e5e7eb;
		border: none;
		border-radius: 4px;
		padding: 3px 8px;
		font-size: 12px;
		text-align: left;
		cursor: pointer;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		transition: all 0.1s;
	}

	.record-chip:hover {
		filter: brightness(0.95);
		transform: translateY(-1px);
	}

	.record-chip.colored {
		border-left: 3px solid currentColor;
	}

	.more-records {
		font-size: 11px;
		color: #888;
		padding: 2px 8px;
	}

	/* Color variables for record chips */
	:global(:root) {
		--color-red-light: #fee2e2;
		--color-orange-light: #ffedd5;
		--color-yellow-light: #fef9c3;
		--color-green-light: #dcfce7;
		--color-blue-light: #dbeafe;
		--color-purple-light: #f3e8ff;
		--color-pink-light: #fce7f3;
		--color-gray-light: #f3f4f6;
	}
</style>
