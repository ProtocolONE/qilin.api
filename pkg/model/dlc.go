package model

import (
	"github.com/satori/go.uuid"
	"qilin-api/pkg/model/utils"
)

type (
	Dlc struct {
		Model
		Image 			JSONB 			`gorm:"type:jsonb"`
		GameID          uuid.UUID
		Game            Game
		Product         ProductEntry 	`gorm:"polymorphic:Entry;"`
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

func (p *Dlc) GetImage() (res *utils.LocalizedString) {
	res = &utils.LocalizedString{}
	if p.Image == nil {
		return
	}
	_ = p.Image.Scan(res)
	return
}