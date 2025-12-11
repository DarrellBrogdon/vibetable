# VibeTable

An Airtable clone built with vibe coding — experimenting with AI-assisted development using unfamiliar technologies.

## Tech Stack

- **Backend:** Go
- **Frontend:** SvelteKit
- **Database:** PostgreSQL
- **Containerization:** Docker

## Features (Target)

- [x] Project setup
- [ ] Single base with multiple tables
- [ ] Grid view with inline editing
- [ ] Field types: text, number, checkbox, date, single/multi-select
- [ ] Linked records (relationships between tables)
- [ ] Filtering and sorting
- [ ] Kanban view

## Getting Started

### Prerequisites

- Docker & Docker Compose

### Run the Project

```bash
# Start all services
docker-compose up --build

# Frontend: http://localhost:5173
# Backend API: http://localhost:8080
# Database: localhost:5432
```

### Development

```bash
# Stop services
docker-compose down

# Stop and remove volumes (reset database)
docker-compose down -v
```

## Project Structure

```
vibetable/
├── backend/          # Go API server
├── frontend/         # SvelteKit app
├── docker-compose.yml
└── README.md
```

## The Experiment

This project documents the journey of "vibe coding" an enterprise application clone — using AI assistance while working in unfamiliar languages and frameworks.

Follow along: [Your X/LinkedIn/Blog links here]
