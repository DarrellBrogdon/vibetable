<svelte:head>
	<title>Forms & Public Access - VibeTable Documentation</title>
</svelte:head>

<h1>Forms & Public Access</h1>

<p>Forms let you collect data from anyone - without giving them access to your base. Perfect for surveys, registrations, feedback collection, and more.</p>

<h2>Overview</h2>

<p>A form is a public page where visitors can submit data that becomes records in your table. Forms:</p>

<ul>
	<li><strong>Require no login</strong> - Anyone with the link can submit</li>
	<li><strong>Control visible fields</strong> - Choose which fields appear</li>
	<li><strong>Customize the experience</strong> - Add descriptions and instructions</li>
	<li><strong>Create records automatically</strong> - Submissions become new records</li>
</ul>

<p>This is ideal for:</p>
<ul>
	<li>Contact forms</li>
	<li>Event registrations</li>
	<li>Feedback surveys</li>
	<li>Job applications</li>
	<li>Order intake</li>
	<li>Lead capture</li>
</ul>

<h2>Creating a Form</h2>

<h3>Basic Setup</h3>

<ol>
	<li>Open the table where you want submissions to go</li>
	<li>Click <strong>Forms</strong> in the sidebar or toolbar</li>
	<li>Click <strong>+ Create form</strong></li>
	<li>Name your form</li>
	<li>Click <strong>Create</strong></li>
</ol>

<h3>Form Builder</h3>

<p>The form builder lets you customize your form:</p>

<ul>
	<li><strong>Add/remove fields</strong> - Toggle which fields appear on the form</li>
	<li><strong>Reorder fields</strong> - Drag to arrange the form layout</li>
	<li><strong>Edit labels</strong> - Customize how field names appear</li>
	<li><strong>Add descriptions</strong> - Help text for each field</li>
	<li><strong>Mark required</strong> - Ensure certain fields are filled</li>
</ul>

<h2>Configuring Fields</h2>

<h3>Showing/Hiding Fields</h3>

<p>Not all table fields need to appear on the form. Toggle fields:</p>

<ol>
	<li>In the form builder, view the field list</li>
	<li>Toggle fields on/off</li>
	<li>Hidden fields won't appear to form visitors</li>
</ol>

<p>Common hidden fields:</p>
<ul>
	<li>Status (auto-set to default)</li>
	<li>Created date (auto-populated)</li>
	<li>Internal notes</li>
	<li>Assigned to</li>
</ul>

<h3>Field Labels</h3>

<p>Customize how field names appear on the form:</p>

<ol>
	<li>Click on a field in the form builder</li>
	<li>Edit the <strong>Label</strong></li>
	<li>This doesn't change the underlying field name</li>
</ol>

<p>Example: Field "contact_email" → Form label "Your Email Address"</p>

<h3>Help Text</h3>

<p>Add descriptions to help users fill out fields:</p>

<ol>
	<li>Click on a field</li>
	<li>Add <strong>Description</strong> text</li>
</ol>

<p>Good help text:</p>
<ul>
	<li>Explains what information is needed</li>
	<li>Provides format examples</li>
	<li>Clarifies any requirements</li>
</ul>

<h3>Required Fields</h3>

<p>Mark fields as required to ensure they're filled:</p>

<ol>
	<li>Click on a field</li>
	<li>Toggle <strong>Required</strong> on</li>
</ol>

<p>Users cannot submit the form until all required fields are completed.</p>

<h2>Form Settings</h2>

<h3>Form Title & Description</h3>

<p>Add context at the top of your form:</p>

<ol>
	<li>Click <strong>Form settings</strong></li>
	<li>Edit the <strong>Title</strong> (defaults to table name)</li>
	<li>Add a <strong>Description</strong> for instructions</li>
</ol>

<h3>Submit Button</h3>

<p>Customize the submit button text:</p>

<ol>
	<li>Click <strong>Form settings</strong></li>
	<li>Edit <strong>Submit button text</strong></li>
</ol>

<p>Examples: "Submit", "Register", "Send Feedback", "Apply Now"</p>

<h3>Confirmation Message</h3>

<p>What users see after submitting:</p>

<ol>
	<li>Click <strong>Form settings</strong></li>
	<li>Edit the <strong>Confirmation message</strong></li>
</ol>

<p>Good confirmation messages:</p>
<ul>
	<li>Thank the user</li>
	<li>Explain what happens next</li>
	<li>Provide contact info if needed</li>
</ul>

<h2>Sharing Your Form</h2>

<h3>Getting the Link</h3>

<ol>
	<li>Click <strong>Share form</strong> in the form builder</li>
	<li>Copy the public form URL</li>
</ol>

