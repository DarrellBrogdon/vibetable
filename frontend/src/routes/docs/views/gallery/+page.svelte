<svelte:head>
	<title>Gallery View - VibeTable Documentation</title>
</svelte:head>

<h1>Gallery View</h1>

<p>The Gallery view displays your records as large cards with image previews, perfect for visual content like portfolios, product catalogs, and media libraries.</p>

<h2>Overview</h2>

<p>Gallery view shows records as visual cards in a grid:</p>

<ul>
	<li><strong>Cover images</strong> - Large image previews from attachment fields</li>
	<li><strong>Titles</strong> - Clear identification of each record</li>
	<li><strong>Details</strong> - Additional fields shown on each card</li>
</ul>

<p>This view is ideal for:</p>
<ul>
	<li>Product catalogs with images</li>
	<li>Portfolio galleries</li>
	<li>Real estate listings</li>
	<li>Team directories with photos</li>
	<li>Media asset libraries</li>
	<li>Recipe collections</li>
</ul>

<h2>Setting Up Gallery</h2>

<h3>Requirements</h3>

<p>Gallery view works best with:</p>

<ul>
	<li>An <strong>Attachment</strong> field containing images (for cover images)</li>
</ul>

<p>Without an attachment field, cards will display without images, showing only field data.</p>

<h3>Creating a Gallery View</h3>

<ol>
	<li>Click <strong>+ Add view</strong> in the view tabs</li>
	<li>Select <strong>Gallery</strong></li>
	<li>Name your view (e.g., "Product Gallery")</li>
	<li>Click <strong>Create</strong></li>
</ol>

<h3>Configuring the Cover Image</h3>

<ol>
	<li>Click the settings icon</li>
	<li>Select <strong>Cover field</strong></li>
	<li>Choose the Attachment field to use</li>
</ol>

<p>The first image in each record's attachment field becomes the card cover.</p>

<h2>Working with Cards</h2>

<h3>Viewing Records</h3>

<p>Click any card to open the expanded record view with:</p>
<ul>
	<li>Full-size images</li>
	<li>All field values</li>
	<li>Comments and activity</li>
</ul>

<h3>Creating Records</h3>

<p>Add new cards:</p>

<ol>
	<li>Click <strong>+ Add record</strong> at the bottom of the gallery</li>
	<li>Fill in the record details</li>
	<li>Upload images to the cover field</li>
	<li>Click <strong>Create</strong></li>
</ol>

<h3>Browsing</h3>

<p>Scroll through the gallery to browse all cards. Cards are arranged in a responsive grid that adjusts to your window size.</p>

<h2>Gallery Configuration</h2>

<h3>Card Size</h3>

<p>Adjust how many cards appear per row:</p>

<ol>
	<li>Click the settings icon</li>
	<li>Select <strong>Card size</strong></li>
	<li>Choose: Small, Medium, or Large</li>
</ol>

<table>
	<thead>
		<tr>
			<th>Size</th>
			<th>Best For</th>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Small</td>
			<td>Large catalogs, quick browsing</td>
		</tr>
		<tr>
			<td>Medium</td>
			<td>Balanced view (default)</td>
		</tr>
		<tr>
			<td>Large</td>
			<td>Detailed previews, fewer items</td>
		</tr>
	</tbody>
</table>

<h3>Cover Fit</h3>

<p>Control how images fit in the card:</p>

<ul>
	<li><strong>Cover</strong> - Image fills the card (may crop edges)</li>
	<li><strong>Contain</strong> - Full image visible (may show letterboxing)</li>
</ul>

<h3>Preview Fields</h3>

<p>Choose which fields appear on cards:</p>

<ol>
	<li>Click the settings icon</li>
	<li>Select <strong>Preview fields</strong></li>
	<li>Toggle fields on/off</li>
</ol>

<p>Keep card content minimal for a clean gallery look. Good preview fields:</p>
<ul>
	<li>Name or title</li>
	<li>Category or type</li>
	<li>Price (for products)</li>
	<li>Date</li>
</ul>

<h3>Hiding Empty Cards</h3>

<p>Show only records with images:</p>

<ol>
	<li>Click <strong>Filter</strong></li>
	<li>Add filter: Cover field "is not empty"</li>
</ol>

<h2>Filtering and Sorting</h2>

<h3>Filters</h3>

<p>Filter to show specific cards:</p>

<ol>
	<li>Click <strong>Filter</strong> in the toolbar</li>
	<li>Add conditions</li>
	<li>Only matching cards appear</li>
</ol>

<p>Example filters for a product gallery:</p>
<ul>
	<li>Category is "Electronics"</li>
	<li>In Stock is checked</li>
	<li>Price is less than $100</li>
</ul>

<h3>Sorting</h3>

<p>Control card order:</p>

<ol>
	<li>Click <strong>Sort</strong> in the toolbar</li>
	<li>Choose field and direction</li>
</ol>

<p>Common sorts:</p>
<ul>
	<li>Name (A→Z)</li>
	<li>Date added (Newest first)</li>
	<li>Price (Low to High)</li>
	<li>Rating (Highest first)</li>
</ul>

<h2>Tips and Best Practices</h2>

<div class="tip">
	<strong>Optimize images:</strong> Large images can slow loading. Consider resizing before uploading, especially for catalogs with many items.
</div>

<div class="tip">
	<strong>Consistent aspect ratios:</strong> For a polished look, use images with consistent aspect ratios (e.g., all 4:3 or all square).
</div>

<div class="tip">
	<strong>Use meaningful titles:</strong> The primary field becomes the card title. Make sure it clearly identifies each item.
</div>

<div class="tip">
	<strong>Create filtered views:</strong> Make multiple gallery views for different categories - "All Products", "Electronics", "Sale Items".
</div>

<h2>Common Use Cases</h2>

<h3>Product Catalog</h3>

<table>
	<tbody>
		<tr>
			<td><strong>Cover field</strong></td>
			<td>Product Images</td>
		</tr>
		<tr>
			<td><strong>Preview fields</strong></td>
			<td>Name, Price, Category</td>
		</tr>
		<tr>
			<td><strong>Sort</strong></td>
			<td>Price (Low to High)</td>
		</tr>
	</tbody>
</table>

<h3>Team Directory</h3>

<table>
	<tbody>
		<tr>
			<td><strong>Cover field</strong></td>
			<td>Headshot</td>
		</tr>
		<tr>
			<td><strong>Preview fields</strong></td>
			<td>Name, Title, Department</td>
		</tr>
		<tr>
			<td><strong>Sort</strong></td>
			<td>Name (A→Z)</td>
		</tr>
	</tbody>
</table>

<h3>Real Estate Listings</h3>

<table>
	<tbody>
		<tr>
			<td><strong>Cover field</strong></td>
			<td>Property Photos</td>
		</tr>
		<tr>
			<td><strong>Preview fields</strong></td>
			<td>Address, Price, Bedrooms, Status</td>
		</tr>
		<tr>
			<td><strong>Filter</strong></td>
			<td>Status is "Available"</td>
		</tr>
	</tbody>
</table>

<h2>Related Topics</h2>

<ul>
	<li><a href="/docs/fields">Fields & Field Types</a> - Setting up Attachment fields</li>
	<li><a href="/docs/views/grid">Grid View</a> - For data entry and editing</li>
	<li><a href="/docs/views/kanban">Kanban View</a> - For status-based workflows</li>
</ul>
