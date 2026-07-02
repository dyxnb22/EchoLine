package conversation

// CanManageMembers returns true if role can invite/kick members.
func CanManageMembers(role Role) bool {
	return role == RoleOwner || role == RoleAdmin
}

// CanPublish returns true if role may send messages in the conversation type.
func CanPublish(convType Type, role Role) bool {
	switch convType {
	case TypeDirect, TypeGroup:
		return role == RoleOwner || role == RoleAdmin || role == RoleMember
	case TypeChannel:
		return role == RoleOwner || role == RoleAdmin
	default:
		return false
	}
}
