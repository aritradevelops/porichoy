//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTenant(name string, parentID *uuid.UUID) *tenant.Tenant {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return &tenant.Tenant{
		ID:                  uuid.New(),
		ParentID:            parentID,
		Name:                name,
		LoginLayout:         tenant.LoginLayoutCentered,
		EnabledLoginMethods: []tenant.LoginMethod{tenant.LoginMethodEmailPassword, tenant.LoginMethodGoogle},
		AuditRetentionDays:  90,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// mustCreateRoot creates a top-level (no parent) tenant via the pre-authentication bootstrap
// path — a convenient fixture for "a tenant with no parent", not a claim that it's THE root.
func mustCreateRoot(t *testing.T, repo *TenantRepository, name string) *tenant.Tenant {
	t.Helper()
	tt := newTestTenant(name, nil)
	require.NoError(t, repo.CreateRoot(context.Background(), tt))
	return tt
}

// mustCreateChild creates a tenant under parentID via the authorized path, using act. Fails
// the test immediately if act isn't authorized to create under parentID.
func mustCreateChild(t *testing.T, repo *TenantRepository, act actor.Actor, name string, parentID uuid.UUID) *tenant.Tenant {
	t.Helper()
	tt := newTestTenant(name, &parentID)
	require.NoError(t, repo.Create(context.Background(), act, tt))
	return tt
}

func rootActor() actor.Actor {
	return actor.Actor{PrincipalID: uuid.New(), Scope: actor.ScopeRoot}
}

func TestTenantRepository_CreateAndFindByID(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, repo, "Brand A")

	got, err := repo.FindByID(ctx, tt.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Brand A", got.Name)
	assert.True(t, got.IsRoot())
	assert.Empty(t, got.Ancestors)
	assert.Equal(t, tenant.LoginLayoutCentered, got.LoginLayout)
	assert.ElementsMatch(t, tt.EnabledLoginMethods, got.EnabledLoginMethods)
	assert.Equal(t, 90, got.AuditRetentionDays)
}

func TestTenantRepository_FindByID_NotFound(t *testing.T) {
	repo := NewTenantRepository(testDB)

	got, err := repo.FindByID(context.Background(), uuid.New())

	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestTenantRepository_Create_ComputesAncestorsFromParent(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	grandparent := mustCreateRoot(t, repo, "Grandparent")
	parent := mustCreateChild(t, repo, root, "Parent", grandparent.ID)
	child := mustCreateChild(t, repo, root, "Child", parent.ID)

	got, err := repo.FindByID(ctx, child.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.ElementsMatch(t, []uuid.UUID{grandparent.ID, parent.ID}, got.Ancestors)
}

func TestTenantRepository_Create_RootCanCreateUnderAnyTenant(t *testing.T) {
	repo := NewTenantRepository(testDB)
	unrelated := mustCreateRoot(t, repo, "Unrelated")

	child := mustCreateChild(t, repo, rootActor(), "Child of Unrelated", unrelated.ID)

	got, err := repo.FindByID(context.Background(), child.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, []uuid.UUID{unrelated.ID}, got.Ancestors)
}

func TestTenantRepository_Create_TenantScopeCanCreateUnderOwnDescendant(t *testing.T) {
	repo := NewTenantRepository(testDB)
	root := rootActor()

	self := mustCreateRoot(t, repo, "Self")
	grandchild := mustCreateChild(t, repo, root, "Grandchild",
		mustCreateChild(t, repo, root, "Child", self.ID).ID)

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}
	newTenant := newTestTenant("Great-grandchild", &grandchild.ID)
	err := repo.Create(context.Background(), act, newTenant)

	require.NoError(t, err)
	wantAncestors := append(append([]uuid.UUID{}, grandchild.Ancestors...), grandchild.ID)
	assert.ElementsMatch(t, wantAncestors, newTenant.Ancestors)
}

func TestTenantRepository_Create_TenantScopeDeniedForUnrelatedParent(t *testing.T) {
	repo := NewTenantRepository(testDB)
	self := mustCreateRoot(t, repo, "Self")
	unrelated := mustCreateRoot(t, repo, "Unrelated")

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}
	err := repo.Create(context.Background(), act, newTestTenant("Nope", &unrelated.ID))

	require.ErrorIs(t, err, tenant.ErrTenantNotFound)
}

