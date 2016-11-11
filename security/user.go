package security

// User is an application user
type User interface {
	GetLogin() string
	GetPassword() string
	GetRoles() []string
	IsEnabled() bool
	IsAccountNonLocked() bool
	IsCredentialsNonExpired() bool
	IsAccountNonExpired() bool
}

// DefaultUser is the default implementation of User
type DefaultUser struct {
	Login                 string
	Password              string
	Enabled               bool
	Roles                 []string
	AccountNonLocked      bool
	CredentialsNonExpired bool
	AccountNonExpired     bool
}

func (user DefaultUser) GetPassword() string {
	return user.Password
}
func (user DefaultUser) GetLogin() string {
	return user.Login
}
func (user DefaultUser) GetRoles() []string {
	return user.Roles
}

func (user DefaultUser) IsEnabled() bool {
	return user.Enabled
}

func (user DefaultUser) IsAccountNonLocked() bool {
	return user.AccountNonLocked
}

func (user DefaultUser) IsCredentialsNonExpired() bool {
	return user.CredentialsNonExpired
}

func (user DefaultUser) IsAccountNonExpired() bool {
	return user.AccountNonExpired
}
