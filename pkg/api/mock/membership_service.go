package mock

import (
	"github.com/ProtocolONE/rbac"
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model"
)

type memebershipService struct {
}

func (memebershipService) Init() error {
	return nil
}

func (memebershipService) GetUsers(vendorId uuid.UUID) ([]*model.UserRole, error) {
	return nil, nil
}

func (memebershipService) GetUser(vendorId uuid.UUID, userId string) (*model.UserRole, error) {
	return nil, nil
}

func (memebershipService) GetUserPermissions(vendorId uuid.UUID, userId string) (*rbac.UserPermissions, error) {
	return nil, nil
}

func (memebershipService) AddRoleToUserInGame(vendorId uuid.UUID, userId string, gameId string, role string) error {
	return nil
}

func (memebershipService) AddRoleToUserInResource(vendorId uuid.UUID, userId string, resourceId []string, role string) error {
	return nil
}

func (memebershipService) AddRoleToUser(userId string, owner string, role string) error {
	return nil
}

func (memebershipService) RemoveRoleToUserInGame(vendorId uuid.UUID, userId string, gameId string, role string) error {
	return nil
}

func (memebershipService) RemoveRoleToUserInResource(vendorId uuid.UUID, userId string, resourceId []string, role string) error {
	return nil
}

func (memebershipService) SendInvite(vendorId uuid.UUID, invite model.Invite) (*model.InviteCreated, error) {
	return nil, nil
}

func (memebershipService) AcceptInvite(vendorId uuid.UUID, inviteId uuid.UUID, userId string) error {
	return nil
}

func (memebershipService) GetInvites(vendorId uuid.UUID, offset, limit int) (total int, invites []model.Invite, err error) {
	return 0, []model.Invite{}, nil
}

func NewMembershipService() model.MembershipService {
	return &memebershipService{}
}

func (memebershipService) RemoveUserRole(userId string, owner string, role string) error {
	return nil
}
