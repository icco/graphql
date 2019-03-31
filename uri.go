package graphql

import (
	"database/sql/driver"
	"fmt"
	"io"
)

// URI is a string representation of a URI.
// TODO: Turn into an actual URI.
type URI struct {
	Raw string
}

func NewURI(raw string) URI {
	u := URI{}
	u.Raw = raw
	return u
}

// String returns the value
func (u *URI) String() string {
	return u.Raw
}

// Scan implements the driver.Scan interface
func (u *URI) Scan(v interface{}) error {
	return u.UnmarshalGQL(v)
}

// UnmarshalGQL implements the graphql.Marshaler interface
func (u *URI) UnmarshalGQL(v interface{}) error {
	if v == nil {
		u.Raw = ""
		return nil
	}

	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("URI must be strings")
	}
	u.Raw = str

	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (u URI) MarshalGQL(w io.Writer) {
	fmt.Fprintf(w, `"%s"`, u.String())
}

// Value implements the driver.Value interface
func (u URI) Value() (driver.Value, error) {
	return u.Raw, nil
}
