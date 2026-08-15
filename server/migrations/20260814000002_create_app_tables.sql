-- +goose Up
CREATE TABLE apps (
    id                             uuid PRIMARY KEY,
    tenant_id                      uuid NOT NULL REFERENCES tenants (id),
    name                           text NOT NULL,
    client_id                      text NOT NULL,
    client_secret_hash             text,
    redirect_uris                  jsonb NOT NULL DEFAULT '[]'::jsonb,
    logo_url                       text,
    is_system                      boolean NOT NULL DEFAULT false,
    supports_organizations         boolean NOT NULL DEFAULT false,
    -- References roles.id, but roles doesn't exist yet (roles.app_id itself references
    -- apps.id — the two tables reference each other, so one FK has to be added after the
    -- fact). See 20260814000003, which ALTERs this column in once roles exists.
    default_signup_role_id         uuid,
    -- signing_algorithm/signing_key_config_encrypted mirror tenants.login_layout's
    -- convention: validated at the Go type level (app.SigningAlgorithm), not a Postgres
    -- ENUM/CHECK. signing_key_config_encrypted holds the raw HS256 secret unencrypted for
    -- now (no encryption port yet) — same column/shape the future encrypted form will use.
    signing_algorithm              text NOT NULL,
    signing_key_config_encrypted   bytea NOT NULL,
    access_token_ttl_seconds       integer NOT NULL DEFAULT 900,
    id_token_ttl_seconds           integer NOT NULL DEFAULT 900,
    refresh_token_ttl_seconds      integer NOT NULL DEFAULT 2592000,

    created_at                     timestamptz NOT NULL DEFAULT now(),
    updated_at                     timestamptz NOT NULL DEFAULT now(),
    created_by                     uuid,
    updated_by                     uuid,

    deleted_at                     timestamptz,
    deleted_by                     uuid
);

CREATE UNIQUE INDEX idx_apps_client_id ON apps (client_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_apps_tenant_id ON apps (tenant_id) WHERE deleted_at IS NULL;
-- Signup resolves "the tenant's default system app" by tenant_id alone — at most one
-- is_system row per tenant, enforced here rather than just assumed.
CREATE UNIQUE INDEX idx_apps_tenant_system ON apps (tenant_id) WHERE is_system AND deleted_at IS NULL;

CREATE TABLE sessions (
    id                  uuid PRIMARY KEY,
    user_id             uuid NOT NULL REFERENCES users (id),
    app_id              uuid NOT NULL REFERENCES apps (id),
    aud                 text NOT NULL,
    refresh_token_hash  text NOT NULL,
    device_label        text,
    ip_address          inet,

    created_at          timestamptz NOT NULL DEFAULT now(),
    last_active_at      timestamptz NOT NULL DEFAULT now(),
    expires_at          timestamptz NOT NULL,

    -- No created_by: a session's creator is always its own user_id (DATA_MODEL.md
    -- `sessions`), tracking that separately would be redundant.
    deleted_at          timestamptz,
    deleted_by          uuid
);

CREATE INDEX idx_sessions_user_id ON sessions (user_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE sessions;
DROP TABLE apps;
