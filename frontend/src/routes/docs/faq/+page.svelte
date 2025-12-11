<svelte:head>
	<title>FAQ - VibeTable Documentation</title>
</svelte:head>

<h1>Frequently Asked Questions</h1>

<p>Quick answers to common questions about VibeTable.</p>

<h2>Getting Started</h2>

<h3>What is VibeTable?</h3>

<p>VibeTable is a flexible database tool that combines the simplicity of spreadsheets with the power of a database. You can organize data in tables, create different views, collaborate with others, and collect data through forms.</p>

<h3>Do I need to create an account?</h3>

<p>Yes, you need an account to create and manage bases. However, public view links and forms can be accessed without an account.</p>

<h3>How do I sign in?</h3>

<p>VibeTable uses magic link authentication. Enter your email address, click the link sent to your inbox, and you're signed in. No password required!</p>

<h3>What if my magic link expires?</h3>

<p>Magic links expire after 15 minutes for security. Simply request a new one from the login page.</p>

<h2>Bases & Tables</h2>

<h3>What's the difference between a base and a table?</h3>

<p>A <strong>base</strong> is a container (like a workbook) that holds related <strong>tables</strong> (like sheets). Tables contain your actual data in rows and columns.</p>

<h3>How many tables can I have in a base?</h3>

<p>You can create as many tables as you need in a base. There's no hard limit.</p>

<h3>Can I move a table to a different base?</h3>

<p>Currently, tables cannot be moved between bases. You would need to export the data and import it into a table in the new base.</p>

<h3>Can I import data from Excel or Google Sheets?</h3>

<p>Yes! You can import CSV files into VibeTable. Export your spreadsheet as CSV first, then use the import feature.</p>

<h2>Fields & Data</h2>

<h3>Can I change a field's type after creating it?</h3>

<p>Yes, but be careful - changing field types may result in data loss if existing values aren't compatible with the new type. For example, changing from Text to Number will clear any non-numeric values.</p>

<h3>What happens if I delete a field?</h3>

<p>Deleting a field permanently removes all data in that column across all records. This cannot be undone.</p>

<h3>How do I create a dropdown/select field?</h3>

<p>Create a <strong>Single Select</strong> field for choosing one option, or <strong>Multi Select</strong> for choosing multiple options. When creating the field, you'll define the available options.</p>

<h3>Can I link tables together?</h3>

<p>Yes! Use <strong>Linked Record</strong> fields to create relationships between tables. For example, link Tasks to Projects, or Orders to Customers.</p>

<h2>Views</h2>

<h3>What's the difference between views?</h3>

<table>
	<thead>
		<tr>
			<th>View</th>
			<th>Best For</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td><strong>Grid</strong></td>
			<td>Data entry, bulk editing, spreadsheet-style work</td>
		</tr>
		<tr>
			<td><strong>Kanban</strong></td>
			<td>Visual workflows, status tracking, project management</td>
		</tr>
		<tr>
			<td><strong>Calendar</strong></td>
			<td>Date-based planning, schedules, deadlines</td>
		</tr>
		<tr>
			<td><strong>Gallery</strong></td>
			<td>Visual content, images, product catalogs</td>
		</tr>
	</tbody>
</table>

<h3>Do different views show different data?</h3>

<p>No, all views show the same underlying data. Views are just different ways to visualize and interact with your records. Changes made in one view appear in all views.</p>

<h3>Can I filter a view?</h3>

<p>Yes! Each view can have its own filters, sorts, and hidden fields. This lets you create focused views like "My Tasks" or "Overdue Items" from the same table.</p>

<h3>Can I have multiple views of the same type?</h3>

<p>Yes! You can create as many views as you need. For example, create multiple Grid views with different filters or column arrangements.</p>

<h2>Collaboration</h2>

<h3>How do I share a base with someone?</h3>

<p>Click the <strong>Share</strong> button, enter their email address, choose a role (Owner, Editor, or Viewer), and click Invite. They'll receive an email invitation.</p>

<h3>What's the difference between Editor and Viewer?</h3>

<p><strong>Editors</strong> can add, edit, and delete records and modify views. <strong>Viewers</strong> can only view data - they cannot make changes.</p>

<h3>Can I share just one table from a base?</h3>

<p>Sharing is done at the base level. When you share a base, collaborators have access to all tables in that base. Use public view links if you want to share a specific filtered view.</p>

<h3>What's a public link?</h3>

<p>A public link allows anyone (without an account) to view your data in read-only mode. Useful for sharing with people who don't need to edit.</p>

<h2>Forms</h2>

<h3>Can people submit data without creating an account?</h3>

<p>Yes! Forms are public - anyone with the form link can submit data without signing up.</p>

<h3>Where do form submissions go?</h3>

<p>Form submissions become new records in the table the form is connected to.</p>

<h3>Can I choose which fields appear on a form?</h3>

<p>Yes! In the form builder, you can show/hide fields, change their order, edit labels, and add help text.</p>

<h3>Can I get notified when someone submits a form?</h3>

<p>Yes! Set up a webhook to send notifications to email, Slack, or other services when new records are created.</p>

<h2>Technical</h2>

<h3>Is my data secure?</h3>

<p>Yes. Data is transmitted over HTTPS and stored securely. Access is controlled through your account and the permissions you set.</p>

<h3>Can I export my data?</h3>

<p>Yes, you can export table data to CSV format at any time.</p>

<h3>Does VibeTable work offline?</h3>

<p>VibeTable requires an internet connection to function. Changes are saved to the server immediately.</p>

<h3>Is there an API?</h3>

<p>Yes! VibeTable has a REST API for programmatic access. Contact us for API documentation.</p>

<h2>Troubleshooting</h2>

<h3>Changes aren't saving</h3>

<p>Check your internet connection. VibeTable saves automatically, but requires connectivity. If you see a connection error, wait for it to resolve and try again.</p>

<h3>I can't edit a record</h3>

<p>You may have <strong>Viewer</strong> access instead of <strong>Editor</strong>. Check with the base owner to upgrade your permissions.</p>

<h3>I accidentally deleted something</h3>

<p>Use <kbd>Cmd/Ctrl</kbd> + <kbd>Z</kbd> immediately to undo. Undo works for cell edits and record deletions, but not for field or table deletions.</p>

<h3>The page is loading slowly</h3>

<p>Large tables with thousands of records may take longer to load. Try creating filtered views to work with smaller subsets of data.</p>

<h3>I'm locked out of my account</h3>

<p>Request a new magic link from the login page. If you're not receiving emails, check your spam folder or try a different email address.</p>

<h2>Still Have Questions?</h2>

<p>If you can't find the answer you're looking for:</p>

<ul>
	<li>Browse the documentation topics in the sidebar</li>
	<li>Use the <kbd>Cmd/Ctrl</kbd> + <kbd>/</kbd> shortcut for quick keyboard reference</li>
</ul>
