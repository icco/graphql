package hexdate

import (
	"fmt"
	"time"
)

var Root = time.Date(1988, time.February, 22, 0, 0, 0, 0, time.UTC)

type Date struct {
	Days int64
}

func Now() *Date {
	delta := time.Since(Root)

	return &Date{Days: int64(delta.Hours()) / 24}
}

func (d *Date) String() string {
	return fmt.Sprintf("%X", d.Days)
}
