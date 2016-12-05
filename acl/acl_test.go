//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package acl_test

import (
	"fmt"
	"github.com/Mparaiso/go-tiger/acl"
)

func Example() {
	// http: //book.cakephp.org/2.0/en/core-libraries/components/access-control-lists.html#creating-access-request-objects-aros-and-access-control-objects-acos
	Acl := acl.NewAcl()
	groups := []acl.Group{{Alias: "warrior"}, {Alias: "wizard"}, {Alias: "hobbits"}, {Alias: "visitors"}}
	for _, group := range groups {
		Acl.SaveAro(group)
	}
	users := []acl.Group{
		{
			Alias:      "Aragorn",
			ParentID:   1,
			Model:      "User",
			ForeignKey: 2356,
		},
		{
			Alias:      "Legolas",
			ParentID:   1,
			Model:      "User",
			ForeignKey: 6342,
		},
		{
			Alias:      "Gimli",
			ParentID:   1,
			Model:      "User",
			ForeignKey: 1564,
		},
		{
			Alias:      "Gandalf",
			ParentID:   2,
			Model:      "User",
			ForeignKey: 7419,
		},
		{
			Alias:      "Frodo",
			ParentID:   3,
			Model:      "User",
			ForeignKey: 7451,
		},
		{
			Alias:      "Bilbo",
			ParentID:   3,
			Model:      "User",
			ForeignKey: 5126,
		},
		{
			Alias:      "Merry",
			ParentID:   3,
			Model:      "User",
			ForeignKey: 5144,
		},
		{
			Alias:      "Pippin",
			ParentID:   3,
			Model:      "User",
			ForeignKey: 1211,
		},
		{
			Alias:      "Gollum",
			ParentID:   4,
			Model:      "User",
			ForeignKey: 1337,
		},
	}
	for _, user := range users {
		Acl.SaveAro(user)
	}
	resources := []acl.Group{{Alias: "Weapons"}, {Alias: "Rings"}, {Alias: "PorkChops"}, {Alias: "DiplomaticEfforts"}, {Alias: "Ales"}}
	for _, resource := range resources {
		Acl.SaveAco(resource)
	}
	Acl.Allow(acl.String("warriors"), acl.String("Weapons"))
	Acl.Deny(acl.String("warriors/Legolas"), acl.String("Weapons"), "delete")
	Acl.Deny(acl.String("warriors/Gimli"), acl.String("Weapons"), "delete")
	// alternative syntax
	Acl.Deny(acl.Group{Model: "User", ForeignKey: 3439}, acl.String("Weapons"), "delete")
	Acl.Deny(acl.Group{Model: "User", ForeignKey: 439}, acl.String("Weapons"), "delete")
	// Checking Permissions
	// http://book.cakephp.org/2.0/en/core-libraries/components/access-control-lists.html#checking-permissions-the-acl-component

	fmt.Println(Acl.Check(acl.String("warriors/Aragorn"), acl.String("Weapons")))
	fmt.Println(Acl.Check(acl.String("warriors/Aragorn"), acl.String("Weapons"), "create"))
	fmt.Println(Acl.Check(acl.String("warriors/Aragorn"), acl.String("Weapons"), "read"))
	fmt.Println(Acl.Check(acl.String("warriors/Aragorn"), acl.String("Weapons"), "update"))
	fmt.Println(Acl.Check(acl.String("warriors/Aragorn"), acl.String("Weapons"), "delete"))

	// Remember, we can use the model/id syntax
	// for our user AROs
	fmt.Println(Acl.Check(acl.Group{Model: "User", ID: 2356}, acl.String("Weapons")))

	// These also return true:
	fmt.Println(Acl.Check(acl.String("warriors/Legolas"), acl.String("Weapons"), "create"))
	fmt.Println(Acl.Check(acl.String("warriors/Gimli"), acl.String("Weapons"), "read"))

	// But these return false:
	fmt.Println(Acl.Check(acl.String("warriors/Legolas"), acl.String("Weapons"), "delete"))
	fmt.Println(Acl.Check(acl.String("warriors/Gimli"), acl.String("Weapons"), "delete"))

	// Output:
	// <nil>
	// <nil>
	// <nil>
	// <nil>
	// <nil>
	// <nil>
	// <nil>
	// <nil>
	// ErrAccessDenied : Error access denied
	// ErrAccessDenied : Error access denied
}

/*
import (
	"fmt"
	"testing"

	"github.com/Mparaiso/go-tiger/acl"
	"github.com/Mparaiso/go-tiger/test"
)

const (
	allowed = "allowed"
	denied  = "denied"
)

func ExampleACL() {
	roles := map[string]acl.Role{
		"guest": acl.NewRole("guest"),
	}
	resources := map[string]acl.Resource{
		"article": acl.NewResource("article"),
	}
	acl := acl.NewACL()

	acl.AddResource(resources["article"])
	acl.Allow(roles["guest"], resources["article"])
	fmt.Println(ternary(acl.IsAllowed(roles["guest"], resources["article"]), allowed, denied))
	fmt.Println(ternary(acl.IsAllowed(roles["anonymous"], resources["article"]), allowed, denied))
	// Output:
	// allowed
	// denied

}

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

	// Precise Access Controls
	 // @link https://framework.zend.com/manual/1.12/en/zend.acl.refining.html
	 //
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

*/
