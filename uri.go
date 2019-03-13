package graphql

import (
	"fmt"
	"io"
)

// TODO: Turn into an actual URI.
type URI struct {
	value string
}

func (u *URI) String() string {
	return u.value
}

func (u *URI) Scan(v interface{}) error {
	return u.UnmarshalGQL(v)
}

func (u *URI) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("URI must be strings")
	}
	u.value = str

	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (u URI) MarshalGQL(w io.Writer) {
	fmt.Fprintf(w, `%s`, u.String())
}
