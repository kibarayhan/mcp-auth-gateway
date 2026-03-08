package auth

// User represents an authenticated caller.
type User struct {
	Name          string
	Roles         []string
	Authenticated bool
}

// HasRole returns true if the user has the given role.
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole returns true if the user has at least one of the given roles.
func (u *User) HasAnyRole(roles []string) bool {
	for _, role := range roles {
		if u.HasRole(role) {
			return true
		}
	}
	return false
}

// Anonymous returns an unauthenticated user with no roles.
func Anonymous() *User {
	return &User{Name: "anonymous", Authenticated: false}
}
