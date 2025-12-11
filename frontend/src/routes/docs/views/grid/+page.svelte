<svelte:head>
	<title>Grid View - VibeTable Documentation</title>
</svelte:head>

<h1>Grid View</h1>

<p>The Grid view is VibeTable's default view - a powerful spreadsheet-like interface for viewing and editing your data. If you're familiar with Excel or Google Sheets, you'll feel right at home.</p>

<h2>Overview</h2>

<p>Grid view displays your data in a familiar rows-and-columns format:</p>

<ul>
	<li><strong>Rows</strong> - Each row is a record</li>
	<li><strong>Columns</strong> - Each column is a field</li>
	<li><strong>Cells</strong> - The intersection where you enter and view data</li>
</ul>

<p>This view is ideal for:</p>
<ul>
	<li>Data entry and bulk editing</li>
	<li>Reviewing all your data at once</li>
	<li>Quick scanning and comparison</li>
	<li>Copy/paste operations</li>
</ul>

<h2>Navigation</h2>

<h3>Mouse Navigation</h3>

<ul>
	<li><strong>Click</strong> a cell to select it</li>
	<li><strong>Double-click</strong> to edit a cell</li>
	<li><strong>Scroll</strong> to navigate large tables</li>
</ul>

<h3>Keyboard Navigation</h3>

<table>
	<thead>
		<tr>
			<th>Key</th>
			<th>Action</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td><kbd>↑</kbd> <kbd>↓</kbd> <kbd>←</kbd> <kbd>→</kbd></td>
			<td>Move between cells</td>
		</tr>
		<tr>
			<td><kbd>Tab</kbd></td>
			<td>Move to next cell (right)</td>
		</tr>
		<tr>
			<td><kbd>Shift</kbd> + <kbd>Tab</kbd></td>
			<td>Move to previous cell (left)</td>
		</tr>
		<tr>
			<td><kbd>Enter</kbd></td>
			<td>Enter edit mode / Move down</td>
		</tr>
		<tr>
			<td><kbd>Escape</kbd></td>
			<td>Exit edit mode / Deselect</td>
		</tr>
		<tr>
			<td><kbd>Home</kbd></td>
			<td>Go to first cell in row</td>
		</tr>
		<tr>
			<td><kbd>End</kbd></td>
			<td>Go to last cell in row</td>
		</tr>
	</tbody>
</table>

<h2>Editing Data</h2>

<h3>Editing Cells</h3>

<p>There are several ways to edit a cell:</p>

<ol>
	<li><strong>Start typing</strong> - Replace the current value</li>
	<li><strong>Press Enter</strong> - Edit the existing value</li>
	<li><strong>Double-click</strong> - Edit the existing value</li>
	<li><strong>Press F2</strong> - Edit the existing value</li>
</ol>

<p>When editing:</p>
<ul>
	<li>Press <kbd>Enter</kbd> to save and move down</li>
	<li>Press <kbd>Tab</kbd> to save and move right</li>
	<li>Press <kbd>Escape</kbd> to cancel changes</li>
	<li>Click outside the cell to save</li>
</ul>

<h3>Clearing Cells</h3>

<ul>
	<li>Select a cell and press <kbd>Delete</kbd> or <kbd>Backspace</kbd></li>
	<li>Select multiple cells to clear them all</li>
</ul>

<h2>Columns</h2>

<h3>Resizing Columns</h3>

<ol>
	<li>Hover over the column border in the header</li>
	<li>Click and drag to resize</li>
	<li>Double-click the border to auto-fit to content</li>
</ol>

<h3>Reordering Columns</h3>

<ol>
	<li>Click and hold a column header</li>
	<li>Drag left or right to the desired position</li>
	<li>Release to drop</li>
</ol>

<h3>Hiding Columns</h3>

<ol>
	<li>Click the column header</li>
	<li>Select <strong>Hide field</strong></li>
</ol>

<p>To show hidden columns:</p>
<ol>
	<li>Click the <strong>Fields</strong> button in the toolbar</li>
	<li>Toggle visibility for hidden fields</li>
</ol>

<div class="note">
	<strong>Note:</strong> Hidden fields are only hidden in the current view. The data still exists and appears in other views.
</div>

<h2>Rows</h2>

<h3>Row Height</h3>

<p>Adjust row height for better visibility:</p>

<ol>
	<li>Click the view settings icon</li>
	<li>Choose row height: Short, Medium, Tall, or Extra Tall</li>
