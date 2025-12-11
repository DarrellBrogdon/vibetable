<svelte:head>
	<title>Calendar View - VibeTable Documentation</title>
</svelte:head>

<h1>Calendar View</h1>

<p>The Calendar view displays your records on a calendar based on date fields, making it easy to visualize schedules, deadlines, and time-based data.</p>

<h2>Overview</h2>

<p>Calendar view shows records as events on a monthly, weekly, or daily calendar:</p>

<ul>
	<li><strong>Events</strong> - Each record with a date appears as an event</li>
	<li><strong>Date range</strong> - Records with start and end dates span multiple days</li>
	<li><strong>Drag and drop</strong> - Move events to reschedule</li>
</ul>

<p>This view is ideal for:</p>
<ul>
	<li>Project timelines and milestones</li>
	<li>Event planning</li>
	<li>Content scheduling</li>
	<li>Deadline tracking</li>
	<li>Resource scheduling</li>
</ul>

<h2>Setting Up Calendar</h2>

<h3>Requirements</h3>

<p>To use Calendar view, your table needs:</p>

<ul>
	<li>At least one <strong>Date</strong> field</li>
</ul>

<h3>Creating a Calendar View</h3>

<ol>
	<li>Click <strong>+ Add view</strong> in the view tabs</li>
	<li>Select <strong>Calendar</strong></li>
	<li>Name your view (e.g., "Project Timeline")</li>
	<li>Choose the date field to use</li>
	<li>Click <strong>Create</strong></li>
</ol>

<h3>Date Field Options</h3>

<p>You can configure:</p>

<ul>
	<li><strong>Start date field</strong> - When events begin</li>
	<li><strong>End date field</strong> (optional) - When events end (for multi-day events)</li>
</ul>

<p>If you only specify a start date, events appear on a single day.</p>

<h2>Calendar Navigation</h2>

<h3>Changing Views</h3>

<p>Switch between different calendar perspectives:</p>

<ul>
	<li><strong>Month</strong> - See the entire month at once</li>
	<li><strong>Week</strong> - Focus on a single week</li>
	<li><strong>Day</strong> - Detailed view of a single day</li>
</ul>

<p>Click the view buttons in the calendar toolbar to switch.</p>

<h3>Navigation</h3>

<ul>
	<li><strong>Previous/Next</strong> - Use arrow buttons to move between periods</li>
	<li><strong>Today</strong> - Click "Today" to jump to the current date</li>
	<li><strong>Click a date</strong> - In month view, click a date to zoom into that day</li>
</ul>

<h2>Working with Events</h2>

<h3>Viewing Events</h3>

<p>Click any event to open the expanded record view with full details.</p>

<h3>Creating Events</h3>

<p>Add new events directly on the calendar:</p>

<ol>
	<li>Click on a date (or click and drag for a date range)</li>
	<li>A new record form appears</li>
	<li>Fill in the details</li>
	<li>Click <strong>Create</strong></li>
</ol>

<p>The date field is automatically set based on where you clicked.</p>

<h3>Moving Events</h3>

<p>Reschedule by dragging:</p>

<ol>
	<li>Click and hold an event</li>
	<li>Drag to the new date</li>
	<li>Release to update</li>
</ol>

<p>This automatically updates the date field in the record.</p>

<h3>Resizing Events</h3>

<p>For multi-day events (when end date is configured):</p>

<ol>
	<li>Hover over the edge of an event</li>
	<li>Drag to extend or shorten the duration</li>
</ol>

<h2>Calendar Configuration</h2>

<h3>Event Display</h3>

<p>Configure how events appear:</p>

<ol>
	<li>Click the settings icon</li>
	<li>Choose which fields to show on events</li>
</ol>

<p>Events typically show:</p>
<ul>
	<li>Title (primary field)</li>
	<li>Time (if the date field includes time)</li>
	<li>Optional color coding</li>
</ul>

<h3>Color Coding</h3>

<p>Color events by a Single Select field:</p>

<ol>
	<li>Click settings</li>
	<li>Select <strong>Color events by</strong></li>
	<li>Choose a Single Select field (e.g., Status, Category)</li>
</ol>

<p>Events will be colored based on their select field value.</p>

<h3>Filtering Events</h3>

<p>Show only certain events:</p>

<ol>
	<li>Click <strong>Filter</strong> in the toolbar</li>
	<li>Add filter conditions</li>
	<li>Only matching events appear</li>
</ol>

<p>Example filters:</p>
<ul>
	<li>Assigned to me</li>
	<li>Status is not "Completed"</li>
	<li>Priority is "High"</li>
</ul>

<h2>Tips and Best Practices</h2>

<div class="tip">
	<strong>Use start and end dates:</strong> For projects and tasks with duration, use separate start and end date fields to see time spans on the calendar.
</div>

<div class="tip">
	<strong>Include time:</strong> If your events have specific times, make sure your Date field has "Include time" enabled for accurate scheduling.
</div>

<div class="tip">
	<strong>Color by priority:</strong> Color-code by priority or status to quickly identify important items on busy days.
</div>

<div class="tip">
	<strong>Create filtered views:</strong> Make separate calendar views for different purposes - "My Schedule" filtered to your assignments, "Team Calendar" showing everyone.
</div>

<h2>Common Use Cases</h2>

<h3>Project Management</h3>

<table>
	<tbody>
		<tr>
			<td><strong>Start date field</strong></td>
			<td>Start Date</td>
		</tr>
		<tr>
			<td><strong>End date field</strong></td>
			<td>Due Date</td>
		</tr>
		<tr>
			<td><strong>Color by</strong></td>
			<td>Priority or Status</td>
		</tr>
	</tbody>
</table>

<h3>Event Planning</h3>

<table>
	<tbody>
		<tr>
			<td><strong>Start date field</strong></td>
			<td>Event Date & Time</td>
		</tr>
		<tr>
			<td><strong>Color by</strong></td>
			<td>Event Type</td>
		</tr>
	</tbody>
</table>

<h3>Content Calendar</h3>

<table>
	<tbody>
		<tr>
			<td><strong>Start date field</strong></td>
			<td>Publish Date</td>
		</tr>
		<tr>
			<td><strong>Color by</strong></td>
			<td>Content Type or Status</td>
		</tr>
	</tbody>
</table>

<h2>Related Topics</h2>

<ul>
	<li><a href="/docs/fields">Fields & Field Types</a> - Setting up Date fields</li>
	<li><a href="/docs/views/kanban">Kanban View</a> - For status-based tracking</li>
	<li><a href="/docs/views/gallery">Gallery View</a> - For visual content</li>
</ul>
