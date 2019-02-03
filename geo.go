package graphql

import (
	"database/sql/driver"
	"fmt"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkb"
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

// GeoFromOrb creates a Geo from github.com/paulmach/orb.Point.
func GeoFromOrb(p *orb.Point) *Geo {
	if p == nil {
		return nil
	}

	return &Geo{
		Long: p[0],
		Lat:  p[1],
	}
}

// GeoScanner is used for unmarshaling data from a database row.
func GeoScanner(g interface{}) *wkb.GeometryScanner {
	return wkb.Scanner(g)
}

// GeoConvertValue is used for marshaling data to a database.
func GeoConvertValue(v interface{}) (driver.Value, error) {
	g, ok := v.(*Geo)
	if !ok {
		return driver.Value(nil), fmt.Errorf("is not a Geo")
	}

	if g == nil {
		return driver.Value(nil), nil
	}

	return wkb.Value(g.ToOrb()), nil
}
