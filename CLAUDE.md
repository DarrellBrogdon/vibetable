# CLAUDE.md - VibeTable Project Context

## The Experiment

This is a **true vibe coding** experiment. The human will never look at the code. They chose technologies they're unfamiliar with (Go, SvelteKit) specifically to prevent themselves from being tempted to hand-fix anything.

**You (Claude) are the sole developer.** The human is the product visionary and will only interact with the running application.

### What This Means

- **You must be autonomous.** The human cannot debug code, fix syntax errors, or understand stack traces.
- **You must verify your own work.** Test endpoints, check the UI, confirm things actually work.
- **You must handle errors gracefully.** When something breaks, diagnose and fix it yourself.
- **Communicate in outcomes, not code.** Say "the grid now supports inline editing" not "I added an event handler to the cell component."
- **The human trusts you.** Don't ask for permission on implementation details — make good decisions.

---

## Project Overview

**VibeTable** is an Airtable clone. The human is documenting this journey on social media to show what's achievable with pure vibe coding.

### Target Features (Priority Order)

1. **Authentication** — Magic link (passwordless email login)
2. **Bases & Tables** — Create, rename, delete bases with multiple tables
3. **Grid view** — Spreadsheet-like UI with inline editing
4. **Field types** — Text, number, checkbox, date, single/multi-select
5. **Collaboration** — Share bases with other users (viewer/editor roles)
6. **Linked records** — Relationships between tables
7. **Filtering and sorting** — Query and organize data
8. **Kanban view** — Cards grouped by a select field

### Non-Goals

- Mobile responsiveness (desktop-first)
- Enterprise-scale performance optimization
- Self-hosting documentation

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go with Chi router |
| Database | PostgreSQL 16 |
| Frontend | SvelteKit 2 |
| Styling | Vanilla CSS with custom properties |
| Email | Resend (or similar transactional email API) |
| Dev Environment | Docker Compose |

### Email for Magic Links

Magic link auth requires sending emails. Options:
- **Resend** — Simple API, generous free tier (3k emails/month)
- **Postmark** — Reliable, good deliverability
- **SendGrid** — Well-known, free tier available

For development, log magic link URLs to console instead of sending real emails.

All services run in Docker. The human will only ever run:
```bash
docker-compose up --build
```

---

## Architecture

```
Frontend (localhost:5173) → Backend API (localhost:8080) → PostgreSQL (5432)
```

### Backend Structure

```
backend/
├── main.go
├── internal/
│   ├── api/handlers/    # HTTP handlers
│   ├── models/          # Data structures
│   └── store/           # Database operations
├── migrations/          # SQL files
└── go.mod
```

### Frontend Structure

```
frontend/src/
├── routes/              # SvelteKit pages
├── lib/
│   ├── components/      # Reusable UI
│   ├── stores/          # Shared state
│   ├── types/           # TypeScript types
│   └── api/             # API client
└── app.css              # Global styles
```

---

## Data Model

### Entities

**User** — A registered user
- id, email, name, created_at, updated_at

**Session** — Active login session
- id, user_id, token, expires_at, created_at

**MagicLink** — Pending email verification
- id, email, token, expires_at, created_at

**Base** — Container for tables (like a spreadsheet file)
- id, name, created_by (user_id), created_at, updated_at

**BaseCollaborator** — Shared access to a base
- id, base_id, user_id, role (owner/editor/viewer), created_at

**Table** — A sheet within a base
- id, base_id, name, position

**Field** — Column definition
- id, table_id, name, type, options (JSONB), position

**Record** — A row of data
- id, table_id, values (JSONB mapping field_id → value), position

**View** — Saved view configuration
- id, table_id, name, type (grid/kanban), config (JSONB)

### Collaboration Roles

| Role | Can View | Can Edit | Can Delete Base | Can Manage Collaborators |
|------|----------|----------|-----------------|--------------------------|
| owner | ✓ | ✓ | ✓ | ✓ |
| editor | ✓ | ✓ | ✗ | ✗ |
| viewer | ✓ | ✗ | ✗ | ✗ |

### Field Types

| Type | Storage | Options |
|------|---------|---------|
| text | string | — |
| number | number | precision, format |
| checkbox | boolean | — |
| date | ISO string | include_time |
| single_select | option_id | options array with id, name, color |
| multi_select | option_id[] | options array with id, name, color |
| linked_record | record_id[] | linked_table_id |

---

## API Endpoints

Base URL: `http://localhost:8080/api/v1`

