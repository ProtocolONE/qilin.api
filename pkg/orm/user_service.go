package orm

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"log"
	"qilin-api/pkg/api"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
)

var (
	ErrLoginAlreadyTaken = errors.New("Login already taken")
	ErrUserNotFound = errors.New("User not found")
	ErrSystemError = errors.New("System error")
)

type UserService struct {
	db 				*gorm.DB
	jwt_signMethod	jwt.SigningMethod
	jwt_signSecret	[]byte
	mailer			api.Mailer
}

func NewUserService(db *Database, jwtConf *conf.Jwt, mailer api.Mailer) (*UserService, error) {
	return &UserService{db.database,
		jwt.GetSigningMethod(jwtConf.Algorithm),
		jwtConf.SignatureSecret,
		mailer, }, nil
}

func (p *UserService) UpdateUser(u *model.User) error {
	return p.db.Update(u).Error
}

func (p *UserService) FindByID(id uuid.UUID) (user model.User, err error) {
	err = p.db.First(&user, model.User{ID: id}).Error
	return
}

func (p *UserService) FindByLoginAndPass(login, pass string) (user model.User, err error) {
	err = p.db.First(&user, "login = ? and password = ?", login, pass).Error
	if err == gorm.ErrRecordNotFound {
		err = ErrUserNotFound
	} else {
		log.Println(err)
		err = ErrSystemError
	}
	return
}

func (p *UserService) Login(login, pass string) (result model.LoginResult, err error) {

	user := model.User{}

	user, err = p.FindByLoginAndPass(login, pass)
	if err != nil {
		return result, err
	}

	token := jwt.NewWithClaims(p.jwt_signMethod, jwt.MapClaims{
		"id": user.ID.Bytes(),
	})

	result.AccessToken, err = token.SignedString(p.jwt_signSecret)
	if err != nil {
		log.Println(err)
		return result, ErrSystemError
	}

	result.User.Id = user.ID
	result.User.Nickname = user.Nickname
	//result.User.Avatar = user.Avatar

	return result, nil
}

func (p *UserService) Register(login, pass string) (userId uuid.UUID, err error) {

	user := model.User{}

	err = p.db.First(&user, "login = ?", login).Error
	if err == nil {
		return uuid.Nil, ErrLoginAlreadyTaken
	}

	user.Login = login
	user.Password = pass
	user.Nickname = login
	user.ID = uuid.NewV4()

	err = p.db.Create(&user).Error
	if err != nil {
		log.Println(err)
		return uuid.Nil, ErrSystemError
	}

	return user.ID, nil
}

func (p *UserService) ResetPassw(email string) (err error) {
	user := model.User{}

	err = p.db.First(&user, "login = ?", email).Error
	if err == nil {
		return ErrUserNotFound
	}

	err = p.mailer.Send(user.Login, "subject", "body")

	return nil
}