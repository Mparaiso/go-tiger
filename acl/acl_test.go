package acl_test

import (
	"testing"

	"github.com/Mparaiso/go-tiger/acl"
	"github.com/Mparaiso/go-tiger/test"
)

const (
	allowed = "allowed"
	denied  = "denied"
)

// func ExampleACL() {
// 	roles := []acl.Role{
// 		acl.NewRole("guest"),
// 		acl.NewRole("member"),
// 		acl.NewRole("admin"),
// 	}
// 	userRole := acl.NewRole("someUser")
// 	resources := []acl.Resource{
// 		acl.NewResource("someResource"),
// 	}
// 	acl := acl.NewACL()
// 	acl.AddRoles(roles)

// 	acl.AddRole(userRole, roles...)

// 	acl.AddResource(resources[0])

// 	acl.Deny(roles[0], resources[0])
// 	acl.Allow(roles[1], resource[0])

// 	fmt.Println(ternary(acl.isAllowed(userRole, resources[0]), allowed, denied))

// 	// Output:
// 	// allowed

// }

func TestACL(t *testing.T) {
	roles := map[string]acl.Role{
		"guest":         acl.NewRole("guest"),
		"staff":         acl.NewRole("staff"),
		"editor":        acl.NewRole("editor"),
		"administrator": acl.NewRole("administrator"),
	}
	list := acl.NewACL()
	list.AddRole(roles["guest"], nil)
	list.AddRole(roles["staff"], roles["guest"])
	list.AddRole(roles["editor"], roles["staff"])
	list.AddRole(roles["administrator"], nil)

	list.Allow(roles["guest"], nil, "view")
	list.Allow(roles["staff"], nil, "edit", "submit", "revise")
	list.Allow(roles["editor"], nil, "publish", "archive", "delete")
	list.Allow(roles["administrator"], nil)

	test.Error(t, list.IsAllowed(roles["guest"], nil, "view"), true)
	test.Error(t, list.IsAllowed(roles["staff"], nil, "publish"), false)
	test.Error(t, list.IsAllowed(roles["staff"], nil, "revise"), true)
	test.Error(t, list.IsAllowed(roles["editor"], nil, "view"), true)
	test.Error(t, list.IsAllowed(roles["editor"], nil, "update"), false)
	test.Error(t, list.IsAllowed(roles["administrator"], nil, "view"), true)
	test.Error(t, list.IsAllowed(roles["administrator"], nil), true)
	test.Error(t, list.IsAllowed(roles["administrator"], nil, "update"), true)

	/** Precise Access Controls
	 * @link https://framework.zend.com/manual/1.12/en/zend.acl.refining.html
	 */
	roles["marketing"] = acl.NewRole("marketing")
	list.AddRole(roles["marketing"], roles["staff"])

	resources := map[string]acl.Resource{
		"news":         acl.NewResource("news"),
		"latest":       acl.NewResource("latest"),
		"newsletter":   acl.NewResource("newsletter"),
		"announcement": acl.NewResource("announcement"),
	}
	list.AddResource(resources["newsletter"])
	list.AddResource(resources["news"])
	list.AddResource(resources["latest"], resources["news"])
	list.AddResource(resources["announcement"], resources["news"])
	list.Allow(roles["marketing"], resources["newsletter"], "publish", "archive")
	list.Allow(roles["marketing"], resources["latest"], "publish", "archive")
	list.Deny(roles["staff"], resources["latest"], "revise")
	list.Deny(nil, resources["announcement"], "archive")

	test.Error(t, list.IsAllowed(roles["staff"], resources["newsletter"], "publish"), false)
	test.Error(t, list.IsAllowed(roles["marketing"], resources["newsletter"], "publish"), true)
	test.Error(t, list.IsAllowed(roles["staff"], resources["latest"], "publish"), false)
	test.Error(t, list.IsAllowed(roles["marketing"], resources["latest"], "publish"), true)
	test.Error(t, list.IsAllowed(roles["marketing"], resources["latest"], "archive"), true)
	test.Error(t, list.IsAllowed(roles["editor"], resources["announcement"], "archive"), false)
	test.Error(t, list.IsAllowed(roles["administrator"], resources["announcement"], "archive"), false)
	test.Error(t, list.IsAllowed(roles["marketing"], nil, "view"), true)

}

// ternary operator helper
func ternary(predicate bool, TrueValue interface{}, FalseValue interface{}) interface{} {
	if predicate {
		return TrueValue
	}
	return FalseValue
}
