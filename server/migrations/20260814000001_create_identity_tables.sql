-- +goose Up
CREATE TABLE users (
    id                    uuid PRIMARY KEY,
    tenant_id             uuid NOT NULL REFERENCES tenants (id),
    -- status/draft_expires_at hold DATA_MODEL.md enum(...) values as plain text, validated at
    -- the Go type level (identity.UserStatus), not a Postgres ENUM/CHECK — same convention as
    -- tenants.login_layout.
    status                text NOT NULL DEFAULT 'active',
    draft_expires_at      timestamptz,
    display_name          text,
    profile_picture_url   text,
    email                 text,
    email_verified        boolean NOT NULL DEFAULT false,
    pending_email         text,
    phone                 text,
    phone_verified        boolean NOT NULL DEFAULT false,
    pending_phone         text,

    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    -- created_by/updated_by/deleted_by are principals (a user or api_credential,
    -- undifferentiated, DATA_MODEL.md §0) — deliberately no FK constraint, same reasoning as
    -- tenants.created_by.
    created_by            uuid,
    updated_by            uuid,

    deleted_at            timestamptz,
    deleted_by            uuid
);

CREATE INDEX idx_users_tenant_id ON users (tenant_id) WHERE deleted_at IS NULL;
-- Not in DATA_MODEL.md's original users table definition — added because signup needs a real
-- DB-level guard against two active accounts sharing an email within a tenant, not just an
-- app-level pre-check (which is TOCTOU-racy under concurrent signups). Partial so a
-- soft-deleted account's email frees up for reuse, same reasoning as domain_registries.domain.
CREATE UNIQUE INDEX idx_users_tenant_email ON users (tenant_id, email)
    WHERE deleted_at IS NULL AND email IS NOT NULL;

CREATE TABLE passwords (
    id             uuid PRIMARY KEY,
    user_id        uuid NOT NULL REFERENCES users (id),
    password_hash  text NOT NULL,

    created_at     timestamptz NOT NULL DEFAULT now(),

    deleted_at     timestamptz,
    deleted_by     uuid
);

-- Not 1:1 with users — a user has many rows over time via soft delete, exactly one active
-- (deleted_at IS NULL); this is also the password-reuse-check history, no separate table
-- (DATA_MODEL.md `passwords`).
CREATE UNIQUE INDEX idx_passwords_user_id ON passwords (user_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE passwords;
DROP TABLE users;
