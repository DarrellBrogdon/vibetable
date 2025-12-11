<svelte:head>
	<title>Records & Data - VibeTable Documentation</title>
</svelte:head>

<h1>Records & Data</h1>

<p>Records are the individual rows in your tables - each record represents a single item, entry, or piece of data. This guide covers everything about creating, editing, and managing your records.</p>

<h2>Understanding Records</h2>

<p>A record contains values for each field in your table. For example, in a "Contacts" table:</p>

<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Email</th>
			<th>Company</th>
			<th>Status</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Jane Doe</td>
			<td>jane@example.com</td>
			<td>Acme Inc</td>
			<td>Active</td>
		</tr>
	</tbody>
</table>

<p>This entire row is one <strong>record</strong>, with values stored in each field (column).</p>

<h2>Creating Records</h2>

<h3>Quick Add</h3>

<p>The fastest way to add records:</p>

<ol>
	<li>Click in the empty row at the bottom of the table</li>
	<li>Start typing to enter data</li>
	<li>Press <kbd>Tab</kbd> to move to the next field</li>
	<li>Press <kbd>Enter</kbd> to save and start a new record</li>
</ol>

<h3>Add Button</h3>

<ol>
	<li>Click the <strong>+ Add record</strong> button below the table</li>
	<li>A new empty row will be created</li>
	<li>Click any cell to start entering data</li>
</ol>

<h3>Duplicate Record</h3>

<p>To copy an existing record:</p>

<ol>
	<li>Click the row number to select the record</li>
	<li>Right-click or use the record menu</li>
	<li>Select <strong>Duplicate</strong></li>
</ol>

<h2>Editing Records</h2>

<h3>Inline Editing</h3>

<p>Edit directly in the Grid view:</p>

<ol>
	<li>Click any cell to select it</li>
	<li>Start typing to replace the value, or</li>
	<li>Press <kbd>Enter</kbd> to edit the existing value</li>
	<li>Press <kbd>Enter</kbd> again or click away to save</li>
	<li>Press <kbd>Escape</kbd> to cancel changes</li>
</ol>

<h3>Expanded Record View</h3>

<p>For a better editing experience, especially with long text or many fields:</p>

<ol>
	<li>Click the expand icon on the left of any row</li>
	<li>Edit fields in the expanded modal view</li>
	<li>Changes save automatically</li>
</ol>

<p>The expanded view also shows:</p>
<ul>
	<li>Full field names and values</li>
	<li>Comments and activity history</li>
	<li>Linked record details</li>
</ul>

<h3>Field-Specific Editing</h3>

<p>Different field types have different editing experiences:</p>

<table>
	<thead>
		<tr>
			<th>Field Type</th>
			<th>How to Edit</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Text</td>
			<td>Type directly</td>
		</tr>
		<tr>
			<td>Number</td>
			<td>Type numbers (decimal point supported)</td>
		</tr>
		<tr>
			<td>Checkbox</td>
			<td>Click to toggle</td>
		</tr>
		<tr>
			<td>Date</td>
			<td>Click to open date picker</td>
		</tr>
		<tr>
			<td>Single Select</td>
			<td>Click to open dropdown</td>
		</tr>
		<tr>
			<td>Multi Select</td>
			<td>Click to open multi-select picker</td>
		</tr>
		<tr>
			<td>Linked Record</td>
			<td>Click to search and select records</td>
		</tr>
	</tbody>
</table>

<h2>Selecting Records</h2>

<h3>Single Selection</h3>

<p>Click the row number on the left to select a single record.</p>

<h3>Multiple Selection</h3>

<p>Select multiple records by:</p>

<ul>
	<li><kbd>Shift</kbd> + Click to select a range</li>
	<li><kbd>Cmd/Ctrl</kbd> + Click to add individual records to selection</li>
</ul>

<h3>Select All</h3>

<p>Click the checkbox in the header row to select all visible records.</p>

<h2>Deleting Records</h2>

<h3>Delete Single Record</h3>

<ol>
	<li>Select the record (click row number)</li>
	<li>Press <kbd>Delete</kbd> or <kbd>Backspace</kbd>, or</li>
	<li>Right-click and select <strong>Delete record</strong></li>
</ol>

<h3>Delete Multiple Records</h3>

