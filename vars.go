package graphql

import "go.icco.me/gutil/logging"

const (
	// AppName is the name of the service in GCP.
	AppName = "graphql"
)

var (
	log = logging.Must(logging.NewLogger(AppName))
)
