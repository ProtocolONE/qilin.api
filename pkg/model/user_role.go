package model

import "github.com/satori/go.uuid"

type GameRole string

type Domain string

const (
	Admin      GameRole = "admin"
	Manager    GameRole = "manager"
	Support    GameRole = "support"
	Accountant GameRole = "accountant"
	Store      GameRole = "store"
	Publisher  GameRole = "publisher"

	VendorDomain string = "vendor"
)

func (role GameRole) String() string {
	return string(role)
}

func (domain Domain) String() string {
	return string(domain)
}

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

type ResourceType string

const GameType string = "game"
const DocumentsType string = "documents"
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
	GetUser(vendorId uuid.UUID, userId uuid.UUID) (*UserRole, error)
	AddRoleToUserInGame(vendorId uuid.UUID, userId uuid.UUID, gameId uuid.UUID, role GameRole) error
	SendInvite(vendorId uuid.UUID, userId uuid.UUID) error
	AcceptInvite(inviteId uuid.UUID) error
}

func (t ResourceType) String() string {
	return string(t)
}