<p>This link can be shared via:</p>
<ul>
	<li>Email</li>
	<li>Social media</li>
	<li>Your website</li>
	<li>QR codes</li>
</ul>

<h3>Embedding Forms</h3>

<p>Embed your form in a website:</p>

<ol>
	<li>Click <strong>Share form</strong></li>
	<li>Copy the embed code</li>
	<li>Paste into your website's HTML</li>
</ol>

<h2>Managing Submissions</h2>

<h3>Viewing Submissions</h3>

<p>Form submissions appear as new records in your table:</p>

<ol>
	<li>Go to the table the form is connected to</li>
	<li>New submissions appear as new rows</li>
	<li>Filter by creation date to see recent submissions</li>
</ol>

<h3>Processing Submissions</h3>

<p>Set up a workflow for handling submissions:</p>

<ol>
	<li>Create a "Status" Single Select field</li>
	<li>Default new records to "New" or "Pending"</li>
	<li>Use Kanban view to track processing</li>
	<li>Update status as you handle each submission</li>
</ol>

<h3>Notifications</h3>

<p>Get notified of new submissions:</p>

<ul>
	<li>Set up a <strong>webhook</strong> to send notifications to Slack, email, etc.</li>
	<li>Use the <strong>Activity feed</strong> to monitor new records</li>
</ul>

<h2>Field Types in Forms</h2>

<p>Different field types render differently in forms:</p>

<table>
	<thead>
		<tr>
			<th>Field Type</th>
			<th>Form Input</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Text</td>
			<td>Single-line text input</td>
		</tr>
		<tr>
			<td>Long Text</td>
			<td>Multi-line textarea</td>
		</tr>
		<tr>
			<td>Number</td>
			<td>Number input</td>
		</tr>
		<tr>
			<td>Checkbox</td>
			<td>Checkbox</td>
		</tr>
		<tr>
			<td>Date</td>
			<td>Date picker</td>
		</tr>
		<tr>
			<td>Single Select</td>
			<td>Dropdown</td>
		</tr>
		<tr>
			<td>Multi Select</td>
			<td>Multi-select checkboxes</td>
		</tr>
		<tr>
			<td>Email</td>
			<td>Email input with validation</td>
		</tr>
		<tr>
			<td>URL</td>
			<td>URL input</td>
		</tr>
		<tr>
			<td>Attachment</td>
			<td>File upload</td>
		</tr>
	</tbody>
</table>

<div class="note">
	<strong>Note:</strong> Some field types (like Linked Record, Formula, Rollup) aren't suitable for forms and won't appear as options.
</div>

<h2>Tips and Best Practices</h2>

<div class="tip">
	<strong>Keep it short:</strong> Only ask for essential information. Longer forms have lower completion rates.
</div>

<div class="tip">
	<strong>Use clear labels:</strong> Form visitors don't know your internal naming conventions. Use friendly, descriptive labels.
</div>

<div class="tip">
	<strong>Add help text:</strong> Brief descriptions reduce confusion and incomplete submissions.
</div>

<div class="tip">
	<strong>Test your form:</strong> Submit a test entry to ensure everything works and data appears correctly in your table.
</div>

<div class="tip">
	<strong>Set up processing:</strong> Create a system for handling submissions - don't let them pile up unreviewed.
</div>

<h2>Example: Contact Form</h2>

<p>A typical contact form setup:</p>

<table>
	<thead>
		<tr>
			<th>Field</th>
			<th>Type</th>
			<th>Required</th>
			<th>Show on Form</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Name</td>
			<td>Text</td>
			<td>Yes</td>
			<td>Yes</td>
		</tr>
		<tr>
			<td>Email</td>
			<td>Email</td>
			<td>Yes</td>
			<td>Yes</td>
		</tr>
		<tr>
			<td>Subject</td>
			<td>Single Select</td>
			<td>Yes</td>
			<td>Yes</td>
		</tr>
		<tr>
			<td>Message</td>
			<td>Long Text</td>
			<td>Yes</td>
			<td>Yes</td>
		</tr>
		<tr>
			<td>Status</td>
			<td>Single Select</td>
			<td>No</td>
			<td>No (hidden)</td>
		</tr>
		<tr>
			<td>Submitted</td>
			<td>Date</td>
			<td>No</td>
			<td>No (auto)</td>
		</tr>
	</tbody>
</table>

<h2>Related Topics</h2>

<ul>
	<li><a href="/docs/sharing">Sharing & Permissions</a> - Other ways to share</li>
	<li><a href="/docs/fields">Fields & Field Types</a> - Understanding field types</li>
	<li><a href="/docs/views/kanban">Kanban View</a> - Processing submissions visually</li>
</ul>
