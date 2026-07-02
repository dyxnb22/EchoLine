package conversation

import "testing"

func TestCanPublish(t *testing.T) {
	if !CanPublish(TypeGroup, RoleMember) {
		t.Fatal("group member should publish")
	}
	if CanPublish(TypeChannel, RoleSubscriber) {
		t.Fatal("channel subscriber should not publish")
	}
	if !CanPublish(TypeChannel, RoleOwner) {
		t.Fatal("channel owner should publish")
	}
}

func TestCanManageMembers(t *testing.T) {
	if !CanManageMembers(RoleAdmin) {
		t.Fatal("admin should manage members")
	}
	if CanManageMembers(RoleMember) {
		t.Fatal("member should not manage members")
	}
}
