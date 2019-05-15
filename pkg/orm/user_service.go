package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"html/template"
	"net/http"
	"path"
	"qilin-api/pkg/model"
	"qilin-api/pkg/sys"
	"runtime"
	"time"
)

type UserService struct {
	db        *gorm.DB
	templates *template.Template
	mailer    sys.Mailer
	langMap   sys.LangMap
}

func NewUserService(db *Database, mailer sys.Mailer) (*UserService, error) {
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
		templates,
		mailer,
		langMap}, nil
}

func (p *UserService) FindByID(id string) (user model.User, err error) {
	err = p.db.First(&user, model.User{ID: id}).Error
	if err == gorm.ErrRecordNotFound {
		return user, NewServiceError(http.StatusNotFound, "User not found")
	} else if err != nil {
		return user, errors.Wrap(err, "search user by id")
	}
	return
}

func (p *UserService) UpdateLastSeen(user model.User) error {
	now := time.Now().UTC()
	user.LastSeen = &now
	if err := p.db.Save(&user).Error; err != nil {
		return NewServiceError(http.StatusInternalServerError, err)
	}
	return nil
}

func (p *UserService) Create(id string, email string, lang string) (user model.User, err error) {
	user = model.User{
		ID:    id,
		Email: email,
		Lang:  lang,
	}

	err = p.db.Save(&user).Error
	if err != nil {
		return user, errors.Wrap(err, "by save new user")
	}

	return user, nil
}
