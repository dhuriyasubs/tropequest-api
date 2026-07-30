# TropeQuest API

Backend API service for [TropeQuest](https://tropequest.com), an AI-native book
discovery platform for romance, fantasy, and thriller readers.

## What this does

Serves catalog data, trope/mood/vibe-based search, and content endpoints to the
TropeQuest web frontend. Built in Go for performance and simplicity.

## Architecture

Part of a three-service architecture:
- **tropequest-web** — frontend (JavaScript)
- **tropequest-api** — this repo, backend API (Go)
- **tropequest-pipeline** — catalog ingestion and enrichment (Python)

Every data-flow decision (what's served live via API vs. what's precomputed and
served statically) was made deliberately based on freshness needs and load
patterns, not by default.

## Infrastructure

Deployed on a traffic server positioned closer to the primary user base (India)
to reduce latency. Integrated with GA4 alongside a self-built internal tracking
system for engagement and traffic-quality monitoring.

## Built by

Architected and built independently, using Claude Code as the engineering tool.
System design, API structure, and infrastructure decisions made and owned
directly.