</ol>

<h3>Selecting Rows</h3>

<table>
	<thead>
		<tr>
			<th>Action</th>
			<th>How</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Select one row</td>
			<td>Click the row number</td>
		</tr>
		<tr>
			<td>Select range</td>
			<td><kbd>Shift</kbd> + Click row numbers</td>
		</tr>
		<tr>
			<td>Select multiple</td>
			<td><kbd>Cmd/Ctrl</kbd> + Click row numbers</td>
		</tr>
		<tr>
			<td>Select all</td>
			<td>Click corner checkbox or <kbd>Cmd/Ctrl</kbd> + <kbd>A</kbd></td>
		</tr>
	</tbody>
</table>

<h3>Row Actions</h3>

<p>With rows selected, you can:</p>

<ul>
	<li><strong>Delete</strong> - Press <kbd>Delete</kbd> or right-click → Delete</li>
	<li><strong>Duplicate</strong> - Right-click → Duplicate</li>
	<li><strong>Copy</strong> - <kbd>Cmd/Ctrl</kbd> + <kbd>C</kbd></li>
</ul>

<h2>Toolbar Features</h2>

<h3>Search</h3>

<p>Find specific data quickly:</p>

<ol>
	<li>Press <kbd>Cmd/Ctrl</kbd> + <kbd>F</kbd> or click the search icon</li>
	<li>Type your search term</li>
	<li>Matching cells are highlighted</li>
</ol>

<h3>Filter</h3>

<p>Show only records matching criteria:</p>

<ol>
	<li>Click <strong>Filter</strong> in the toolbar</li>
	<li>Add filter conditions</li>
	<li>Only matching records appear</li>
</ol>

<p>See <a href="/docs/records">Records & Data</a> for more on filtering.</p>

<h3>Sort</h3>

<p>Order records by field values:</p>

<ol>
	<li>Click <strong>Sort</strong> in the toolbar</li>
	<li>Choose field and direction (A→Z or Z→A)</li>
	<li>Add multiple sort levels</li>
</ol>

<h3>Group</h3>

<p>Organize records by a field value:</p>

<ol>
	<li>Click <strong>Group</strong> in the toolbar</li>
	<li>Select a field to group by</li>
	<li>Records are organized into collapsible groups</li>
</ol>

<p>Grouping works best with:</p>
<ul>
	<li>Single Select fields (Status, Category)</li>
	<li>Linked Record fields</li>
	<li>Date fields (grouped by day, week, month)</li>
</ul>

<h2>Copy and Paste</h2>

<h3>Copy Cells</h3>

<ol>
	<li>Select cells (single or range)</li>
	<li>Press <kbd>Cmd/Ctrl</kbd> + <kbd>C</kbd></li>
</ol>

<h3>Paste Cells</h3>

<ol>
	<li>Select destination cell</li>
	<li>Press <kbd>Cmd/Ctrl</kbd> + <kbd>V</kbd></li>
</ol>

<h3>Paste from Spreadsheets</h3>

<p>Copy data from Excel, Google Sheets, or other sources:</p>

<ol>
	<li>Copy the data from your spreadsheet</li>
	<li>Select the starting cell in VibeTable</li>
	<li>Paste with <kbd>Cmd/Ctrl</kbd> + <kbd>V</kbd></li>
</ol>

<p>VibeTable will fill multiple rows and columns as needed.</p>

<h2>View Settings</h2>

<p>Customize your Grid view:</p>

<ul>
	<li><strong>Row height</strong> - Compact to expanded views</li>
	<li><strong>Field order</strong> - Arrange columns your way</li>
	<li><strong>Hidden fields</strong> - Focus on what matters</li>
	<li><strong>Filters</strong> - Pre-filter for specific data</li>
	<li><strong>Sorts</strong> - Default ordering</li>
</ul>

<div class="tip">
	<strong>Tip:</strong> Create multiple Grid views with different configurations. For example, "All Tasks" and "My Tasks" can be separate views of the same table with different filters.
</div>

<h2>Related Topics</h2>

<ul>
	<li><a href="/docs/records">Records & Data</a> - More on working with records</li>
	<li><a href="/docs/shortcuts">Keyboard Shortcuts</a> - Full shortcut reference</li>
	<li><a href="/docs/views/kanban">Kanban View</a> - Visual card-based view</li>
</ul>
