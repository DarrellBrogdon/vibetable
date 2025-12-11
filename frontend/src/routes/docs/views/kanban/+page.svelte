<svelte:head>
	<title>Kanban View - VibeTable Documentation</title>
</svelte:head>

<h1>Kanban View</h1>

<p>The Kanban view displays your records as cards organized into columns, perfect for managing workflows, tracking progress, and visualizing status-based data.</p>

<h2>Overview</h2>

<p>Kanban boards organize records into columns based on a Single Select field:</p>

<ul>
	<li><strong>Columns</strong> - Each option in the select field becomes a column</li>
	<li><strong>Cards</strong> - Each record appears as a card</li>
	<li><strong>Drag and drop</strong> - Move cards between columns to update status</li>
</ul>

<p>This view is ideal for:</p>
<ul>
	<li>Project management (To Do, In Progress, Done)</li>
	<li>Sales pipelines (Lead, Qualified, Proposal, Won)</li>
	<li>Content workflows (Draft, Review, Published)</li>
	<li>Any status-based tracking</li>
</ul>

<h2>Setting Up Kanban</h2>

<h3>Requirements</h3>

<p>To use Kanban view, your table needs:</p>

<ul>
	<li>At least one <strong>Single Select</strong> field (for grouping columns)</li>
</ul>

<h3>Creating a Kanban View</h3>

<ol>
	<li>Click <strong>+ Add view</strong> in the view tabs</li>
	<li>Select <strong>Kanban</strong></li>
	<li>Name your view (e.g., "Task Board")</li>
	<li>Choose the Single Select field to group by</li>
	<li>Click <strong>Create</strong></li>
</ol>

<h3>Choosing a Grouping Field</h3>

<p>The grouping field determines your columns. Good candidates include:</p>

<table>
	<thead>
		<tr>
			<th>Use Case</th>
			<th>Field Options</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Task tracking</td>
			<td>To Do, In Progress, Review, Done</td>
		</tr>
		<tr>
			<td>Sales pipeline</td>
			<td>Lead, Contacted, Qualified, Proposal, Won, Lost</td>
		</tr>
		<tr>
			<td>Bug tracking</td>
			<td>New, Triaged, In Progress, Fixed, Closed</td>
		</tr>
		<tr>
			<td>Content calendar</td>
			<td>Idea, Writing, Editing, Scheduled, Published</td>
		</tr>
	</tbody>
</table>

<h2>Working with Cards</h2>

<h3>Card Display</h3>

<p>Each card shows:</p>
<ul>
	<li><strong>Title</strong> - The primary field value</li>
	<li><strong>Preview fields</strong> - Additional fields you choose to display</li>
	<li><strong>Cover image</strong> - Optional image from an attachment field</li>
</ul>

<h3>Moving Cards</h3>

<p>Drag and drop to move cards:</p>

<ol>
	<li>Click and hold a card</li>
	<li>Drag to the target column</li>
	<li>Drop to update the status</li>
</ol>

<p>Moving a card automatically updates the underlying record's status field.</p>

<h3>Opening Cards</h3>

<p>Click any card to open the expanded record view where you can:</p>
<ul>
	<li>View and edit all fields</li>
	<li>Add comments</li>
	<li>See activity history</li>
</ul>

<h3>Creating Cards</h3>

<p>Add new cards directly in Kanban:</p>

<ol>
	<li>Click <strong>+ Add card</strong> at the bottom of any column</li>
	<li>Enter the card title</li>
	<li>Press <kbd>Enter</kbd> or click away to create</li>
</ol>

<p>The new record will automatically have the column's status value.</p>

<h2>Kanban Configuration</h2>

<h3>Card Preview Fields</h3>

<p>Choose which fields appear on cards:</p>

<ol>
	<li>Click the settings icon in the view</li>
	<li>Select <strong>Card preview</strong></li>
	<li>Toggle fields on/off</li>
</ol>

<p>Recommended preview fields:</p>
<ul>
	<li>Assignee or owner</li>
	<li>Due date</li>
	<li>Priority</li>
	<li>Tags or categories</li>
</ul>

<h3>Cover Images</h3>

<p>Add visual previews to cards:</p>

<ol>
	<li>Click the settings icon</li>
	<li>Select <strong>Cover image</strong></li>
	<li>Choose an Attachment field</li>
</ol>

<p>The first image in each record's attachment field will display as the card cover.</p>

<h3>Column Order</h3>

<p>Columns appear in the same order as options in your Single Select field. To reorder:</p>

<ol>
	<li>Edit the Single Select field</li>
	<li>Drag options to reorder them</li>
	<li>Changes reflect in all Kanban views using that field</li>
</ol>

<h3>Hiding Columns</h3>

<p>Hide specific columns from view:</p>

<ol>
	<li>Click the column menu (three dots)</li>
	<li>Select <strong>Hide column</strong></li>
</ol>

<p>Hidden columns' records still exist - they're just not displayed.</p>

<h2>Filtering and Sorting</h2>

<h3>Filters</h3>

<p>Apply filters to show only certain cards:</p>

<ol>
	<li>Click <strong>Filter</strong> in the toolbar</li>
	<li>Add filter conditions</li>
	<li>Only matching cards appear</li>
</ol>

<p>Example: Filter by assignee to see only your tasks.</p>

<h3>Sorting Cards</h3>

<p>Control card order within columns:</p>

<ol>
	<li>Click <strong>Sort</strong> in the toolbar</li>
	<li>Choose a field and direction</li>
</ol>

<p>Common sorts:</p>
<ul>
	<li>Priority (High to Low)</li>
	<li>Due Date (Earliest first)</li>
	<li>Created Date (Newest first)</li>
</ul>

<h2>Tips and Best Practices</h2>

<div class="tip">
	<strong>Limit columns:</strong> Keep your board manageable with 4-7 columns. Too many columns make the board hard to use.
</div>

<div class="tip">
	<strong>Use WIP limits:</strong> Mentally set limits for "In Progress" columns. Too many items in progress usually means nothing gets finished.
</div>

<div class="tip">
	<strong>Color-code by priority:</strong> Use the Single Select option colors to indicate priority at a glance.
</div>

<div class="tip">
	<strong>Create focused views:</strong> Make separate Kanban views for different purposes - "My Tasks" filtered to you, "Team Board" showing everyone.
</div>

<h2>Related Topics</h2>

<ul>
	<li><a href="/docs/fields">Fields & Field Types</a> - Creating Single Select fields</li>
	<li><a href="/docs/views/grid">Grid View</a> - For bulk editing</li>
	<li><a href="/docs/views/calendar">Calendar View</a> - For date-based planning</li>
</ul>