<ol>
	<li>Select multiple records</li>
	<li>Press <kbd>Delete</kbd> or <kbd>Backspace</kbd></li>
	<li>Confirm the deletion</li>
</ol>

<div class="warning">
	<strong>Warning:</strong> Deleted records are permanently removed. Consider exporting important data before deletion.
</div>

<h2>Sorting Records</h2>

<p>Sort your records to organize data by any field:</p>

<ol>
	<li>Click a column header</li>
	<li>Select <strong>Sort ascending</strong> or <strong>Sort descending</strong></li>
</ol>

<p>Or use the sort button in the toolbar to:</p>
<ul>
	<li>Sort by multiple fields</li>
	<li>Change sort order</li>
	<li>Remove sorts</li>
</ul>

<div class="note">
	<strong>Note:</strong> Sorts are saved per view. Different views can have different sort configurations.
</div>

<h2>Filtering Records</h2>

<p>Filter to show only records matching specific criteria:</p>

<ol>
	<li>Click the <strong>Filter</strong> button in the toolbar</li>
	<li>Click <strong>Add filter</strong></li>
	<li>Select a field, condition, and value</li>
	<li>Add more filters as needed</li>
</ol>

<h3>Filter Conditions</h3>

<table>
	<thead>
		<tr>
			<th>Condition</th>
			<th>Description</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Is</td>
			<td>Exact match</td>
		</tr>
		<tr>
			<td>Is not</td>
			<td>Does not match</td>
		</tr>
		<tr>
			<td>Contains</td>
			<td>Text includes value</td>
		</tr>
		<tr>
			<td>Does not contain</td>
			<td>Text excludes value</td>
		</tr>
		<tr>
			<td>Is empty</td>
			<td>Field has no value</td>
		</tr>
		<tr>
			<td>Is not empty</td>
			<td>Field has a value</td>
		</tr>
		<tr>
			<td>Greater than</td>
			<td>Number/date comparison</td>
		</tr>
		<tr>
			<td>Less than</td>
			<td>Number/date comparison</td>
		</tr>
	</tbody>
</table>

<div class="tip">
	<strong>Tip:</strong> Filters are saved per view. Create different views with pre-configured filters for quick access to different record subsets.
</div>

<h2>Searching Records</h2>

<p>Use the search feature to quickly find records:</p>

<ol>
	<li>Press <kbd>Cmd/Ctrl</kbd> + <kbd>F</kbd> or click the search icon</li>
	<li>Type your search term</li>
	<li>Results are highlighted in real-time</li>
</ol>

<p>Search looks across all visible fields in the current view.</p>

<h2>Copy and Paste</h2>

<h3>Copy Cell Value</h3>

<ol>
	<li>Select a cell</li>
	<li>Press <kbd>Cmd/Ctrl</kbd> + <kbd>C</kbd></li>
</ol>

<h3>Paste Value</h3>

<ol>
	<li>Select target cell</li>
	<li>Press <kbd>Cmd/Ctrl</kbd> + <kbd>V</kbd></li>
</ol>

<h3>Paste Multiple Values</h3>

<p>Paste tabular data from spreadsheets:</p>

<ol>
	<li>Copy cells from Excel, Google Sheets, etc.</li>
	<li>Select the starting cell in VibeTable</li>
	<li>Press <kbd>Cmd/Ctrl</kbd> + <kbd>V</kbd></li>
	<li>Data fills across rows and columns</li>
</ol>

<h2>Undo and Redo</h2>

<p>Quickly reverse changes:</p>

<ul>
	<li><kbd>Cmd/Ctrl</kbd> + <kbd>Z</kbd> - Undo last change</li>
	<li><kbd>Cmd/Ctrl</kbd> + <kbd>Shift</kbd> + <kbd>Z</kbd> - Redo</li>
</ul>

<p>Undo works for:</p>
<ul>
	<li>Cell edits</li>
	<li>Record creation</li>
	<li>Record deletion</li>
</ul>

<h2>Related Topics</h2>

<ul>
	<li><a href="/docs/fields">Fields & Field Types</a> - Understanding your data structure</li>
	<li><a href="/docs/views/grid">Grid View</a> - The primary view for data editing</li>
	<li><a href="/docs/shortcuts">Keyboard Shortcuts</a> - Speed up your workflow</li>
</ul>
