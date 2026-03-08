package auth

import "testing"

func TestUserHasRole(t *testing.T) {
	u := &User{Name: "akibar", Roles: []string{"engineer", "oncall"}}

	if !u.HasRole("engineer") {
		t.Error("should have role engineer")
	}
	if u.HasRole("admin") {
		t.Error("should not have role admin")
	}
}

func TestUserHasAnyRole(t *testing.T) {
	u := &User{Name: "akibar", Roles: []string{"engineer"}}

	if !u.HasAnyRole([]string{"admin", "engineer"}) {
		t.Error("should match when any role overlaps")
	}
	if u.HasAnyRole([]string{"admin", "superadmin"}) {
		t.Error("should not match when no role overlaps")
	}
}

func TestAnonymousUser(t *testing.T) {
	u := Anonymous()
	if u.Name != "anonymous" {
		t.Errorf("Name = %q, want %q", u.Name, "anonymous")
	}
	if u.HasRole("engineer") {
		t.Error("anonymous should have no roles")
	}
	if u.Authenticated {
		t.Error("anonymous should not be authenticated")
	}
}
