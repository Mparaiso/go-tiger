package security_test

import (
	"reflect"
	"testing"

	"github.com/Mparaiso/go-tiger/security"
	"github.com/Mparaiso/go-tiger/test"
)

func TestUser(t *testing.T) {

	user := &security.DefaultUser{Login: "john doe", Password: "password"}
	_, ok := (interface{})(user).(security.User)
	test.Fatal(t, ok, true)
	test.Fatal(t, user.GetLogin(), "john doe")
	test.Fatal(t, user.GetPassword(), "password")
	test.Fatal(t, len(user.GetRoles()) == 0, true)
	user = &security.DefaultUser{Login: "john doe", Password: "password", Roles: []string{"ROLE_ADMIN"}}
	test.Fatal(t, reflect.DeepEqual(user.GetRoles(), []string{"ROLE_ADMIN"}), true)
	user = &security.DefaultUser{Login: "john doe", Password: "password", Enabled: false}
	test.Fatal(t, user.IsEnabled(), false)
}
