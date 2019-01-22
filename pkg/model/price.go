package model

import "github.com/jinzhu/gorm"

// Price is a base object to represent Game SKU prices for different
// currency in regions. It`s fixed object because game in the system should
// have prices in all local currency after release.
type Price struct {
	gorm.Model
	USD float32 `json:"usd"`
	RUR float32 `json:"rur"`
	EUR float32 `json:"eur"`
}
