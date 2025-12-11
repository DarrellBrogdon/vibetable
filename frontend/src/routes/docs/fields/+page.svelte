<svelte:head>
	<title>Fields & Field Types - VibeTable Documentation</title>
</svelte:head>

<h1>Fields & Field Types</h1>

<p>Fields define the columns in your tables and determine what kind of data can be stored. VibeTable supports a variety of field types to handle different data needs.</p>

<h2>Understanding Fields</h2>

<p>Each field has:</p>
<ul>
	<li><strong>Name</strong> - The column header displayed in the table</li>
	<li><strong>Type</strong> - Determines what kind of data can be stored</li>
	<li><strong>Options</strong> - Type-specific settings (e.g., date format, select choices)</li>
</ul>

<h2>Field Types</h2>

<h3>Text</h3>

<p>A single-line text field for short text entries.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Names, titles, short descriptions, IDs</td>
		</tr>
		<tr>
			<td><strong>Example values</strong></td>
			<td>"John Smith", "Project Alpha", "SKU-12345"</td>
		</tr>
	</tbody>
</table>

<h3>Long Text</h3>

<p>A multi-line text field for longer content.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Descriptions, notes, comments, paragraphs</td>
		</tr>
		<tr>
			<td><strong>Features</strong></td>
			<td>Supports multiple lines, expands when editing</td>
		</tr>
	</tbody>
</table>

<h3>Number</h3>

<p>Stores numeric values with optional formatting.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Quantities, prices, scores, measurements</td>
		</tr>
		<tr>
			<td><strong>Options</strong></td>
			<td>Decimal precision, format (number, currency, percent)</td>
		</tr>
		<tr>
			<td><strong>Example values</strong></td>
			<td>42, 3.14, $99.99, 75%</td>
		</tr>
	</tbody>
</table>

<h3>Checkbox</h3>

<p>A true/false toggle for binary data.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Completion status, yes/no questions, feature flags</td>
		</tr>
		<tr>
			<td><strong>Interaction</strong></td>
			<td>Click to toggle on/off</td>
		</tr>
	</tbody>
</table>

<h3>Date</h3>

<p>Stores date values with optional time.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Due dates, birthdays, events, deadlines</td>
		</tr>
		<tr>
			<td><strong>Options</strong></td>
			<td>Include time, date format</td>
		</tr>
		<tr>
			<td><strong>Features</strong></td>
			<td>Date picker, calendar integration</td>
		</tr>
	</tbody>
</table>

<div class="tip">
	<strong>Tip:</strong> Fields with dates can power the Calendar view. Create a Date field to unlock calendar visualization for your table.
</div>

<h3>Single Select</h3>

<p>Choose one option from a predefined list.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Status, category, priority, type</td>
		</tr>
		<tr>
			<td><strong>Options</strong></td>
			<td>Define choices with custom colors</td>
		</tr>
		<tr>
			<td><strong>Example uses</strong></td>
			<td>Status (To Do, In Progress, Done), Priority (Low, Medium, High)</td>
		</tr>
	</tbody>
</table>

<div class="tip">
	<strong>Tip:</strong> Single Select fields can power Kanban views. Group your cards by any Single Select field.
</div>

<h3>Multi Select</h3>

<p>Choose multiple options from a predefined list.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Tags, labels, categories, skills</td>
		</tr>
		<tr>
			<td><strong>Options</strong></td>
			<td>Define choices with custom colors</td>
		</tr>
		<tr>
			<td><strong>Example uses</strong></td>
			<td>Tags (Frontend, Backend, Design), Features (Wi-Fi, Parking, Pool)</td>
		</tr>
	</tbody>
</table>

<h3>Linked Record</h3>

<p>Create relationships between tables by linking to records in another table.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Relationships, associations, references</td>
		</tr>
		<tr>
			<td><strong>Options</strong></td>
			<td>Target table, allow multiple links</td>
		</tr>
		<tr>
			<td><strong>Example uses</strong></td>
			<td>Tasks → Project, Orders → Customer, Students → Courses</td>
		</tr>
	</tbody>
