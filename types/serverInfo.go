package types

import "time"

type ServerInfo struct {
	// When the server was started
	ServerStartedAt time.Time
	// Short githash for current commit
	GitHash string
	// Version-number for commit
	Version string
	// Date of build
	BuildDate time.Time

	// Size of database.
	DatabaseSize    int64
	DatabaseSizeStr string
}

// Server info
// swagger:response
type serverInfo struct {
	// in:body
	Body []ServerInfo
}
