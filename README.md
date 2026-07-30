# TropeQuest API

Go/Gin HTTP API exposing books, tropes, and search endpoints. Originally built for a Flutter client (CORS is open, GET-only).

> The main web app ([tropequest-web](../tropequest-web)) reads from Supabase directly and does not depend on this service.

## Stack

- Go 1.25, Gin
- Backend selected at startup, first available wins:
  1. Supabase REST API (`SUPABASE_URL` + `SUPABASE_SERVICE_KEY`)
  2. Direct Postgres (`SUPABASE_DB_URL`, via `lib/pq`)
  3. Google Sheets CSV fallback (no env vars needed)

## Structure

- `main.go` — service selection, CORS, routes
- `handlers/` — request handlers (`search.go`)
- `models/` — data models (`book.go`)
- `services/` — backend implementations: `supabase.go`, `db.go`, `sheets.go`, `fuzzy.go`
- `schema.sql`, `user_schema.sql` — reference Postgres schema

## Endpoints

- `GET /api/books`
- `GET /api/tropes`
- `GET /api/search`

## Setup

```bash
go mod download
export SUPABASE_URL=...
export SUPABASE_SERVICE_KEY=...
go run main.go   # serves on :8080
```

## Deploy

Docker build via `Dockerfile`; deployed on Render (`render.yaml`, `singapore` region, health check on `/api/books`).

```bash
docker build -t tropequest-api .
docker run -p 8080:8080 -e SUPABASE_URL=... -e SUPABASE_SERVICE_KEY=... tropequest-api
```
