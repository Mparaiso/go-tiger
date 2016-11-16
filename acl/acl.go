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

/*
Package acl is an Access Console List (https://en.wikipedia.org/wiki/Access_control_list)
allowing Role/Resource/Priviledge complex authorizations for an application.

This package is a port of https://framework.zend.com/manual/1.12/en/zend.acl.html, the main
difference being that roles do not support multiple inheritance for the time being. Also support
for custom type assertions is a work in progress.
*/
package acl

// Type is a rule type
type Type string

// Operation is a rule management operation
type Operation byte

const (
	// Allow is an allow rule type
	Allow Type = "Allow"
	// Deny is a deny rule type
	Deny Type = "Deny"
)

const (
	// Add is is used when a rule is added to a rule set of an ACL
	Add Operation = iota
	// Remove is used when a rule is removed from an ACL rule set
	Remove
)

// Rule is an ACL rule
type Rule struct {
	ID int64
	Type
	Role
	Resource
	AllPrivileges bool
	Assertion
	Privilege string
}

// ResourceNode is a node in a tree of nodes
type ResourceNode struct {
	ID       int64
	Instance Resource
	Children []Resource
	Parent   Resource
}

// RoleNode is a node in a tree of nodes
type RoleNode struct {
	ID       int64
	Instance Role
	Children []Role
	Parent   Role
}

// ACL is an access control list
type ACL struct {
	RoleTree     map[string]*RoleNode
	ResourceTree map[string]*ResourceNode
	Rules        []*Rule
}

// NewACL returns a new access control list
func NewACL() *ACL {
	acl := &ACL{RoleTree: map[string]*RoleNode{}, ResourceTree: map[string]*ResourceNode{}, Rules: []*Rule{}}
	// By default, deny everything to everybody
	acl.Deny(nil, nil)
	return acl
}

// GetRole returns the Role or nil if not exists
func (acl *ACL) GetRole(role Role) Role {
	if role == nil {
		return nil
	}
	return acl.RoleTree[role.GetRoleID()].Instance
}

// HasRole returns true if ACL has role
func (acl *ACL) HasRole(role Role) bool {
	if acl.GetRole(role) != nil {
		return true
	}
	return false
}

// AddRole add role and its parents to the ACL
func (acl *ACL) AddRole(role Role, parent Role) *ACL {
	roleNode := &RoleNode{Instance: role}
	acl.RoleTree[role.GetRoleID()] = roleNode

	// prevents cyclic dependencies

	if parent != nil {
		if acl.InheritsRole(parent, role) {
			return acl
		}
		roleNode.Parent = parent
		if parentNode, ok := acl.RoleTree[parent.GetRoleID()]; ok {
			parentNode.Children = append(parentNode.Children, role)
		} else {
			// creates an add the parent even if it doesnt exist
			parentNode = &RoleNode{Instance: parent, Children: []Role{role}}
			acl.RoleTree[parent.GetRoleID()] = parentNode
		}
	}

	return acl
}

// RemoveRole removes a role from the role tree
func (acl *ACL) RemoveRole(role Role) *ACL {
	delete(acl.RoleTree, role.GetRoleID())
	return acl
}

// InheritsRole returns true if role inherits from parent
func (acl *ACL) InheritsRole(role, parent Role, direct ...bool) bool {
	if len(direct) == 0 {
		direct = []bool{false}
	}

	if roleNode, ok := acl.RoleTree[role.GetRoleID()]; ok && roleNode.Parent != nil && roleNode.Parent.GetRoleID() == parent.GetRoleID() {
		return true
	}
	if direct[0] == true {
		return false
	}

	if roleNode, ok := acl.RoleTree[role.GetRoleID()]; ok && roleNode.Parent != nil {
		return acl.InheritsRole(roleNode.Parent, parent)
	}

	return false
}

// AddResource add resource and its parent to the ACL
// A resource can only have 1 parent
func (acl *ACL) AddResource(resource Resource, parent ...Resource) *ACL {

	acl.ResourceTree[resource.GetResourceID()] = &ResourceNode{
		Instance: resource,
		Children: []Resource{},
	}
	if len(parent) > 0 {
		// check for potential cyclic dependency before appending a child
		if !acl.InheritsResource(parent[0], resource) {
			acl.ResourceTree[resource.GetResourceID()].Parent = parent[0]
			if parent, ok := acl.ResourceTree[parent[0].GetResourceID()]; ok {
				parent.Children = append(parent.Children, resource)
			}
		}
	}

	return acl
}

// GetResource returns a resource or nil if it doesn't exist
func (acl *ACL) GetResource(resource Resource) Resource {
	if resource == nil {
		return nil
	}
	if node, ok := acl.ResourceTree[resource.GetResourceID()]; ok {
		return node.Instance
	}
	return nil
}

