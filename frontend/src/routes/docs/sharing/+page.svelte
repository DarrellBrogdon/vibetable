<svelte:head>
	<title>Sharing & Permissions - VibeTable Documentation</title>
</svelte:head>

<h1>Sharing & Permissions</h1>

<p>Collaborate with others by sharing your bases. VibeTable offers flexible permission levels and multiple ways to share your data.</p>

<h2>Sharing Overview</h2>

<p>VibeTable supports several ways to share:</p>

<ul>
	<li><strong>Invite collaborators</strong> - Give specific people access with defined roles</li>
	<li><strong>Public view links</strong> - Share read-only access with anyone</li>
	<li><strong>Public forms</strong> - Collect data from anyone without giving them access</li>
</ul>

<h2>Inviting Collaborators</h2>

<h3>How to Invite</h3>

<ol>
	<li>Open the base you want to share</li>
	<li>Click the <strong>Share</strong> button in the header</li>
	<li>Enter the person's email address</li>
	<li>Select their role (Owner, Editor, or Viewer)</li>
	<li>Click <strong>Invite</strong></li>
</ol>

<p>The person will receive an email invitation with a link to access the base.</p>

<h3>Permission Roles</h3>

<table>
	<thead>
		<tr>
			<th>Role</th>
			<th>View</th>
			<th>Edit</th>
			<th>Delete Base</th>
			<th>Manage Access</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td><strong>Owner</strong></td>
			<td>Yes</td>
			<td>Yes</td>
			<td>Yes</td>
			<td>Yes</td>
		</tr>
		<tr>
			<td><strong>Editor</strong></td>
			<td>Yes</td>
			<td>Yes</td>
			<td>No</td>
			<td>No</td>
		</tr>
		<tr>
			<td><strong>Viewer</strong></td>
			<td>Yes</td>
			<td>No</td>
			<td>No</td>
			<td>No</td>
		</tr>
	</tbody>
</table>

<h3>Owner</h3>

<p>Owners have full control over the base:</p>
<ul>
	<li>Create, edit, and delete tables</li>
	<li>Create, edit, and delete records</li>
	<li>Manage fields and views</li>
	<li>Invite and remove collaborators</li>
	<li>Change collaborator roles</li>
	<li>Delete the entire base</li>
	<li>Configure webhooks and automations</li>
</ul>

<p>The person who creates a base is automatically the owner. Ownership can be transferred by making another user an owner.</p>

<h3>Editor</h3>

<p>Editors can work with data but not manage access:</p>
<ul>
	<li>Create, edit, and delete records</li>
	<li>Create and modify views</li>
	<li>Add and configure fields</li>
	<li>Comment on records</li>
</ul>

<p>Editors cannot:</p>
<ul>
	<li>Invite or remove collaborators</li>
	<li>Delete the base</li>
	<li>Access webhook configuration</li>
</ul>

<h3>Viewer</h3>

<p>Viewers have read-only access:</p>
<ul>
	<li>View all tables and records</li>
	<li>Switch between views</li>
	<li>Use filters and search</li>
	<li>Export data</li>
</ul>

<p>Viewers cannot:</p>
<ul>
	<li>Add, edit, or delete records</li>
	<li>Modify fields or views</li>
	<li>Invite others</li>
</ul>

<h2>Managing Collaborators</h2>

<h3>Viewing Collaborators</h3>

<ol>
	<li>Click <strong>Share</strong> in the header</li>
	<li>View the list of current collaborators and their roles</li>
</ol>

<h3>Changing Roles</h3>

<ol>
	<li>Open the Share dialog</li>
	<li>Find the collaborator</li>
	<li>Click the role dropdown</li>
	<li>Select the new role</li>
</ol>

<div class="note">
	<strong>Note:</strong> Only owners can change collaborator roles.
</div>

<h3>Removing Collaborators</h3>

<ol>
	<li>Open the Share dialog</li>
	<li>Find the collaborator</li>
	<li>Click <strong>Remove</strong></li>
	<li>Confirm the removal</li>
</ol>

<p>Removed collaborators immediately lose access to the base.</p>

<h2>Public View Links</h2>

<p>Share a read-only view of your data with anyone, without requiring them to sign up:</p>

<h3>Creating a Public Link</h3>

<ol>
	<li>Click <strong>Share</strong> in the header</li>
	<li>Toggle on <strong>Enable public link</strong></li>
	<li>Copy the generated link</li>
</ol>

<p>Anyone with the link can view the data in its current state.</p>

<h3>Public Link Features</h3>

<ul>
	<li><strong>Read-only</strong> - Visitors cannot edit data</li>
	<li><strong>No sign-up required</strong> - Anyone with the link can view</li>
	<li><strong>Real-time updates</strong> - Changes are reflected immediately</li>
	<li><strong>Respects filters</strong> - You can share a filtered view</li>
</ul>

<h3>Disabling Public Links</h3>

<ol>
	<li>Open the Share dialog</li>
	<li>Toggle off <strong>Enable public link</strong></li>
</ol>

<p>The link will immediately stop working.</p>

<div class="warning">
	<strong>Warning:</strong> Anyone with the public link can view your data. Be careful about sharing links containing sensitive information.
</div>

<h2>Real-Time Collaboration</h2>

<p>When multiple people work on the same base:</p>

<ul>
	<li><strong>Presence indicators</strong> - See who else is viewing the base</li>
	<li><strong>Real-time updates</strong> - Changes appear instantly for everyone</li>
	<li><strong>Activity feed</strong> - Track who made changes</li>
</ul>

<h3>Presence Indicator</h3>

<p>The presence indicator in the header shows:</p>
<ul>
	<li>Avatars of active collaborators</li>
	<li>Number of people currently viewing</li>
</ul>

<h3>Handling Conflicts</h3>

<p>When two people edit the same cell:</p>
<ul>
	<li>The most recent change wins</li>
	<li>Changes are saved instantly</li>
	<li>The activity feed records all changes</li>
</ul>

<h2>Activity & History</h2>

<h3>Activity Feed</h3>

<p>Track changes made to your base:</p>

<ol>
	<li>Click <strong>Activity</strong> in the sidebar</li>
	<li>View a chronological list of changes</li>
	<li>See who made each change and when</li>
</ol>

<p>The activity feed shows:</p>
<ul>
	<li>Record creates, updates, and deletes</li>
	<li>Field additions and modifications</li>
	<li>Collaborator joins and role changes</li>
</ul>

<h3>Record Comments</h3>

<p>Collaborate on specific records:</p>

<ol>
	<li>Open the expanded record view</li>
	<li>Scroll to the Comments section</li>
	<li>Add comments and mentions</li>
</ol>

<h2>Best Practices</h2>

<div class="tip">
	<strong>Start with Viewer:</strong> When unsure about permission level, start with Viewer access. You can always upgrade to Editor later.
</div>

<div class="tip">
	<strong>Review access regularly:</strong> Periodically check who has access to your bases, especially for sensitive data.
</div>

<div class="tip">
	<strong>Use public links carefully:</strong> Public links are great for sharing, but remember anyone with the link can view the data.
</div>

<div class="tip">
	<strong>Create view-specific shares:</strong> Create filtered views before sharing public links to control exactly what others see.
</div>

<h2>Related Topics</h2>

<ul>
	<li><a href="/docs/forms">Forms & Public Access</a> - Collect data from anyone</li>
	<li><a href="/docs/bases">Bases & Tables</a> - Understanding base structure</li>
</ul>
