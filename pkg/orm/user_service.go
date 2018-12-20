package orm

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
)

type UserService struct {
	db 		*gorm.DB
	jwt_signMethod	jwt.SigningMethod
	jwt_signSecret	[]byte
}

func NewUserService(db *Database, jwtConf *conf.Jwt) (*UserService, error) {
	return &UserService{db.database,
		jwt.GetSigningMethod(jwtConf.Algorithm),
		jwtConf.SignatureSecret }, nil
}

func (p *UserService) CreateUser(u *model.User) error {
	return p.db.Create(u).Error
}

func (p *UserService) UpdateUser(u *model.User) error {
	return p.db.Update(u).Error
}

func (p *UserService) FindByID(id int) (user model.User, err error) {
	err = p.db.First(&user, model.User{ID: uint(id)}).Error
	return
}

func (p *UserService) FindByLoginAndPass(login, pass string) (user model.User, err error) {
	err = p.db.First(&user, "login = ? and password = ?", login, pass).Error
	return
}

func (p *UserService) Login(login, pass string) (result model.LoginResult, err error) {

	user := model.User{}

	user, err = p.FindByLoginAndPass(login, pass)
	if err != nil {
		return result, err
	}

	token := jwt.NewWithClaims(p.jwt_signMethod, jwt.MapClaims{
		"user_id": user.ID,
	})

	result.AccessToken, err = token.SignedString(p.jwt_signSecret)
	if err != nil {
		return result, err
	}

	result.User.Id = user.ID
	result.User.Nickname = user.Nickname
	//result.User.Avatar = user.Avatar

	return result, nil
}
