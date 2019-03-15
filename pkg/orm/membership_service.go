package orm

import (
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/shersh/rbac"
	"net/http"
	"qilin-api/pkg/model"
	"qilin-api/pkg/orm/utils"
)

type membershipService struct {
	db       *Database
	enforcer *rbac.Enforcer
}

func NewMembershipService(db *Database, enforcer *rbac.Enforcer) model.MembershipService {
	return &membershipService{db: db, enforcer: enforcer}
}

func (service *membershipService) Init() error {
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Manager.String(), Domain: "vendor", ResourceType: "game", ResourceId: "*", Action: "any", Effect: "allow"})
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Accountant.String(), Domain: "vendor", ResourceType: "game", ResourceId: "*", Action: "any", Effect: "allow"})
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Publisher.String(), Domain: "vendor", ResourceType: "game", ResourceId: "*", Action: "any", Effect: "allow"})
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Store.String(), Domain: "vendor", ResourceType: "game", ResourceId: "*", Action: "any", Effect: "allow"})
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Support.String(), Domain: "vendor", ResourceType: "game", ResourceId: "*", Action: "any", Effect: "allow"})
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Admin.String(), Domain: "vendor", ResourceType: "game", ResourceId: "*", Action: "any", Effect: "allow"})

	return nil
}

func (service *membershipService) GetUsers(vendorId uuid.UUID) ([]model.UserRole, error) {
	ownerId, err := GetOwnerForVendor(service.db.DB(), vendorId)

	if err != nil {
		return nil, err
	}

	enf := service.enforcer

	//Retrieve all users that have membership for vendor
	roles := []model.GameRole{model.Admin, model.Manager, model.Support, model.Accountant, model.Store}
	users := make([]string, 0)
	for _, role := range roles {
		result := enf.GetUsersForRole(role.String(), model.VendorDomain, ownerId.String())
		users = appendIfMissing(users, result)
	}

	usersRoles := make([]model.UserRole, 0)
	for _, userId := range users {
		userPermissions := enf.GetPermissionsForUser(userId, model.VendorDomain, ownerId.String())
		if userPermissions == nil {
			return nil, NewServiceErrorf(http.StatusInternalServerError, "Could not find permissions for userId `%s` and vendor `%s`", userId, vendorId)
		}
		user := model.User{}
		err := service.db.DB().Model(&model.User{}).Where("id = ?", userId).First(&user).Error
		if err != nil {
			return nil, NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get info about userId `%s`", userId))
		}

		roles := make([]model.RoleRestriction, 0)
		gamesCache := make(map[string]model.ResourceMeta)

		gamesCache["*"] = model.ResourceMeta{
			InternalName: "global",
			Preview:      "",
		}

		for _, role := range userPermissions.Permissions {
			restrictions := role.Restrictions
			if restrictions == nil {
				restrictions = []*rbac.Restriction{
					{
						UUID:  role.UUID,
						Role:  role.Role,
						Owner: ownerId.String(),
					},
				}
			}

			for _, rest := range restrictions {
				meta, ok := gamesCache[rest.UUID]
				if !ok {
					game := model.Game{}
					err = service.db.DB().Model(&model.Game{}).Where("id = ?", rest.UUID).First(&game).Error
					if err != nil {
						return nil, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Get game by id"))
					}
					meta = model.ResourceMeta{
						InternalName: game.InternalName,
						//TODO: add new field to game object
						//Preview: game.Icon
					}
					gamesCache[rest.UUID] = meta
				}

				resType := model.GlobalType
				if rest.UUID != "*" {
					resType = model.GameType
				}

				roles = append(roles, model.RoleRestriction{
					Role:   rest.Role,
					Domain: model.VendorDomain,
					Resource: model.ResourceRestriction{
						Id:    rest.UUID,
						Type:  resType,
						Owner: ownerId.String(),
						Meta:  meta,
					},
				})
			}
		}
		usersRoles = append(usersRoles, model.UserRole{
			Email: user.Email,
			Name:  user.FullName,
			Roles: roles,
		})
	}

	return usersRoles, nil
}

func (service *membershipService) GetUser(vendorId uuid.UUID, userId uuid.UUID) (*model.UserRole, error) {
	return nil, errors.New("Not implemented yet")
}

func (service *membershipService) AddRoleToUserInGame(vendorId uuid.UUID, userId uuid.UUID, gameId uuid.UUID, role model.GameRole) error {
	if exist, err := utils.CheckExists(service.db.DB(), &model.User{}, userId); !(exist && err == nil) {
		if err != nil {
			return NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get user by id `%s`", userId))
		}
		return NewServiceErrorf(http.StatusNotFound, "User `%s` not found", userId)
	}

	isGlobal := gameId == uuid.Nil
	var restrict []string

	if !isGlobal {
		if exist, err := utils.CheckExists(service.db.DB(), &model.Game{}, gameId); !(exist && err == nil) {
			if err != nil {
				return NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get game by id `%s`", gameId))
			}
			return NewServiceErrorf(http.StatusNotFound, "Game `%s` not found", gameId)
		}
		restrict = []string{gameId.String()}
	}

	owner, err := GetOwnerForVendor(service.db.DB(), vendorId)
	if err != nil {
		return err
	}

	if service.enforcer.AddRole(rbac.Role{Role: role.String(), User: userId.String(), Owner: owner.String(), Domain: model.VendorDomain, RestrictedResourceId: restrict}) == false {
		return NewServiceErrorf(http.StatusInternalServerError, "Could not add role `%s` to user `%s`", role.String(), userId.String())
	}

	return nil
}

func (service *membershipService) SendInvite(vendorId uuid.UUID, userId uuid.UUID) error {
	return errors.New("Not implemented yet")
}

func (service *membershipService) AcceptInvite(inviteId uuid.UUID) error {
	return errors.New("Not implemented yet")
}

func appendIfMissing(slice []string, users []string) []string {
	for _, user := range users {
		exist := false
		for _, ele := range slice {
			if ele == user {
				exist = true
				break
			}
		}
		if exist == false {
			slice = append(slice, user)
		}
	}

	return slice
}
