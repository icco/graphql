package graphql

// TODO: Turn into an actual URI.
type URI struct {
	value string
}

func (u *URI) String() string {
	return u.value
}
