package orm

import (
	"github.com/ProtocolONE/rbac"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
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
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Support, Domain: "vendor", ResourceType: model.GameType, ResourceId: "*", Action: "read", Effect: "allow"})
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Support, Domain: "vendor", ResourceType: model.GameListType, ResourceId: "skip", Action: "read", Effect: "allow"})
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Support, Domain: "vendor", ResourceType: model.VendorType, ResourceId: "skip", Action: "read", Effect: "allow"})

	service.enforcer.AddPolicy(rbac.Policy{Role: model.Admin, Domain: "vendor", ResourceType: model.GameType, ResourceId: "*", Action: "any", Effect: "allow"})
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Admin, Domain: "vendor", ResourceType: model.GameListType, ResourceId: "skip", Action: "any", Effect: "allow"})
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Admin, Domain: "vendor", ResourceType: model.MessagesType, ResourceId: "skip", Action: "any", Effect: "allow"})
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Admin, Domain: "vendor", ResourceType: model.DocumentsType, ResourceId: "skip", Action: "any", Effect: "allow"})
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Admin, Domain: "vendor", ResourceType: model.RolesType, ResourceId: "skip", Action: "read", Effect: "allow"})
	service.enforcer.AddPolicy(rbac.Policy{Role: model.Admin, Domain: "vendor", ResourceType: model.VendorType, ResourceId: "skip", Action: "any", Effect: "allow"})

	service.enforcer.LinkRoles(model.SuperAdmin, model.Admin, "vendor")
	service.enforcer.AddPolicy(rbac.Policy{Role: model.SuperAdmin, Domain: "vendor", ResourceType: model.GameListType, ResourceId: "skip", Action: "any", Effect: "allow"})
	service.enforcer.AddPolicy(rbac.Policy{Role: model.SuperAdmin, Domain: "vendor", ResourceType: model.RolesType, ResourceId: "skip", Action: "any", Effect: "allow"})

	return nil
}

func (service *membershipService) GetUsers(vendorId uuid.UUID) ([]*model.UserRole, error) {
	ownerId, err := GetOwnerForVendor(service.db.DB(), vendorId)

	if err != nil {
		return nil, err
	}

	enf := service.enforcer

	//Retrieve all users that have membership for vendor
	roles := []string{model.Admin, model.Manager, model.Support, model.Accountant, model.Store, model.Developer, model.Publisher}
	namesToSkip := []string{model.Admin, model.Manager, model.Support, model.Accountant, model.Store, model.Developer, model.Publisher, model.SuperAdmin}
	users := make([]string, 0)
	for _, role := range roles {
		result := enf.GetUsersForRole(role, model.VendorDomain, ownerId)
		users = appendIfMissing(users, result, namesToSkip)
	}

	usersRoles := make([]*model.UserRole, 0)
	for _, userId := range users {
		user, err := service.getUser(userId, ownerId)
		if err != nil {
			return nil, err
		}
		usersRoles = append(usersRoles, user)
	}

	return usersRoles, nil
}

func (service *membershipService) getUser(userId string, ownerId string) (*model.UserRole, error) {
	user := model.User{}
	err := service.db.DB().Model(&model.User{}).Where("id = ?", userId).First(&user).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, NewServiceErrorf(http.StatusNotFound, "User %s not found", userId)
	} else if err != nil {
		return nil, NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get info about userId `%s`", userId))
	}

	userPermissions := service.enforcer.GetPermissionsForUser(userId, model.VendorDomain, ownerId)
	if userPermissions == nil {
		return nil, NewServiceErrorf(http.StatusInternalServerError, "Could not find permissions for userId `%s`", userId)
	}

	roles := make([]model.RoleRestriction, 0)
	gamesCache := make(map[string]model.ResourceMeta)

	gamesCache["*"] = model.ResourceMeta{
		InternalName: "global",
		Preview:      "",
	}

	gamesCache["skip"] = model.ResourceMeta{
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
					Owner: ownerId,
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
					Owner: ownerId,
					Meta:  meta,
				},
			})
		}
	}

	return &model.UserRole{
		Email: user.Email,
		Name:  user.FullName,
		Roles: roles,
	}, nil
}

