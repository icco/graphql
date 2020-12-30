package neralie

import (
	"fmt"
	"time"
)

type Time struct {
	Beat  int
	Pulse int
}

func Now() *Time {
	return FromTime(time.Now())
}

func FromTime(t time.Time) *Time {
	utc := t.UTC()
	secToday := utc.Hour()*60*60 + utc.Minute()*60 + utc.Second()
	root := int((float64(secToday) / 86400.0) * 1000000.0)
	return &Time{
		Beat:  root / 1000,
		Pulse: root % 1000,
	}
}

func (t *Time) String() string {
	return fmt.Sprintf("%03d:%03d", t.Beat, t.Pulse)
}
