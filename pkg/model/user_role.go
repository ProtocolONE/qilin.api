package model

import (
	"github.com/ProtocolONE/rbac"
	"github.com/satori/go.uuid"
)

const (
	Admin      string = "admin"
	Manager    string = "manager"
	Support    string = "support"
	Accountant string = "accountant"
	Store      string = "store"
	Publisher  string = "publisher"

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

const GameType string = "game"
const DocumentsType string = "documents"
const RolesType string = "roles"
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
	GetUsers(vendorId uuid.UUID) ([]UserRole, error)
	GetUser(vendorId uuid.UUID, userId string) (*UserRole, error)
	GetUserPermissions(vendorId uuid.UUID, userId string) (*rbac.UserPermissions, error)
	AddRoleToUserInGame(vendorId uuid.UUID, userId string, gameId string, role string) error
	RemoveRoleToUserInGame(vendorId uuid.UUID, userId string, gameId string, role string) error
	SendInvite(vendorId uuid.UUID, userId string) error
	AcceptInvite(inviteId uuid.UUID) error
}
