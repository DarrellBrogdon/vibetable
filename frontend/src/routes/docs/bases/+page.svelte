<svelte:head>
	<title>Bases & Tables - VibeTable Documentation</title>
</svelte:head>

<h1>Bases & Tables</h1>

<p>VibeTable organizes your data into <strong>bases</strong> and <strong>tables</strong>. Understanding this structure is key to using VibeTable effectively.</p>

<h2>What is a Base?</h2>

<p>A <strong>base</strong> is a container that holds related tables. Think of it like:</p>

<ul>
	<li>A workbook in Excel (containing multiple sheets)</li>
	<li>A database (containing multiple tables)</li>
	<li>A project folder (containing related documents)</li>
</ul>

<p>Each base has its own:</p>
<ul>
	<li><strong>Tables</strong> - The actual data structures</li>
	<li><strong>Collaborators</strong> - People who can access the base</li>
	<li><strong>Webhooks</strong> - Automated notifications to external services</li>
</ul>

<h3>Creating a Base</h3>

<ol>
	<li>From the home screen, click <strong>"+ New Base"</strong></li>
	<li>Enter a name for your base</li>
	<li>Click <strong>Create</strong></li>
</ol>

<p>Your base will be created with a default table called "Table 1".</p>

<h3>Managing Bases</h3>

<p>From the base detail page, you can:</p>

<ul>
	<li><strong>Rename</strong> - Click the base name to edit it</li>
	<li><strong>Delete</strong> - Use the menu (only available to owners)</li>
	<li><strong>Share</strong> - Invite collaborators via the Share button</li>
</ul>

<h2>What is a Table?</h2>

<p>A <strong>table</strong> stores your actual data in rows and columns. Each table has:</p>

<ul>
	<li><strong>Fields</strong> (columns) - Define the structure and types of data</li>
	<li><strong>Records</strong> (rows) - Individual data entries</li>
	<li><strong>Views</strong> - Different ways to visualize the same data</li>
</ul>

<h3>Creating Tables</h3>

<p>Within a base, you can create multiple tables:</p>

<ol>
	<li>In the left sidebar, click <strong>"+ Add Table"</strong></li>
	<li>Enter a name for your table</li>
	<li>Press Enter or click Create</li>
</ol>

<h3>When to Use Multiple Tables</h3>

<p>Create separate tables for different types of things. For example, a "Project Management" base might have:</p>

<table>
	<thead>
		<tr>
			<th>Table</th>
			<th>Purpose</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Projects</td>
			<td>Track individual projects</td>
		</tr>
		<tr>
			<td>Tasks</td>
			<td>Store tasks linked to projects</td>
		</tr>
		<tr>
			<td>Team Members</td>
			<td>List of people who can be assigned</td>
		</tr>
		<tr>
			<td>Milestones</td>
			<td>Key dates and deliverables</td>
		</tr>
	</tbody>
</table>

<p>Tables in the same base can be linked together using <strong>Linked Record</strong> fields, allowing you to create relationships between data.</p>

<h3>Managing Tables</h3>

<p>To manage a table:</p>

<ul>
	<li><strong>Rename</strong> - Right-click the table name in the sidebar</li>
	<li><strong>Delete</strong> - Right-click and select Delete (cannot be undone)</li>
	<li><strong>Reorder</strong> - Drag tables in the sidebar to rearrange</li>
</ul>

<div class="warning">
	<strong>Warning:</strong> Deleting a table permanently removes all its data, including records, fields, and views. This cannot be undone.
</div>

<h2>Organization Best Practices</h2>

<h3>When to Create a New Base</h3>

<p>Create a new base when:</p>

<ul>
	<li>The data is for a completely different project or purpose</li>
	<li>You need different people to have access</li>
	<li>The tables won't need to link to each other</li>
</ul>

<h3>When to Add Tables to an Existing Base</h3>

<p>Add tables to the same base when:</p>

<ul>
	<li>The data is related and needs to be linked</li>
	<li>The same team will work with all the tables</li>
	<li>You want unified reporting across tables</li>
</ul>

<h3>Naming Conventions</h3>

<p>Good naming makes your bases easier to navigate:</p>

<ul>
	<li><strong>Bases:</strong> Use project or purpose names (e.g., "Q4 Marketing Campaign")</li>
	<li><strong>Tables:</strong> Use plural nouns for what they contain (e.g., "Tasks", "Customers", "Products")</li>
	<li><strong>Fields:</strong> Use clear, descriptive names (e.g., "Due Date" instead of "Date")</li>
</ul>

<h2>Related Topics</h2>

<ul>
	<li><a href="/docs/fields">Fields & Field Types</a> - Learn about the different types of columns</li>
	<li><a href="/docs/records">Records & Data</a> - Working with your actual data</li>
	<li><a href="/docs/sharing">Sharing & Permissions</a> - Collaborating with others</li>
</ul>
