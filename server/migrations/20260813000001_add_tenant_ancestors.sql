-- +goose Up
-- ancestors precomputes the tenant's full chain of ancestor IDs (root-most first is not
-- required — order is irrelevant, only membership matters) at creation time: a new tenant's
-- ancestors is its parent's own ancestors plus the parent's ID, no recursive query needed
-- (DATA_MODEL.md `tenants`). This is what makes descendant-scope authorization a single
-- containment check instead of a recursive tree walk.
ALTER TABLE tenants ADD COLUMN ancestors uuid[] NOT NULL DEFAULT '{}';

-- GIN index so `ancestors @> ARRAY[?]::uuid[]` containment checks (the only form that can
-- use this index — `? = ANY(ancestors)` cannot) stay fast as the tree grows.
CREATE INDEX idx_tenants_ancestors ON tenants USING GIN (ancestors);

-- +goose Down
DROP INDEX idx_tenants_ancestors;
ALTER TABLE tenants DROP COLUMN ancestors;
