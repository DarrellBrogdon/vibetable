<script lang="ts">
	const isMac = typeof navigator !== 'undefined' && navigator.platform.toUpperCase().indexOf('MAC') >= 0;
	const cmdKey = isMac ? '⌘' : 'Ctrl';
</script>

<svelte:head>
	<title>Keyboard Shortcuts - VibeTable Documentation</title>
</svelte:head>

<h1>Keyboard Shortcuts</h1>

<p>Master these keyboard shortcuts to navigate and edit faster. Press <kbd>{cmdKey}</kbd> + <kbd>/</kbd> anywhere in the app to see a quick reference.</p>

<h2>Navigation</h2>

<table>
	<thead>
		<tr>
			<th>Shortcut</th>
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
			<td>Enter edit mode / Move down when editing</td>
		</tr>
		<tr>
			<td><kbd>Escape</kbd></td>
			<td>Exit edit mode / Deselect / Close modal</td>
		</tr>
		<tr>
			<td><kbd>Home</kbd></td>
			<td>Go to first cell in row</td>
		</tr>
		<tr>
			<td><kbd>End</kbd></td>
			<td>Go to last cell in row</td>
		</tr>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>Home</kbd></td>
			<td>Go to first cell in table</td>
		</tr>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>End</kbd></td>
			<td>Go to last cell in table</td>
		</tr>
		<tr>
			<td><kbd>Page Up</kbd></td>
			<td>Scroll up one page</td>
		</tr>
		<tr>
			<td><kbd>Page Down</kbd></td>
			<td>Scroll down one page</td>
		</tr>
	</tbody>
</table>

<h2>Selection</h2>

<table>
	<thead>
		<tr>
			<th>Shortcut</th>
			<th>Action</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td><kbd>Shift</kbd> + <kbd>↑</kbd> <kbd>↓</kbd> <kbd>←</kbd> <kbd>→</kbd></td>
			<td>Extend selection</td>
		</tr>
		<tr>
			<td><kbd>Shift</kbd> + Click</td>
			<td>Select range of cells</td>
		</tr>
		<tr>
			<td><kbd>{cmdKey}</kbd> + Click</td>
			<td>Add to selection</td>
		</tr>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>A</kbd></td>
			<td>Select all records</td>
		</tr>
		<tr>
			<td><kbd>Shift</kbd> + <kbd>Space</kbd></td>
			<td>Select entire row</td>
		</tr>
	</tbody>
</table>

<h2>Editing</h2>

<table>
	<thead>
		<tr>
			<th>Shortcut</th>
			<th>Action</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td><kbd>Enter</kbd> or <kbd>F2</kbd></td>
			<td>Start editing selected cell</td>
		</tr>
		<tr>
			<td>Start typing</td>
			<td>Replace cell content</td>
		</tr>
		<tr>
			<td><kbd>Delete</kbd> or <kbd>Backspace</kbd></td>
			<td>Clear cell content</td>
		</tr>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>C</kbd></td>
			<td>Copy selected cells</td>
		</tr>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>V</kbd></td>
			<td>Paste</td>
		</tr>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>X</kbd></td>
			<td>Cut selected cells</td>
		</tr>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>Z</kbd></td>
			<td>Undo</td>
		</tr>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>Shift</kbd> + <kbd>Z</kbd></td>
			<td>Redo</td>
		</tr>
		<tr>
			<td><kbd>Space</kbd></td>
			<td>Toggle checkbox (when on checkbox cell)</td>
		</tr>
	</tbody>
</table>

<h2>Records</h2>

<table>
	<thead>
		<tr>
			<th>Shortcut</th>
			<th>Action</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>Enter</kbd></td>
			<td>Create new record</td>
		</tr>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>D</kbd></td>
			<td>Duplicate selected record(s)</td>
		</tr>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>Backspace</kbd></td>
			<td>Delete selected record(s)</td>
		</tr>
		<tr>
			<td><kbd>Space</kbd></td>
			<td>Open expanded record view (when row selected)</td>
		</tr>
	</tbody>
</table>

<h2>Search & Filter</h2>

<table>
	<thead>
		<tr>
			<th>Shortcut</th>
			<th>Action</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>F</kbd></td>
			<td>Open search</td>
		</tr>
		<tr>
			<td><kbd>Enter</kbd></td>
			<td>Find next match (in search)</td>
		</tr>
		<tr>
			<td><kbd>Shift</kbd> + <kbd>Enter</kbd></td>
			<td>Find previous match (in search)</td>
		</tr>
		<tr>
			<td><kbd>Escape</kbd></td>
			<td>Close search</td>
		</tr>
	</tbody>
</table>

<h2>Views & Navigation</h2>

<table>
	<thead>
		<tr>
			<th>Shortcut</th>
			<th>Action</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>1</kbd> - <kbd>9</kbd></td>
			<td>Switch to view 1-9</td>
		</tr>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>[</kbd></td>
			<td>Previous view</td>
		</tr>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>]</kbd></td>
			<td>Next view</td>
		</tr>
	</tbody>
</table>

<h2>Help</h2>

<table>
	<thead>
		<tr>
			<th>Shortcut</th>
			<th>Action</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td><kbd>{cmdKey}</kbd> + <kbd>/</kbd></td>
			<td>Show keyboard shortcuts</td>
		</tr>
		<tr>
			<td><kbd>?</kbd></td>
			<td>Show keyboard shortcuts (when not editing)</td>
		</tr>
	</tbody>
</table>

<h2>Tips for Efficient Editing</h2>

<h3>Quick Data Entry</h3>

<p>The fastest way to enter data:</p>

<ol>
	<li>Click the first cell</li>
	<li>Type the value</li>
	<li>Press <kbd>Tab</kbd> to move right</li>
	<li>Press <kbd>Enter</kbd> at end of row to move to next row</li>
</ol>

<h3>Batch Operations</h3>

<p>Edit multiple cells at once:</p>

<ol>
	<li>Select a range of cells</li>
	<li>Type a value</li>
	<li>Press <kbd>Enter</kbd> to apply to all selected cells</li>
</ol>

<h3>Quick Toggle</h3>

<p>For checkbox fields, just press <kbd>Space</kbd> to toggle without entering edit mode.</p>

<h3>Duplicate Workflow</h3>

<p>When creating similar records:</p>

<ol>
	<li>Create the first record with common values</li>
	<li>Select it and press <kbd>{cmdKey}</kbd> + <kbd>D</kbd></li>
	<li>Edit just the fields that differ</li>
</ol>

<h2>Modal Shortcuts</h2>

<p>When a modal is open:</p>

<table>
	<thead>
		<tr>
			<th>Shortcut</th>
			<th>Action</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td><kbd>Escape</kbd></td>
			<td>Close modal</td>
		</tr>
		<tr>
			<td><kbd>Enter</kbd></td>
			<td>Confirm / Submit (in most modals)</td>
		</tr>
		<tr>
			<td><kbd>Tab</kbd></td>
			<td>Move to next field</td>
		</tr>
	</tbody>
</table>

<h2>Customizing Shortcuts</h2>

<p>VibeTable uses standard shortcuts that work across most applications. This makes it easier to switch between VibeTable and other tools like spreadsheets.</p>

<div class="note">
	<strong>Note:</strong> Custom keyboard shortcuts are not currently supported. We use industry-standard shortcuts for familiarity.
</div>

<h2>Related Topics</h2>

<ul>
	<li><a href="/docs/views/grid">Grid View</a> - Where most shortcuts apply</li>
	<li><a href="/docs/records">Records & Data</a> - Working with your data</li>
</ul>
