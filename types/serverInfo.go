package types

import "time"

type ServerInfo struct {
	// When the server was started
	ServerStartedAt time.Time `json:"server_started_at"`
	// Short githash for current commit
	GitHash string `json:"git_hash"`
	// Version-number for commit
	Version string `json:"version"`
	// Date of build
	BuildDate time.Time `json:"build_date"`

	// Size of database.
	DatabaseSize    int64  `json:"database_size"`
	DatabaseSizeStr string `json:"database_size_str"`
}

// Server info
// swagger:response
type serverInfo struct {
	// in:body
	Body []ServerInfo
}
