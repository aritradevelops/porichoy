-- +goose Up
CREATE TABLE roles (
    id            uuid PRIMARY KEY,
    tenant_id     uuid NOT NULL REFERENCES tenants (id),
    -- null = org-management system role not tied to a specific app's business catalog;
    -- set = that app's business role (PRD §7.2). org_id (roles.org_id) is intentionally
    -- omitted from this pass — organizations don't exist yet.
    app_id        uuid REFERENCES apps (id),
    name          text NOT NULL,
    is_system     boolean NOT NULL DEFAULT false,
    -- permissions/policies are embedded directly on the role as plain arrays, not normalized
    -- into child tables (DATA_MODEL.md `roles`) — editing a role's grants is a single-row
    -- UPDATE, and the runtime authorization check never touches this table directly anyway.
    permissions   text[] NOT NULL DEFAULT '{}',
    policies      jsonb NOT NULL DEFAULT '[]'::jsonb,

    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    created_by    uuid,
    updated_by    uuid,

    deleted_at    timestamptz,
    deleted_by    uuid
);

CREATE INDEX idx_roles_tenant_id ON roles (tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_roles_app_id ON roles (app_id) WHERE deleted_at IS NULL;

-- apps.default_signup_role_id was left FK-less in 20260814000002 since roles didn't exist
-- yet — added now that it does.
ALTER TABLE apps ADD CONSTRAINT fk_apps_default_signup_role
    FOREIGN KEY (default_signup_role_id) REFERENCES roles (id);

CREATE TABLE role_assignments (
    id            uuid PRIMARY KEY,
    -- principal_id is a user or an api_credential, undifferentiated (DATA_MODEL.md §0) —
    -- deliberately no FK constraint, same reasoning as tenants.created_by.
    principal_id  uuid NOT NULL,
    role_id       uuid NOT NULL REFERENCES roles (id),

    created_at    timestamptz NOT NULL DEFAULT now(),
    created_by    uuid,

    deleted_at    timestamptz,
    deleted_by    uuid
);

CREATE INDEX idx_role_assignments_principal_id ON role_assignments (principal_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE role_assignments;
ALTER TABLE apps DROP CONSTRAINT fk_apps_default_signup_role;
DROP TABLE roles;
