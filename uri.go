package graphql

import (
	"database/sql/driver"
	"fmt"
	"io"
)

// URI is a string representation of a URI.
// TODO: Turn into an actual URI.
type URI struct {
	raw string
}

func NewURI(raw string) URI {
	u := URI{}
	u.raw = raw
	return u
}

// String returns the value
func (u *URI) String() string {
	return u.raw
}

// Scan implements the driver.Scan interface
func (u *URI) Scan(v interface{}) error {
	return u.UnmarshalGQL(v)
}

// UnmarshalGQL implements the graphql.Marshaler interface
func (u *URI) UnmarshalGQL(v interface{}) error {
	if v == nil {
		u.raw = ""
		return nil
	}

	in, ok := v.(URI)
	if ok {
		u.raw = in.String()
		return nil
	}

	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("URI must be a string")
	}
	u.raw = str

	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (u URI) MarshalGQL(w io.Writer) {
	fmt.Fprintf(w, `"%s"`, u.String())
}

// Value implements the driver.Value interface
func (u URI) Value() (driver.Value, error) {
	return u.raw, nil
}

func (u URI) MarshalJSON() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *URI) UnmarshalJSON(value []byte) error {
	u.raw = string(value)
	return nil
}
