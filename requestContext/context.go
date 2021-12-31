package requestContext

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/runar-rkmedia/gabyoall/logger"
	"github.com/runar-rkmedia/skiver/types"
)

type Context struct {
	L               logger.AppLogger
	DB              types.Storage
	StructValidater func(interface{}) error
}
type ReqContext struct {
	Context     *Context
	Req         *http.Request
	L           logger.AppLogger
	Rw          http.ResponseWriter
	ContentKind OutputKind
	Accept      OutputKind
	RemoteIP    string
}

func NewReqContext(context *Context, req *http.Request, rw http.ResponseWriter) ReqContext {
	// TODO: parse this value into a ip. For now, we do not actually need it.
	// (we only use the ip for reducing session-duplications if there are lots of logins.)
	remoteIP := req.Header.Get("Forwarded")
	if remoteIP == "" {
		remoteIP = req.Header.Get("X-Forwarded-For")
	}
	if remoteIP == "" {
		remoteIP = req.Header.Get("X-Originating-IP")
	}
	if remoteIP == "" {
		remoteIP = req.RemoteAddr
	}
	h := make(http.Header)
	for k, v := range req.Header {
		switch strings.ToLower(k) {
		case "cookie", "authorization":
			continue
		}
		for i := 0; i < len(v); i++ {

			h.Add(k, v[i])
		}

	}
	return ReqContext{
		Context:     context,
		L:           logger.With(context.L.With().Str("method", req.Method).Str("path", req.URL.Path).Interface("headers", h).Logger()),
		Req:         req,
		Rw:          rw,
		ContentKind: contentType(req.Header.Get("Content-Type")),
		Accept:      WantedOutputFormat(req),
		RemoteIP:    remoteIP,
	}
}

func (rc ReqContext) WriteAuto(output interface{}, error error, errCode ErrorCodes) {
	err := WriteAuto(output, error, errCode, rc.Req, rc.Rw)
	if err != nil {
		l := rc.L.Error().
			Err(err).
			Str("path", rc.Req.URL.String()).
			Str("method", rc.Req.Method)
		if error != nil {
			l = l.
				Str("for-error-code", string(errCode)).
				Str("for-error", error.Error())
		}
		l.Msg("Failure during WriteAuto")
	}
}
func (rc ReqContext) WriteError(msg string, errCode ErrorCodes) {
	WriteError(msg, errCode, rc.Req, rc.Rw)
}
func (rc ReqContext) WriteErr(err error, errCode ErrorCodes) {
	WriteErr(err, errCode, rc.Req, rc.Rw)
}
func (rc ReqContext) WriteOutput(output interface{}, statusCode int) {
	WriteOutput(false, statusCode, output, rc.Req, rc.Rw)
}
func (rc ReqContext) ValidateStruct(input interface{}) error {
	return rc.Context.StructValidater(input)
}
func (rc ReqContext) Unmarshal(body []byte, j interface{}) error {
	if body == nil {
		if rc.L.HasDebug() {
			rc.L.Debug().Msg("Body was nil")
		}
		return fmt.Errorf("Body was nil")
	}
	err := UnmarshalWithKind(rc.ContentKind, body, j)
	if err != nil && rc.L.HasDebug() {
		rc.L.Debug().
			Bytes("body", body).
			Err(err).
			Msg("unmarshalling failed with input")
	}
	return err
}

// Will perform validation and write errors to responsewriter if validation failed.
// If err is non-nill, the caller should simply return
func (rc ReqContext) ValidateBytes(body []byte, j interface{}) error {
	err := rc.Unmarshal(body, j)
	if err != nil {
		rc.WriteErr(err, CodeErrMarhal)
		return err
	}
	err = rc.ValidateStruct(j)
	if err != nil {
		// rw.Header().Set("Content-Type", "application/json")
		// rw.WriteHeader(http.StatusBadRequest)
		rc.WriteErr(err, CodeErrInputValidation)
		return err
	}
	return err
}