```
# Authentication
POST   /auth/magic-link          # Request magic link (send email)
POST   /auth/verify              # Verify magic link token, create session
GET    /auth/me                  # Get current user
POST   /auth/logout              # End session

# Bases
GET    /bases                    # List bases (user has access to)
POST   /bases                    # Create base (user becomes owner)
GET    /bases/:id                # Get base with tables
PATCH  /bases/:id                # Update base
DELETE /bases/:id                # Delete base (owner only)

# Collaboration
GET    /bases/:id/collaborators  # List collaborators
POST   /bases/:id/collaborators  # Invite collaborator (by email)
PATCH  /bases/:id/collaborators/:userId   # Change role
DELETE /bases/:id/collaborators/:userId   # Remove collaborator

# Tables
POST   /bases/:baseId/tables
GET    /tables/:id
PATCH  /tables/:id
DELETE /tables/:id

# Fields
GET    /tables/:tableId/fields
POST   /tables/:tableId/fields
PATCH  /fields/:id
DELETE /fields/:id

# Records
GET    /tables/:tableId/records
POST   /tables/:tableId/records
PATCH  /records/:id
DELETE /records/:id

# Views
GET    /tables/:tableId/views
POST   /tables/:tableId/views
PATCH  /views/:id
DELETE /views/:id
```

### Authentication Flow

1. User enters email → `POST /auth/magic-link`
2. System sends email with link containing token
3. User clicks link → `POST /auth/verify` with token
4. Server returns session token (stored in HTTP-only cookie or returned for localStorage)
5. All subsequent requests include session token

---

## Current Status

### Completed
- [x] Docker Compose setup
- [x] Go API server with health endpoint
- [x] SvelteKit app with landing page
- [x] Database connectivity

### Next Up
- [ ] Database migrations (users, sessions, bases, tables, fields, records)
- [ ] Magic link authentication
- [ ] User session management
- [ ] Bases CRUD (with ownership)
- [ ] Base sharing/collaboration
- [ ] Tables CRUD
- [ ] Fields with type support
- [ ] Records with JSONB storage
- [ ] Grid view UI
- [ ] Inline editing
- [ ] Filtering/sorting
- [ ] Kanban view

---

## Your Operating Guidelines

### 1. Be Fully Autonomous

The human cannot help you debug. When something fails:
- Read the error carefully
- Check logs (`docker-compose logs backend`)
- Fix the issue yourself
- Verify the fix works

### 2. Verify Your Own Work

After making changes:
- Restart services if needed (`docker-compose up --build`)
- Test API endpoints with curl or by hitting them from the frontend
- Check the browser to confirm UI changes work
- Don't just assume code is correct — prove it

### 3. Communicate in Plain Language

**Good:** "You can now create bases. Try clicking the 'New Base' button."

**Bad:** "I added a POST handler that validates the request body and returns a 201 with the created resource."

### 4. Make Decisions, Don't Ask

The human trusts your judgment on:
- Code organization
- Library choices (within the stack)
- Implementation approach
- Error handling patterns
- UI/UX details not explicitly specified

### 5. Flag Content-Worthy Moments

This project is being documented publicly. Flag moments like:
- Major features working for the first time
- Particularly interesting challenges
- Visual milestones (screenshot-worthy UI states)
- Anything that would make a good story

Say something like: "📸 **Content moment:** The grid now has working inline editing. This would make a great demo video."

### 6. Keep Scope Tight

The human has ~3 weeks. When you see an opportunity to over-engineer:
- Don't
- Build the minimal working version first
- Note what could be improved later

### 7. Manage the Docker Environment

Remember:
- Changes to Go code require: `docker-compose up --build backend`
- Changes to frontend often hot-reload, but if stuck: `docker-compose restart frontend`
- Database schema changes may require: `docker-compose down -v` (resets data)
- If things are weird: `docker-compose down && docker-compose up --build`

---

## Design Reference

VibeTable should feel like Airtable:
- Clean, minimal interface
- Blue primary color (#2d7ff9)
- Lots of white space
- Subtle shadows and borders
- Professional but friendly

The color palette and base styles are in `/frontend/src/app.css`.

---

## File Locations

| What | Where |
|------|-------|
| This file | `/CLAUDE.md` |
| Docker config | `/docker-compose.yml` |
| Backend entry | `/backend/main.go` |
| Frontend pages | `/frontend/src/routes/` |
| Global styles | `/frontend/src/app.css` |
| Components | `/frontend/src/lib/components/` |