</table>

<div class="note">
	<strong>Note:</strong> When you create a Linked Record field, a reverse link field is automatically created in the target table.
</div>

<h3>Attachment</h3>

<p>Upload and attach files to records.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Documents, images, files</td>
		</tr>
		<tr>
			<td><strong>Features</strong></td>
			<td>Multiple attachments, preview, download</td>
		</tr>
		<tr>
			<td><strong>Supported types</strong></td>
			<td>Images, PDFs, documents, and more</td>
		</tr>
	</tbody>
</table>

<div class="tip">
	<strong>Tip:</strong> Attachment fields with images can power Gallery views. Set an attachment field as the cover image for visual browsing.
</div>

<h3>URL</h3>

<p>Store web links with clickable functionality.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Websites, resources, references</td>
		</tr>
		<tr>
			<td><strong>Features</strong></td>
			<td>Clickable links, opens in new tab</td>
		</tr>
	</tbody>
</table>

<h3>Email</h3>

<p>Store email addresses with mailto links.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Contact information, user emails</td>
		</tr>
		<tr>
			<td><strong>Features</strong></td>
			<td>Clickable mailto links, validation</td>
		</tr>
	</tbody>
</table>

<h2>Computed Fields</h2>

<p>Some fields are calculated automatically based on other data:</p>

<h3>Formula</h3>

<p>Calculate values based on other fields in the same record.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Calculations, concatenations, conditionals</td>
		</tr>
		<tr>
			<td><strong>Examples</strong></td>
			<td>Total = Price × Quantity, Full Name = First + " " + Last</td>
		</tr>
	</tbody>
</table>

<h3>Rollup</h3>

<p>Aggregate values from linked records.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Summaries, counts, averages across relationships</td>
		</tr>
		<tr>
			<td><strong>Examples</strong></td>
			<td>Total Tasks per Project, Average Score per Student</td>
		</tr>
	</tbody>
</table>

<h3>Count</h3>

<p>Count the number of linked records.</p>

<table>
	<tbody>
		<tr>
			<td><strong>Best for</strong></td>
			<td>Counting relationships</td>
		</tr>
		<tr>
			<td><strong>Examples</strong></td>
			<td>Number of Tasks, Number of Attendees</td>
		</tr>
	</tbody>
</table>

<h2>Working with Fields</h2>

<h3>Adding a Field</h3>

<ol>
	<li>Click the <strong>+</strong> button at the end of the header row</li>
	<li>Select a field type from the dropdown</li>
	<li>Enter a name for the field</li>
	<li>Configure any type-specific options</li>
	<li>Click <strong>Create</strong></li>
</ol>

<h3>Editing a Field</h3>

<ol>
	<li>Click the field header to open the field menu</li>
	<li>Select <strong>Edit field</strong></li>
	<li>Modify the name or options</li>
	<li>Click <strong>Save</strong></li>
</ol>

<div class="warning">
	<strong>Warning:</strong> Changing a field's type may result in data loss if existing values aren't compatible with the new type.
</div>

<h3>Deleting a Field</h3>

<ol>
	<li>Click the field header to open the field menu</li>
	<li>Select <strong>Delete field</strong></li>
	<li>Confirm the deletion</li>
</ol>

<div class="warning">
	<strong>Warning:</strong> Deleting a field permanently removes all data in that column. This cannot be undone.
</div>

<h3>Reordering Fields</h3>

<p>Drag field headers left or right to reorder columns in your table.</p>

<h2>Related Topics</h2>

<ul>
	<li><a href="/docs/records">Records & Data</a> - Working with the data in your fields</li>
	<li><a href="/docs/views/grid">Grid View</a> - Spreadsheet-like editing</li>
	<li><a href="/docs/views/kanban">Kanban View</a> - Using Single Select for card grouping</li>
</ul>
