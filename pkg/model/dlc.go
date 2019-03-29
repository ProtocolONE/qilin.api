package model

import "github.com/satori/go.uuid"

type (
	Dlc struct {
		Model
		Image           string
		GameID          uuid.UUID
		Game            Game
		Product         ProductEntry `gorm:"polymorphic:Entry;"`
	}
)

func (p *Dlc) GetID() uuid.UUID {
	return p.Model.ID
}

func (p *Dlc) GetName() string {
	return p.Game.InternalName
}

func (p *Dlc) GetType() ProductType {
	return ProductDLC
}

func (p *Dlc) GetImage(lang string) string {
	return p.Image
}