func (service *membershipService) GetUser(vendorId uuid.UUID, userId string) (*model.UserRole, error) {
	ownerId, err := GetOwnerForVendor(service.db.DB(), vendorId)
	if err != nil {
		return nil, err
	}

	return service.getUser(userId, ownerId)
}

func (service *membershipService) RemoveRoleToUserInGame(vendorId uuid.UUID, userId string, gameId string, role string) error {
	if exist, err := utils.CheckExists(service.db.DB(), &model.User{}, userId); !(exist && err == nil) {
		if err != nil {
			return NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get user by id `%s`", userId))
		}
		return NewServiceErrorf(http.StatusNotFound, "User `%s` not found", userId)
	}

	isGlobal := gameId == "" || gameId == "*"
	restrict := []string{"*"}

	if !isGlobal {
		if exist, err := utils.CheckExists(service.db.DB(), &model.Game{}, gameId); !(exist && err == nil) {
			if err != nil {
				return NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get game by id `%s`", gameId))
			}
			return NewServiceErrorf(http.StatusNotFound, "Game `%s` not found", gameId)
		}
		restrict = []string{gameId}
	}

	owner, err := GetOwnerForVendor(service.db.DB(), vendorId)
	if err != nil {
		return err
	}

	if service.enforcer.RemoveRole(rbac.Role{Role: role, User: userId, Owner: owner, Domain: model.VendorDomain, RestrictedResourceId: restrict}) == false {
		return NewServiceErrorf(http.StatusInternalServerError, "Could not remove role `%s` to user `%s`", role, userId)
	}

	return nil
}

func (service *membershipService) AddRoleToUserInGame(vendorId uuid.UUID, userId string, gameId string, role string) error {
	if exist, err := utils.CheckExists(service.db.DB(), &model.User{}, userId); !(exist && err == nil) {
		if err != nil {
			return NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get user by id `%s`", userId))
		}
		return NewServiceErrorf(http.StatusNotFound, "User `%s` not found", userId)
	}

	isGlobal := gameId == "" || gameId == "*"
	restrict := []string{"*"}

	if !isGlobal {
		if exist, err := utils.CheckExists(service.db.DB(), &model.Game{}, gameId); !(exist && err == nil) {
			if err != nil {
				return NewServiceError(http.StatusInternalServerError, errors.Wrapf(err, "Get game by id `%s`", gameId))
			}
			return NewServiceErrorf(http.StatusNotFound, "Game `%s` not found", gameId)
		}
		restrict = []string{gameId}
	}

	owner, err := GetOwnerForVendor(service.db.DB(), vendorId)
	if err != nil {
		return err
	}

	if service.enforcer.AddRole(rbac.Role{Role: role, User: userId, Owner: owner, Domain: model.VendorDomain, RestrictedResourceId: restrict}) == false {
		return NewServiceErrorf(http.StatusInternalServerError, "Could not add role `%s` to user `%s`", role, userId)
	}

	return nil
}

func (service *membershipService) SendInvite(vendorId uuid.UUID, userId string) error {
	return errors.New("Not implemented yet")
}

func (service *membershipService) AcceptInvite(inviteId uuid.UUID) error {
	return errors.New("Not implemented yet")
}

func (service *membershipService) GetUserPermissions(vendorId uuid.UUID, userId string) (*rbac.UserPermissions, error) {
	if exist, err := utils.CheckExists(service.db.DB(), &model.Vendor{}, vendorId); !(exist && err == nil) {
		if err != nil {
			return nil, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Check vendor exist"))
		}
		return nil, NewServiceError(http.StatusNotFound, "Vendor not found")
	}

	if exist, err := utils.CheckExists(service.db.DB(), &model.User{}, userId); !(exist && err == nil) {
		if err != nil {
			return nil, NewServiceError(http.StatusInternalServerError, errors.Wrap(err, "Check user exist"))
		}
		return nil, NewServiceError(http.StatusNotFound, "User not found")
	}

	return service.enforcer.GetPermissionsForUser(userId, "vendor", vendorId.String()), nil
}

func appendIfMissing(slice []string, users []string, skipNames []string) []string {
	for _, user := range users {
		exist := false
		skip := false
		for _, ele := range skipNames {
			if ele == user {
				skip = true
				break
			}
		}

		if skip {
			continue
		}

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
