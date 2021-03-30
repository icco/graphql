package graphql

import (
	"database/sql/driver"
	"fmt"
	"io"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// Duration is a float64 representation of a Duration.
type Duration struct {
	raw float64
}

// NewDuration creates a new Duration.
func NewDuration(raw float64) Duration {
	d := Duration{}
	d.raw = raw
	return d
}

// ParseDurationFromString takes a duration string and turns it into a
// duration. A duration string is a possibly signed sequence of decimal
// numbers, each with optional fraction and a unit suffix, such as "300ms",
// "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s",
// "m", "h".
func ParseDurationFromString(dur string) Duration {
	i, err := time.ParseDuration(dur)
	if err != nil {
		log.Errorw("could not parse duration", zap.Error(err))
		return NewDuration(0)
	}

	return NewDuration(i.Seconds())
}

// float64 returns the value
func (d *Duration) float64() float64 {
	return d.raw
}

// Scan implements the driver.Scan interface
func (d *Duration) Scan(v interface{}) error {
	return d.UnmarshalGQL(v)
}

// UnmarshalGQL implements the graphql.Marshaler interface
func (d *Duration) UnmarshalGQL(v interface{}) error {
	if v == nil {
		d.raw = 0
		return nil
	}

	in, ok := v.(Duration)
	if ok {
		d.raw = in.float64()
		return nil
	}

	f, ok := v.(float64)
	if !ok {
		return fmt.Errorf("Duration must be a float64")
	}
	d.raw = f

	return nil
}

// MarshalGQL implements the graphql.Marshaler interface.
func (d Duration) MarshalGQL(w io.Writer) {
	fmt.Fprintf(w, `%f`, d.float64())
}

// Value implements the driver.Value interface.
func (d Duration) Value() (driver.Value, error) {
	return d.raw, nil
}

// MarshalJSON turns a duration to a string.
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`%f`, d.float64())), nil
}

// UnmarshalJSON turns a string to a Duration.
func (d *Duration) UnmarshalJSON(value []byte) error {
	f, err := strconv.ParseFloat(string(value), 64)
	if err != nil {
		return err
	}

	d.raw = f
	return nil
}
