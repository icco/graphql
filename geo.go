package graphql

import (
	"github.com/paulmach/orb"
)

// Geo is a simple type for wrapping a point.
type Geo struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

func (g *Geo) ToOrb() orb.Point {
	return orb.Point{g.Long, g.Lat}
}
