package graphql

import (
	"database/sql/driver"
	"fmt"

	"github.com/paulmach/orb"
)

// Geo is a simple type for wrapping a point. Units are in Degrees.
type Geo struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

// ToOrb translates a Geo point into one that has lots of useful functions on
// it.
func (g *Geo) ToOrb() *orb.Point {
	return &orb.Point{g.Long, g.Lat}
}

func (g *Geo) ToDatabaseString() string {
	return fmt.Sprintf("POINT(%f %f)", g.Lat, g.Long)
}

func GeoConvertValue(v interface{}) (driver.Value, error) {
	g, ok := v.(*Geo)
	if !ok {
		return driver.Value(""), fmt.Errorf("is not a Geo")
	}

	if g == nil {
		return driver.Value(nil), nil
	}

	return driver.Value(g.ToDatabaseString()), nil
}