// HasResource returns true if the resource exists
func (acl *ACL) HasResource(resource Resource) bool {
	if resource := acl.GetResource(resource); resource != nil {
		return true
	}
	return false
}
func (acl *ACL) setRule(operation Operation, Type Type, role Role, resource Resource, privileges ...string) *Rule {
	var returnedRule *Rule
	switch operation {
	case Add:
		if len(privileges) > 0 {
			for _, privilege := range privileges {
				returnedRule = &Rule{Type: Type, Role: role, Resource: resource, Privilege: privilege}
				acl.Rules = append(
					[]*Rule{returnedRule}, acl.Rules...)
			}
		} else {
			returnedRule = &Rule{Type: Type, Role: role, Resource: resource, AllPrivileges: true}
			acl.Rules = append([]*Rule{returnedRule}, acl.Rules...)
		}
	case Remove:
		if len(privileges) > 0 {
			for _, privilege := range privileges {
				for i, rule := range acl.Rules {
					if rule.Type == Type && role.GetRoleID() == rule.GetRoleID() && rule.GetResourceID() == resource.GetResourceID() && privilege == rule.Privilege && rule.Assertion == nil && rule.AllPrivileges == false {
						returnedRule = acl.Rules[i]
						acl.Rules = append(acl.Rules[0:i], acl.Rules[i+1:len(acl.Rules)]...)
					}
				}
			}
		} else {
			for i, rule := range acl.Rules {
				if rule.Type == Type && role.GetRoleID() == rule.GetRoleID() && rule.GetResourceID() == resource.GetResourceID() && rule.AllPrivileges == true {
					returnedRule = acl.Rules[i]
					acl.Rules = append(acl.Rules[0:i], acl.Rules[i+1:len(acl.Rules)]...)
				}
			}
		}
	}
	return returnedRule
}

// IsAllowed return true if role is allowed all privileges on resource
// When multiple priviledges are checked, ALL priviledges must be allowed.
func (acl *ACL) IsAllowed(role Role, resource Resource, privileges ...string) bool {
	authorization := []bool{}
	if len(privileges) > 0 {
		for _, privilege := range privileges {
			if !acl.isAllowed(role, resource, privilege) {
				return false
			} else {
				authorization = append(authorization, true)
			}
		}
		if len(authorization) == len(privileges) {
			return true
		}
		return false
	}
	return acl.isAllowed(role, resource, "")
}

func (acl *ACL) isAllowed(role Role, resource Resource, privilege string) bool {
	for _, rule := range acl.Rules {
		if ((rule.Role != nil && role != nil) && /* if neither roles are nil , then either the roles are equal or role inherits from rule.Role */
			((role.GetRoleID() == rule.GetRoleID()) ||
				acl.InheritsRole(role, rule.Role))) ||
			rule.Role == nil { /* or rule.Role is nil so rule applies to all roles */
			if ((rule.Resource != nil && resource != nil) && /* if neither resources are nil and */
				((rule.GetResourceID() == resource.GetResourceID()) || acl.InheritsResource(resource, rule.Resource))) ||
				/* either both resources are equal OR resource inherits from rule.Resource */
				(rule.Resource == nil) { /* or the rule applies to app resources */
				if rule.AllPrivileges || rule.Privilege == privilege { /* if this rule applies to all priviledges OR priviledges are equal */
					if rule.Type == Deny {
						return false
					}
					return true
				}
			}
		}
	}
	return false
}

// Allow adds an allow rule
func (acl *ACL) Allow(role Role, resource Resource, privilege ...string) *Rule {
	return acl.setRule(Add, Allow, role, resource, privilege...)
}

// Deny adds a deny rule
func (acl *ACL) Deny(role Role, resource Resource, privilege ...string) *Rule {
	return acl.setRule(Add, Deny, role, resource, privilege...)
}

// RemoveAllow removes a allow rule
func (acl *ACL) RemoveAllow(role Role, resource Resource, privilege ...string) *Rule {
	return acl.setRule(Remove, Allow, role, resource, privilege...)
}

// RemoveDeny deny removes a deny rule
func (acl *ACL) RemoveDeny(role Role, resource Resource, privilege ...string) *Rule {
	return acl.setRule(Remove, Deny, role, resource, privilege...)
}

// InheritsResource retruns true if resource is a child of parent
func (acl *ACL) InheritsResource(resource, parent Resource, direct ...bool) bool {
	if len(direct) == 0 {
		direct = []bool{false}
	}
	if !acl.HasResource(resource) {
		return false
	}
	if parentResource := acl.ResourceTree[resource.GetResourceID()].Parent; parentResource != nil && parentResource.GetResourceID() == parent.GetResourceID() {
		return true
	} else if direct[0] == true {
		return false
	} else if found := acl.GetResource(parentResource); found != nil {
		return acl.InheritsResource(found, parent)
	}
	return false
}

// Role is an ACL role
type Role interface {
	GetRoleID() string
}
type defaultRole struct {
	roleID string
}

// NewRole returns a Role
func NewRole(id string) Role {
	return defaultRole{id}
}
func (role defaultRole) GetRoleID() string {
	return role.roleID
}

// Resource is an ACL resource
type Resource interface {
	GetResourceID() string
}
type defaultResource struct {
	resourceID string
}

// NewResource returns a resource from a string
func NewResource(id string) Resource {
	return defaultResource{id}
}

func (resource defaultResource) GetResourceID() string {
	return resource.resourceID
}

type Assertion interface {
	Assert(acl ACL, role Role, resource Resource, permission string)
}
