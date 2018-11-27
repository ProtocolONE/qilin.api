package mongo

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"qilin-api/pkg"
	"time"
)

// GameService is service to interact with database and Game object.
type GameService struct {
	collection *mgo.Collection
}

// NewGameService initialize this service.
func NewGameService(session *Session) (*GameService, error) {
	collection := session.GetCollection("game")
	if err := collection.EnsureIndex(gameModelIndex()); err != nil {
		return nil, err
	}

	return &GameService{collection}, nil
}

// CreateGame creates new Game object in database
func (p *GameService) CreateGame(u *qilin.Game) error {
	game, err := newGameModel(u)
	if err != nil {
		return err
	}

	game.CreatedAt = time.Now()

	err = p.collection.Insert(&game)
	if err != nil {
		return err
	}

	u.ID = game.ID.Hex()
	u.CreatedAt = game.CreatedAt

	return nil
}

func (p *GameService) UpdateGame(u *qilin.Game) error {
	game, err := newGameModel(u)
	if err != nil {
		return err
	}

	game.UpdatedAt = time.Now()
	return p.collection.UpdateId(game.ID, &game)
}

// FindByID return Game object by given id
func (p *GameService) FindByID(id string) (*qilin.Game, error) {
	if !bson.IsObjectIdHex(id) {
		return nil, fmt.Errorf("Given `%s` is not the ObjectId Hex", id)
	}

	gameObj := game{}

	err := p.collection.FindId(bson.ObjectIdHex(id)).One(&gameObj)
	if err != nil {
		return nil, err
	}

	return gameObj.toQilinGame(), nil
}

func (p *GameService) GetAll() ([]*qilin.Game, error) {
	var games []game
	if err := p.collection.Find(nil).All(&games); err != nil {
		return nil, err
	}
	return p.mapGames(games), nil
}

func (p *GameService) FindByName(name string) ([]*qilin.Game, error) {
	var games []game
	if err := p.collection.Find(bson.M{"name": name}).All(&games); err != nil {
		return nil, err
	}
	return p.mapGames(games), nil
}

func (p *GameService) mapGames(games []game) []*qilin.Game {
	vsm := make([]*qilin.Game, len(games))
	for i, v := range games {
		vsm[i] = v.toQilinGame()
	}
	return vsm
}
