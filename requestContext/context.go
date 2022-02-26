package requestContext

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/pelletier/go-toml"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/types"
	"gopkg.in/yaml.v2"
)

type Context struct {
	L               logger.AppLogger
	DB              types.Storage
	StructValidater func(interface{}) error
}

// Deprecated.
// The useful methods here should be returned into structs
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

func (c *Context) NewReqContext(rw http.ResponseWriter, r *http.Request) ReqContext {
	return NewReqContext(c, r, rw)
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
func (rc ReqContext) WriteError(msg string, errCode ErrorCodes, details ...interface{}) {
	_err := WriteError(msg, errCode, rc.Req, rc.Rw, details...)
	if _err != nil {
		rc.L.Error().Err(_err).Msg("Failure in WriteErr")
	}
}
func (rc ReqContext) WriteNotFound(errCode ErrorCodes) {
	_err := WriteErr(errors.New("Not found"), errCode, rc.Req, rc.Rw)
	if _err != nil {
		rc.L.Error().Err(_err).Msg("Failure in WriteErr")
	}
}
func (rc ReqContext) WriteErr(err error, errCode ErrorCodes) {
	if apiErr, ok := err.(APIError); ok {
		code := apiErr.Err.Code
		if errCode != "" {
			code = errCode + ": " + code
		}

		_err := WriteError(apiErr.Err.Message, apiErr.Err.Code, rc.Req, rc.Rw, apiErr.Details)
		if _err != nil {
			rc.L.Error().Err(_err).Msg("Failure in WriteErr")
		}
		return
	}
	_err := WriteErr(err, errCode, rc.Req, rc.Rw)
	if _err != nil {
		rc.L.Error().Err(_err).Msg("Failure in WriteErr")
	}
}
func (rc ReqContext) WriteOutput(output interface{}, statusCode int) {
	_err := WriteOutput(false, statusCode, output, rc.Req, rc.Rw)
	if _err != nil {
		rc.L.Error().Err(_err).Msg("Failure in WriteErr")
	}
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

type decoder interface {
	Decode(v interface{}) error
}

func (rc ReqContext) GetDecoder() decoder {
	switch rc.ContentKind {
	case OutputJson:
		return json.NewDecoder(rc.Req.Body)
	case OutputToml:
		return toml.NewDecoder(rc.Req.Body)
	case OutputYaml:
		return yaml.NewDecoder(rc.Req.Body)
	}
	return json.NewDecoder(rc.Req.Body)
}

// Reads the requests body, and validates it.
// with writeErr = true, upon validation error it will write the error to the body. In this case, the caller should simply return
func (rc ReqContext) ValidateBody(j interface{}, writeErr bool) error {
	if rc.Req.ContentLength == 0 {
		err := ErrEmptyBody
		if writeErr {
			rc.WriteErr(err, CodeErrMissingBody)
		}
		return err
	}
	decoder := rc.GetDecoder()
	err := decoder.Decode(j)
	if err != nil {
		if writeErr {
			rc.WriteErr(err, CodeErrMarshal)
		}
		return err
	}
	err = rc.ValidateStruct(j)
	if err != nil {
		if writeErr {
			rc.WriteErr(err, CodeErrInputValidation)
		}
		return err
	}
	return err
}

// Will perform validation and write errors to responsewriter if validation failed.
// If err is non-nill, the caller should simply return
func (rc ReqContext) ValidateBytes(body []byte, j interface{}) error {
	err := rc.Unmarshal(body, j)
	if err != nil {
		rc.WriteErr(err, CodeErrMarshal)
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
