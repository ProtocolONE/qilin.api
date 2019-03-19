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

func (p *UserService) Create(id string, lang string) (user model.User, err error) {
	user = model.User{
		ID:         id,
		Lang:       lang,
	}

	err = p.db.Create(&user).Error
	if err != nil {
		return user, errors.Wrap(err, "by insert new user")
	}

	return user, nil
}
