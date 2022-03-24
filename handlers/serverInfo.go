package handlers

import (
	"net/http"

	"github.com/dustin/go-humanize"
	"github.com/runar-rkmedia/skiver/requestContext"
	"github.com/runar-rkmedia/skiver/types"
)

type Sizer interface {
	Size() (int64, error)
}

func GetServerInfo(sizer Sizer, serverInfo func() *types.ServerInfo) AppHandler {
	return func(rc requestContext.ReqContext, rw http.ResponseWriter, r *http.Request) (interface{}, error) {
		size, sizeErr := sizer.Size()
		info := serverInfo()
		if sizeErr != nil {
			rc.L.Warn().Err(sizeErr).Msg("Failed to retrieve size of database")
		} else {
			info.DatabaseSize = size
			info.DatabaseSizeStr = humanize.Bytes(uint64(size))
		}
		return info, nil
	}
}
