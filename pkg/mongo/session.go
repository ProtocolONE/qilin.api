package mongo

import (
	"gopkg.in/mgo.v2"
	"qilin-api/pkg/conf"
)

type Session struct {
	session  *mgo.Session
	database *mgo.Database
}

func NewSession(config *conf.Database) (*Session, error) {
	//var err error
	session, err := mgo.Dial(config.Host)
	if err != nil {
		return nil, err
	}
	session.SetMode(mgo.Monotonic, true)
	database := session.DB(config.Database)

	return &Session{session, database}, err
}

func (s *Session) GetCollection(col string) *mgo.Collection {
	return s.database.C(col)
}

func (s *Session) Copy() *mgo.Session {
	return s.session.Copy()
}

func (s *Session) Close() {
	if s.session != nil {
		s.session.Close()
	}
}

func (s *Session) DropDatabase() error {
	if s.session != nil {
		return s.database.DropDatabase()
	}
	return nil
}
