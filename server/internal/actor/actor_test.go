package actor

import "testing"

func TestScope_AtLeast(t *testing.T) {
	order := []Scope{ScopeOwn, ScopeOrg, ScopeApp, ScopeTenant, ScopeRoot}

	for i, s := range order {
		for j, other := range order {
			want := i >= j
			if got := s.AtLeast(other); got != want {
				t.Errorf("%s.AtLeast(%s) = %v, want %v", s, other, got, want)
			}
		}
	}
}