func TestTenantRepository_Create_NonTenantScopeDenied(t *testing.T) {
	repo := NewTenantRepository(testDB)
	self := mustCreateRoot(t, repo, "Self")

	for _, s := range []actor.Scope{actor.ScopeOrg, actor.ScopeApp, actor.ScopeOwn} {
		act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: s}
		err := repo.Create(context.Background(), act, newTestTenant("Nope", &self.ID))
		require.ErrorIsf(t, err, tenant.ErrTenantNotFound, "scope %q should be denied", s)
	}
}

func TestTenantRepository_GetByID_SelfAndDescendantsAllowed(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	self := mustCreateRoot(t, repo, "Self")
	child := mustCreateChild(t, repo, root, "Child", self.ID)
	grandchild := mustCreateChild(t, repo, root, "Grandchild", child.ID)

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}

	gotSelf, err := repo.GetByID(ctx, act, self.ID)
	require.NoError(t, err)
	assert.NotNil(t, gotSelf)

	gotGrandchild, err := repo.GetByID(ctx, act, grandchild.ID)
	require.NoError(t, err)
	assert.NotNil(t, gotGrandchild)
}

func TestTenantRepository_GetByID_UnrelatedTenantDenied(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()

	self := mustCreateRoot(t, repo, "Self")
	unrelated := mustCreateRoot(t, repo, "Unrelated")

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}
	got, err := repo.GetByID(ctx, act, unrelated.ID)

	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestTenantRepository_GetByID_NonTenantScopeDenied(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()
	self := mustCreateRoot(t, repo, "Self")

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeOrg}
	got, err := repo.GetByID(ctx, act, self.ID)

	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestTenantRepository_Update(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, repo, "Original")

	tt.Name = "Renamed"
	tt.MFARequired = true
	require.NoError(t, repo.Update(ctx, tt))

	got, err := repo.FindByID(ctx, tt.ID)
	require.NoError(t, err)
	assert.Equal(t, "Renamed", got.Name)
	assert.True(t, got.MFARequired)
}

func TestTenantRepository_SoftDelete_ExcludesFromFindByID(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, repo, "To Delete")

	require.NoError(t, repo.SoftDelete(ctx, rootActor(), tt.ID))

	got, err := repo.FindByID(ctx, tt.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestTenantRepository_SoftDelete_DeniedOutsideSubtreeIsNoop(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()

	self := mustCreateRoot(t, repo, "Self")
	unrelated := mustCreateRoot(t, repo, "Unrelated")

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}
	require.NoError(t, repo.SoftDelete(ctx, act, unrelated.ID))

	got, err := repo.FindByID(ctx, unrelated.ID)
	require.NoError(t, err)
	require.NotNil(t, got, "unrelated tenant must survive a denied delete attempt")
}

func TestTenantRepository_SoftDelete_StripsAncestorReferenceFromDescendants(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	grandparent := mustCreateRoot(t, repo, "Grandparent")
	parent := mustCreateChild(t, repo, root, "Parent", grandparent.ID)
	child := mustCreateChild(t, repo, root, "Child", parent.ID)

	require.NoError(t, repo.SoftDelete(ctx, root, parent.ID))

	got, err := repo.FindByID(ctx, child.ID)
	require.NoError(t, err)
	require.NotNil(t, got, "soft-deleting an ancestor must not cascade-delete descendants")
	assert.NotContains(t, got.Ancestors, parent.ID)
	assert.Contains(t, got.Ancestors, grandparent.ID)
}

func TestTenantRepository_ListChildren(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	parent := mustCreateRoot(t, repo, "Parent")
	child := mustCreateChild(t, repo, root, "Child", parent.ID)
	mustCreateRoot(t, repo, "Unrelated")

	children, err := repo.ListChildren(ctx, root, parent.ID)

	require.NoError(t, err)
	require.Len(t, children, 1)
	assert.Equal(t, child.Name, children[0].Name)
}

