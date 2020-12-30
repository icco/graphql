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
	now := time.Now().UTC()
	secToday := now.Hour()*60*60 + now.Minute()*60 + now.Second()
	root := secToday / 86400
	return &Time{
		Beat:  root / 1000,
		Pulse: root % 1000,
	}
}

func (t *Time) String() string {
	return fmt.Sprintf("%03d:%03d", t.Beat, t.Pulse)
}
