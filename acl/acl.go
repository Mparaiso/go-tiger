package acl

// Type is a rule type
type Type byte

// Operation is a rule management operation
type Operation byte

const (
	// Allow is an allow rule type
	Allow Type = iota
	// Deny is a deny rule type
	Deny
)

const (
	// Add is is used when a rule is added to a rule set of an ACL
	Add Operation = iota
	// Remove is used when a rule is removed from an ACL rule set
	Remove
)

// Rule is an ACL rule
type Rule struct {
	Type
	Role
	Resource
	AllPriviledges bool
	Assertion
	Priviledge string
}

// ResourceNode is a node in a tree of nodes
type ResourceNode struct {
	Instance Resource
	Children []Resource
	Parent   Resource
}

// RoleNode is a node in a tree of nodes
type RoleNode struct {
	Instance Role
	Children []Role
	Parent   Role
}

// ACL is an access control list
type ACL struct {
	RoleTree     map[string]*RoleNode
	ResourceTree map[string]*ResourceNode
	rules        []*Rule
}

// NewACL returns a new access control list
func NewACL() *ACL {
	return &ACL{RoleTree: map[string]*RoleNode{}, ResourceTree: map[string]*ResourceNode{}, rules: []*Rule{}}
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
func (acl *ACL) setRule(operation Operation, Type Type, role Role, resource Resource, priviledges ...string) *ACL {

	switch operation {
	case Add:
		if len(priviledges) > 0 {
			for _, priviledge := range priviledges {
				acl.rules = append(
					[]*Rule{{Type: Type, Role: role, Resource: resource, Priviledge: priviledge}},
					acl.rules...)
			}
		} else {
			acl.rules = append([]*Rule{{Type: Type, Role: role, Resource: resource, AllPriviledges: true}}, acl.rules...)
		}
	case Remove:
		if len(priviledges) > 0 {
			for _, priviledge := range priviledges {
				for i, rule := range acl.rules {
					if rule.Type == Type && role.GetRoleID() == rule.GetRoleID() && rule.GetResourceID() == resource.GetResourceID() && priviledge == rule.Priviledge && rule.Assertion == nil && rule.AllPriviledges == false {
						acl.rules = append(acl.rules[0:i], acl.rules[i+1:len(acl.rules)]...)
					}
				}
			}
		} else {
			for i, rule := range acl.rules {
				if rule.Type == Type && role.GetRoleID() == rule.GetRoleID() && rule.GetResourceID() == resource.GetResourceID() && rule.AllPriviledges == true {
					acl.rules = append(acl.rules[0:i], acl.rules[i+1:len(acl.rules)]...)
				}
			}
		}
	}
	return acl
}

// IsAllowed return true if role is allowed all priviledges on resource
func (acl *ACL) IsAllowed(role Role, resource Resource, priviledges ...string) bool {
	for _, priviledge := range priviledges {
		if !acl.isAllowed(role, resource, priviledge) {
			return false
		}
	}
	return true
}

func (acl *ACL) isAllowed(role Role, resource Resource, priviledge string) bool {
	// check for a direct rule
	for _, rule := range acl.rules {
		if (rule.Role != nil && role != nil && role.GetRoleID() == rule.GetRoleID()) || (rule.Role == nil) {
			if (rule.Resource != nil && resource != nil && rule.GetResourceID() == resource.GetResourceID()) || (rule.Resource == nil) {
				if rule.AllPriviledges || rule.Priviledge == priviledge {
					if rule.Type == Deny {
						return false
					}
					return true
				}
			}

		}
	}
	// check for a rule on the resource's parent
	if resource != nil && acl.HasResource(resource) {
		if node := acl.ResourceTree[resource.GetResourceID()]; node != nil {
			if node.Parent != nil {
				return acl.isAllowed(role, node.Parent, priviledge)
			}
		}
	}
	// check for a rule on the role's parent
	if role != nil && acl.HasRole(role) {
		if node := acl.RoleTree[role.GetRoleID()]; node != nil {
			if node.Parent != nil {
				return acl.isAllowed(node.Parent, resource, priviledge)
			}
		}
	}
	return false
}

// Allow adds an allow rule
func (acl *ACL) Allow(role Role, resource Resource, priviledge ...string) *ACL {
	return acl.setRule(Add, Allow, role, resource, priviledge...)
}

// Deny adds a deny rule
func (acl *ACL) Deny(role Role, resource Resource, priviledge ...string) *ACL {
	return acl.setRule(Add, Deny, role, resource, priviledge...)
}

// RemoveAllow removes a allow rule
func (acl *ACL) RemoveAllow(role Role, resource Resource, priviledge ...string) *ACL {
	return acl.setRule(Remove, Allow, role, resource, priviledge...)
}

// RemoveDeny deny removes a deny rule
func (acl *ACL) RemoveDeny(role Role, resource Resource, priviledge ...string) *ACL {
	return acl.setRule(Remove, Deny, role, resource, priviledge...)
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
