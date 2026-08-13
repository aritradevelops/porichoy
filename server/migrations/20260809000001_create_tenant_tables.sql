-- +goose Up
CREATE TABLE tenants (
    id                     uuid PRIMARY KEY,
    parent_id              uuid REFERENCES tenants (id),
    name                   text NOT NULL,

    logo_url               text NOT NULL DEFAULT '',
    brand_image_url        text,
    -- login_layout/enabled_login_methods hold DATA_MODEL.md enum(...) values as plain
    -- text/jsonb, not a Postgres ENUM type or CHECK constraint — CODING_STANDARDS.md §4
    -- keeps that validation at the Go type level (tenant.LoginLayout), not the database's.
    login_layout           text NOT NULL DEFAULT 'centered',
    mfa_required           boolean NOT NULL DEFAULT false,
    enabled_login_methods  jsonb NOT NULL DEFAULT '[]'::jsonb,
    audit_retention_days   integer NOT NULL DEFAULT 90,

    created_at             timestamptz NOT NULL DEFAULT now(),
    updated_at             timestamptz NOT NULL DEFAULT now(),
    -- created_by/updated_by/deleted_by are principals (a user or api_credential,
    -- undifferentiated, DATA_MODEL.md §0) — deliberately no FK constraint, since Postgres
    -- can't FK one column to two tables. Resolved at the application layer instead.
    created_by             uuid,
    updated_by             uuid,

    deleted_at             timestamptz,
    deleted_by             uuid
);

CREATE INDEX idx_tenants_parent_id ON tenants (parent_id) WHERE deleted_at IS NULL;

CREATE TABLE domain_registries (
    id          uuid PRIMARY KEY,
    tenant_id   uuid NOT NULL REFERENCES tenants (id),
    domain      text NOT NULL,

    created_at  timestamptz NOT NULL DEFAULT now(),
    created_by  uuid,

    deleted_at  timestamptz,
    deleted_by  uuid
);

-- Partial (not plain) unique index: a soft-deleted registration must free the domain up for
-- re-registration (soft delete never hard-deletes the old row, DATA_MODEL.md §0) — a plain
-- unique constraint would permanently block reuse of any domain that was ever removed.
CREATE UNIQUE INDEX idx_domain_registries_domain ON domain_registries (domain) WHERE deleted_at IS NULL;
CREATE INDEX idx_domain_registries_tenant_id ON domain_registries (tenant_id) WHERE deleted_at IS NULL;

CREATE TABLE tenant_provider_credentials (
    id                uuid PRIMARY KEY,
    tenant_id         uuid NOT NULL REFERENCES tenants (id),
    provider_type     text NOT NULL,
    config_encrypted  bytea NOT NULL,

    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    created_by        uuid,
    updated_by        uuid,

    deleted_at        timestamptz,
    deleted_by        uuid
);

-- Same partial-uniqueness reasoning as domain_registries above: at most one active credential
-- per (tenant, provider type), but a soft-deleted one must not block a new one being added.
CREATE UNIQUE INDEX idx_tenant_provider_credentials_tenant_type
    ON tenant_provider_credentials (tenant_id, provider_type) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE tenant_provider_credentials;
DROP TABLE domain_registries;
DROP TABLE tenants;
