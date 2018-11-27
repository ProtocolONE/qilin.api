package qilin

// GamePrice is a base object to represent Game SKU prices for different
// currency in regions. It`s fixed object because game in the system should
// have prices in all local currency after release.
type GamePrice struct {
	USD float32
	RUR float32
	EUR float32
}
