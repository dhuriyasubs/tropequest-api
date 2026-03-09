-- TropeQuest Supabase PostgreSQL Schema
-- Run this in the Supabase SQL editor or via psql

-- Enable the pg_trgm extension for full-text / trigram search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ─────────────────────────────────────────────
-- books table
-- ─────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS books (
    id               SERIAL PRIMARY KEY,
    title            TEXT        NOT NULL,
    author           TEXT        NOT NULL,
    tropes           TEXT[]      NOT NULL DEFAULT '{}',
    description      TEXT,
    cover_url        TEXT,
    book_id          TEXT        UNIQUE,
    buy_url_amazon_us TEXT,
    buy_url_amazon_in TEXT,
    buy_url_flipkart  TEXT,
    rating           NUMERIC(3,2),
    rating_count     INTEGER,
    rating_source    TEXT,
    subjects         TEXT[]      NOT NULL DEFAULT '{}',
    language         TEXT,
    series_name      TEXT,
    series_number    TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────────
-- Indexes
-- ─────────────────────────────────────────────

-- GIN index on tropes array — fast "contains any" / overlap queries
CREATE INDEX IF NOT EXISTS books_tropes_gin_idx
    ON books USING GIN (tropes);

-- GIN index on subjects array
CREATE INDEX IF NOT EXISTS books_subjects_gin_idx
    ON books USING GIN (subjects);

-- Full-text search index on title + author (tsvector)
CREATE INDEX IF NOT EXISTS books_fts_idx
    ON books USING GIN (
        to_tsvector('english', coalesce(title, '') || ' ' || coalesce(author, ''))
    );

-- Trigram indexes for ILIKE / partial-match on title and author
CREATE INDEX IF NOT EXISTS books_title_trgm_idx
    ON books USING GIN (title gin_trgm_ops);

CREATE INDEX IF NOT EXISTS books_author_trgm_idx
    ON books USING GIN (author gin_trgm_ops);

-- ─────────────────────────────────────────────
-- Auto-update updated_at on row changes
-- ─────────────────────────────────────────────
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS books_set_updated_at ON books;
CREATE TRIGGER books_set_updated_at
    BEFORE UPDATE ON books
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─────────────────────────────────────────────
-- Row Level Security
-- ─────────────────────────────────────────────
ALTER TABLE books ENABLE ROW LEVEL SECURITY;

-- Allow anyone (including unauthenticated / anon key) to SELECT
CREATE POLICY "Public read access"
    ON books
    FOR SELECT
    USING (true);

-- Only authenticated service-role can INSERT / UPDATE / DELETE
-- (the migration script uses the direct DB URL with full privileges,
--  so this policy is not needed for writes — it just blocks anon writes
--  through the Supabase API layer)
