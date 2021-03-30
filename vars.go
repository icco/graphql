package graphql

import "github.com/icco/gutil/logging"

const (
	// AppName is the name of the service in GCP.
	AppName = "graphql"
)

var (
	log = logging.Must(logging.NewLogger(AppName))
)
