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
	VendorOwner string = "owner"

	VendorDomain string = "vendor"
)

type UserRole struct {
	ID       string            `json:"id"`
	Email    string            `json:"email"`
	Name     string            `json:"name"`
	Roles    []RoleRestriction `json:"roles"`
	LastSeen string            `json:"lastSeen"`
}

type RoleRestriction struct {
	Role     string              `json:"role"`
	Domain   string              `json:"domain"`
	Resource ResourceRestriction `json:"resource"`
}

const GameType string = "games"
const PublishGame string = "games.publish"
const GameListType string = "vendors.games"
const VendorGameType string = "vendor.games"
const DocumentsType string = "vendors.documents.*"
const MessagesType string = "vendors.messages.*"
const VendorType string = "vendors"
const AdminDocumentsType string = "admin.vendors.*"
const RoleUserType string = "vendors.memberships"
const RolesType string = "vendors.memberships.permissions"
const InvitesType string = "vendors.memberships.invites"
const GlobalType string = "global"
const PackageType string = "packages"
const RoleBundle string = "bundles"
const PackageListType string = "vendors.packages.*"
const RoleBundleList string = "vendors.bundles.*"

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
	RemoveUserRole(userId string, owner string, role string) error
	RemoveRoleToUserInGame(vendorId uuid.UUID, userId string, gameId string, role string) error
	SendInvite(vendorId uuid.UUID, invite Invite) (*InviteCreated, error)
	AcceptInvite(vendorId uuid.UUID, inviteId uuid.UUID, userId string) error
	AddRoleToUserInResource(vendorId uuid.UUID, userId string, resourceId []string, role string) error
	RemoveRoleToUserInResource(vendorId uuid.UUID, userId string, resourceId []string, role string) error
}
