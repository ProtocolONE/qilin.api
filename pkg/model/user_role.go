package model

import (
	"github.com/ProtocolONE/rbac"
	"github.com/satori/go.uuid"
)

const (
	Admin      string = "admin"
	SuperAdmin string = "super_admin"
	Manager    string = "manager"
	Support    string = "support"
	Developer  string = "developer"
	Accountant string = "accountant"
	Store      string = "store"
	Publisher  string = "publisher"

	AnyRole     string = "any"
	NotApproved string = "not_approved"

	VendorDomain string = "vendor"
)

type UserRole struct {
	Email string            `json:"email"`
	Name  string            `json:"name"`
	Roles []RoleRestriction `json:"roles"`
}

type RoleRestriction struct {
	Role     string              `json:"role"`
	Domain   string              `json:"domain"`
	Resource ResourceRestriction `json:"resource"`
}

const GameType string = "games"
const GameListType string = "vendor.games"
const DocumentsType string = "vendor.documents"
const MessagesType string = "vendor.messages"
const VendorType string = "vendor"
const AdminDocumentsType string = "admin.vendor.documents"
const RoleUserType string = "roles.user"
const RolesType string = "roles"
const PackageType string = "packages"
const GlobalType string = "global"

type ResourceMeta struct {
	Preview      string `json:"preview"`
	InternalName string `json:"internalName"`
}

type ResourceRestriction struct {
	Id    string       `json:"id"`
	Type  string       `json:"type"`
	Owner string       `json:"owner"`
	Meta  ResourceMeta `json:"meta"`
}

type MembershipService interface {
	Init() error
	GetUsers(vendorId uuid.UUID) ([]*UserRole, error)
	GetUser(vendorId uuid.UUID, userId string) (*UserRole, error)
	GetUserPermissions(vendorId uuid.UUID, userId string) (*rbac.UserPermissions, error)
	AddRoleToUserInGame(vendorId uuid.UUID, userId string, gameId string, role string) error
	AddRoleToUser(userId string, owner string, role string) error
	RemoveRoleToUserInGame(vendorId uuid.UUID, userId string, gameId string, role string) error
	SendInvite(vendorId uuid.UUID, invite Invite) (*InviteCreated, error)
	AcceptInvite(vendorId uuid.UUID, inviteId uuid.UUID, userId string) error
}
