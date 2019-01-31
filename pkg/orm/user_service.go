package orm

import (
	"bytes"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"html/template"
	"net/http"
	"path"
	"qilin-api/pkg/conf"
	"qilin-api/pkg/model"
	"qilin-api/pkg/sys"
	"runtime"
)

type UserService struct {
	db             *gorm.DB
	jwt_signMethod jwt.SigningMethod
	jwt_signSecret []byte
	templates      *template.Template
	mailer         sys.Mailer
	langMap        sys.LangMap
}

func NewUserService(db *Database, jwtConf *conf.Jwt, mailer sys.Mailer) (*UserService, error) {

	_, moduleFile, _, _ := runtime.Caller(0)
	rootProj := path.Dir(moduleFile) + "/../.."

	langMap, err := sys.NewLangMap(rootProj + "/locale/*.json")
	if err != nil {
		return nil, errors.Wrap(err, "loading lang files")
	}

	templates, err := template.New("").
		Funcs(langMap.GetTemplFunc()).
		ParseGlob(rootProj + "/templates/*.gohtml")
	if err != nil {
		return nil, errors.Wrap(err, "loading templates")
	}

	return &UserService{db.database,
		jwt.GetSigningMethod(jwtConf.Algorithm),
		jwtConf.SignatureSecret,
		templates,
		mailer,
		langMap}, nil
}

func (p *UserService) UpdateUser(u *model.User) error {
	return p.db.Update(u).Error
}

func (p *UserService) FindByID(id *uuid.UUID) (user model.User, err error) {
	err = p.db.First(&user, model.User{ID: *id}).Error
	if err == gorm.ErrRecordNotFound {
		return user, NewServiceError(http.StatusNotFound, "User not found")
	} else if err != nil {
		return user, errors.Wrap(err, "search user by id")
	}
	return
}

func (p *UserService) Login(login, pass string) (result model.LoginResult, err error) {

	user := model.User{}

	err = p.db.First(&user, "login = ? and password = ?", login, pass).Error
	if err == gorm.ErrRecordNotFound {
		return result, NewServiceError(http.StatusNotFound, "User not found")
	} else if err != nil {
		return result, errors.Wrap(err, "when searching user by login and passwd")
	}

	token := jwt.NewWithClaims(p.jwt_signMethod, jwt.MapClaims{
		"id": user.ID.Bytes(),
	})

	result.AccessToken, err = token.SignedString(p.jwt_signSecret)
	if err != nil {
		return result, errors.Wrap(err, "when signing token")
	}

	result.User.Id = user.ID
	result.User.Nickname = user.Nickname
	result.User.Lang = user.Lang
	//result.User.Avatar = user.Avatar

	return result, nil
}

func (p *UserService) Register(login, pass, lang string) (userId uuid.UUID, err error) {
	foundUsr := 0
	err = p.db.Model(&model.User{}).Where("login = ?", login).Count(&foundUsr).Error
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "while check user login")
	}
	if foundUsr > 0 {
		return uuid.Nil, NewServiceError(http.StatusConflict, "User already registered")
	}

	user := model.User{
		ID: uuid.NewV4(),
		Login: login,
		Password: pass,
		Nickname: login,
		Lang: lang,
	}

	err = p.db.Create(&user).Error
	if err != nil {
		return uuid.Nil, errors.Wrap(err, "by insert new user")
	}

	return user.ID, nil
}

type templ_ResetPasswd struct {
	User		*model.User
	ResetURL	string
}

func (p *UserService) ResetPassw(email string) (err error) {
	user := model.User{}

	err = p.db.First(&user, "login = ?", email).Error
	if err != nil {
		return NewServiceError(http.StatusNotFound, "User not found")
	}

	body := bytes.Buffer{}
	err = p.templates.ExecuteTemplate(&body, "reset-passwd.gohtml", templ_ResetPasswd{&user, "http://localhost/"})
	if err != nil {
		return errors.Wrap(err, "when rendering template in reset password")
	}
	subject := p.langMap.Locale(user.Lang, "reset-password")

	err = p.mailer.Send(user.Login, subject, body.String())

	return nil
}