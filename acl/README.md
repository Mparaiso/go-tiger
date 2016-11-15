acl
===

[![GoDoc](https://godoc.org/github.com/Mparaiso/go-tiger/acl?status.png)](https://godoc.org/github.com/Mparaiso/go-tiger/acl)


author: mparaiso <mparaiso@online.fr>

license: APACHE 2-0

ACL is a port of [Zend/ACL](https://github.com/zf1/zend-acl) to Golang . 
ACL stands for [Access Control List](https://en.wikipedia.org/wiki/Access_control_list) and helps 
the developer design complex application authorization capabilities.

### Install

	go get github.com/Mparaiso/go-tiger/acl
	
### Basic Usage

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