func TestTenantRepository_ListChildren_GrandparentScopeReachesGrandchildrenOfSelf(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	self := mustCreateRoot(t, repo, "Self")
	child := mustCreateChild(t, repo, root, "Child", self.ID)
	grandchild := mustCreateChild(t, repo, root, "Grandchild", child.ID)

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}
	children, err := repo.ListChildren(ctx, act, child.ID)

	require.NoError(t, err)
	require.Len(t, children, 1)
	assert.Equal(t, grandchild.Name, children[0].Name)
}

func TestTenantRepository_ListChildren_UnrelatedParentDenied(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	self := mustCreateRoot(t, repo, "Self")
	unrelated := mustCreateRoot(t, repo, "Unrelated")
	mustCreateChild(t, repo, root, "Unrelated's Child", unrelated.ID)

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}
	children, err := repo.ListChildren(ctx, act, unrelated.ID)

	require.NoError(t, err)
	assert.Empty(t, children)
}

func TestTenantRepository_List_RootSeesEveryTenant(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	parent := mustCreateRoot(t, repo, "Parent")
	child := mustCreateChild(t, repo, root, "Child", parent.ID)
	other := mustCreateRoot(t, repo, "Other")

	got, err := repo.List(ctx, root, tenant.ListParams{Page: 1, PageSize: 1000})

	require.NoError(t, err)
	ids := make([]uuid.UUID, len(got.Items))
	for i, tt := range got.Items {
		ids[i] = tt.ID
	}
	assert.Subset(t, ids, []uuid.UUID{parent.ID, child.ID, other.ID})
	assert.GreaterOrEqual(t, got.Total, 3)
}

func TestTenantRepository_List_TenantScopeSeesOnlySelfAndDescendants(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	self := mustCreateRoot(t, repo, "Self")
	child := mustCreateChild(t, repo, root, "Child", self.ID)
	grandchild := mustCreateChild(t, repo, root, "Grandchild", child.ID)
	unrelated := mustCreateRoot(t, repo, "Unrelated")

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}
	got, err := repo.List(ctx, act, tenant.ListParams{Page: 1, PageSize: 1000})

	require.NoError(t, err)
	assert.Equal(t, 3, got.Total)
	ids := make([]uuid.UUID, len(got.Items))
	for i, tt := range got.Items {
		ids[i] = tt.ID
	}
	assert.ElementsMatch(t, []uuid.UUID{self.ID, child.ID, grandchild.ID}, ids)
	assert.NotContains(t, ids, unrelated.ID)
}

func TestTenantRepository_List_NonTenantScopeDenied(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()
	self := mustCreateRoot(t, repo, "Self")

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeOrg}
	got, err := repo.List(ctx, act, tenant.ListParams{Page: 1, PageSize: 10})

	require.NoError(t, err)
	assert.Equal(t, 0, got.Total)
	assert.Empty(t, got.Items)
}

func TestTenantRepository_List_PaginatesWithoutDuplicatesOrGaps(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	self := mustCreateRoot(t, repo, "Paginate Self")
	wantIDs := []uuid.UUID{self.ID}
	for i := range 5 {
		child := mustCreateChild(t, repo, root, fmt.Sprintf("Paginate Child %d", i), self.ID)
		wantIDs = append(wantIDs, child.ID)
	}
	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}

	seen := map[uuid.UUID]bool{}
	for page := 1; page <= 4; page++ {
		got, err := repo.List(ctx, act, tenant.ListParams{Page: page, PageSize: 2})
		require.NoError(t, err)
		assert.Equal(t, 6, got.Total)
		if page <= 3 {
			assert.Len(t, got.Items, 2)
		} else {
			assert.Empty(t, got.Items, "page beyond the last full page should be empty")
		}
		for _, tt := range got.Items {
			assert.Falsef(t, seen[tt.ID], "tenant %s returned on more than one page", tt.ID)
			seen[tt.ID] = true
		}
	}
	gotIDs := make([]uuid.UUID, 0, len(seen))
	for id := range seen {
		gotIDs = append(gotIDs, id)
	}
	assert.ElementsMatch(t, wantIDs, gotIDs)
}
