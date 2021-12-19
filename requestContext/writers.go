package requestContext

import (
	"net/http"
	"strings"

	"github.com/runar-rkmedia/skiver/models"
)

type OutputKind = int

const (
	OutputJson OutputKind = iota + 1
	OutputYaml
	OutputToml
)

func WriteAuto(output interface{}, err error, errCode ErrorCodes, r *http.Request, rw http.ResponseWriter) error {
	if err != nil {
		return WriteError(err.Error(), errCode, r, rw)
	}

	return WriteOutput(false, http.StatusOK, output, r, rw)
}
func WriteErr(err error, code ErrorCodes, r *http.Request, rw http.ResponseWriter) error {
	return WriteError(err.Error(), code, r, rw, err)
}
func WriteError(msg string, code ErrorCodes, r *http.Request, rw http.ResponseWriter, details ...interface{}) error {
	ae := models.APIError{Error: msg, Code: string(code)}
	if details != nil && details[0] != nil {
		ae.Details = details[0]
	}
	statusCode := http.StatusBadGateway
	switch code {
	case CodeErrMethodNotAllowed:
		statusCode = http.StatusMethodNotAllowed
	case CodeErrRequestEntityTooLarge:
		statusCode = http.StatusRequestEntityTooLarge
		// duplicates??
	case CodeErrNotFoundLocale, CodeErrNoRoute:
		statusCode = http.StatusNotFound
	case CodeErrReadBody, CodeErrDBCreateLocale:
		statusCode = http.StatusBadGateway
	case CodeErrUnmarshal, CodeErrMarhal, CodeErrJmesPath, CodeErrJmesPathMarshal, CodeErrInputValidation, CodeErrIDNonValid, CodeErrIDTooLong, CodeErrIDEmpty:
		statusCode = http.StatusBadRequest
	}
	return WriteOutput(true, statusCode, ae, r, rw)

}

// attempts to guess at what kind of output the user most likely wants
func WantedOutputFormat(r *http.Request) OutputKind {
	if o := contentType(r.Header.Get("Accept")); o > 0 {
		return o
	}
	// If the content-type is set, the user probably wants the same type.
	if o := contentType(r.Header.Get("Content-Type")); o > 0 {
		return o
	}
	q := r.URL.Query().Get("format")
	if o := contentType(q); o > 0 {
		return o
	}

	// Fallback to a readable format.
	return OutputToml
}

func contentType(kind string) OutputKind {
	switch {
	case strings.Contains(kind, "json"):
		return OutputJson
	case strings.Contains(kind, "yaml"):
		return OutputYaml
	case strings.Contains(kind, "toml"):
		return OutputToml
	}
	return 0
}
