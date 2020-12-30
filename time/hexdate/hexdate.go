package hexdate

import (
	"fmt"
	"time"
)

const Root = time.Date(1988, time.February, 22, 0, 0, 0, 0, time.UTC)

type Date string

func Now() *Date {
	delta := time.Since(Root)
	days := delta.Hours() / 24

	return &Date{fmt.Sprintf("%X", days)}
}

func (d *Date) String() string {
	return string(d)
}
