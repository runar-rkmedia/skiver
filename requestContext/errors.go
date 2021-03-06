package requestContext

import (
	"fmt"
)

type ErrorCodes string
type Error struct {
	Code    ErrorCodes `json:"code,omitempty"`
	Message string     `json:"error,omitempty"`
}
type APIError struct {
	Details       interface{} `json:"details,omitempty"`
	Err           Error       `json:"error"`
	InternalError error       `json:"-"`
	StatusCode    int         `json:"-"`
}

// swagger:response
type apiError struct {
	// in:body
	Body APIError
}

func (e APIError) Error() string {
	if e.Err.Message == "" && e.InternalError != nil {
		e.Err.Message = e.InternalError.Error()
	}
	return e.Err.Error()
}
func (e Error) Error() string {
	return e.Message
}
func (e APIError) ErrCode() string {
	return fmt.Sprintf("%s", e.Err.Code)
}
func (e APIError) ErrHttpStatus() int {
	return e.StatusCode
}

func NewError(msg string, code ErrorCodes) Error {
	return Error{code, msg}
}

var (
	ErrEmptyBody  = NewError("Body was empty", CodeErrInputValidation)
	ErrIDNonValid = NewError("Id was not valid", CodeErrIDNonValid)
	ErrIDTooLong  = NewError("Id was too long", CodeErrIDTooLong)
	ErrIDEmpty    = NewError("Id was empty", CodeErrIDEmpty)
)

const (
	CodeErrMethodNotAllowed ErrorCodes = "Error: HTTP-Method is not allowed"
	CodeErrNotImplemented   ErrorCodes = "Error: Not implemented"
	CodeErrNoRoute          ErrorCodes = "Error: No route matched for this http-path"
	CodeErrReadBody         ErrorCodes = "Error: Failed to read body"
	CodeErrMarshal          ErrorCodes = "Error: Failed to marshal"
	CodeErrUnmarshal        ErrorCodes = "Error: Failed to unmarshal"
	CodeErrJmesPath         ErrorCodes = "Error: JmesPath"
	CodeErrJmesPathMarshal  ErrorCodes = "Error: JmesPathMarshal"
	CodeErrPasswordHashing  ErrorCodes = "Error: Password-creation"
	CodeErrUnknown          ErrorCodes = "Error: Unknown"
	CodeErrTemplating       ErrorCodes = "Error: Templating"

	CodeErrRequestEntityTooLarge ErrorCodes = "Error: Request Entity too large"
	CodeErrInputValidation       ErrorCodes = "Error: General input validation"
	CodeErrMissingBody           ErrorCodes = "Error: No body"
	CodeErrIDNonValid            ErrorCodes = "Error: ID not valid"
	CodeErrIDTooLong             ErrorCodes = "Error: ID is too long"
	CodeErrIDEmpty               ErrorCodes = "Error: ID was Empty"

	CodeErrAuthenticationRequired ErrorCodes = "Error: Authentication required"
	CodeErrAuthoriziationFailed   ErrorCodes = "Error: Authorization failed"
	CodeErrAuthoriziation         ErrorCodes = "Error: Authorization missing"

	CodeErrLocale               ErrorCodes = "Error: Locale error"
	CodeErrUser                 ErrorCodes = "Error: User error"
	CodeErrProject              ErrorCodes = "Error: Project error"
	CodeErrSnapshot             ErrorCodes = "Error: Snapshot error"
	CodeErrReportMissing        ErrorCodes = "Error: Report missing"
	CodeErrTranslation          ErrorCodes = "Error: Translation error"
	CodeErrCategory             ErrorCodes = "Error: Category error"
	CodeErrTranslationValue     ErrorCodes = "Error: TranslationValue error"
	CodeErrOrganization         ErrorCodes = "Error: Organization error"
	CodeErrOrganizationNotFound ErrorCodes = "Error: Organization not found"
	CodeErrImport               ErrorCodes = "Error: Import error"

	CodeErrNotFoundLocale      ErrorCodes = "Error: Locale not found"
	CodeErrNotFoundProject     ErrorCodes = "Error: Project not found"
	CodeErrNotFoundTranslation ErrorCodes = "Error: Translation not found"
	CodeErrNotFoundUser        ErrorCodes = "Error: User not found"

	CodeErrDBCreateLocale         ErrorCodes = "Error: Database Create Locale"
	CodeErrDBCreateUser           ErrorCodes = "Error: Database Create User"
	CodeErrCreateProject          ErrorCodes = "Error: Database Create Project"
	CodeErrCreateTranslation      ErrorCodes = "Error: Database Create Translation"
	CodeErrCreateCategory         ErrorCodes = "Error: Database Create Category"
	CodeErrCreateOrganization     ErrorCodes = "Error: Database Create Organization"
	CodeErrCreateTranslationValue ErrorCodes = "Error: Database Create TranslationValue"

	CodeErrDBUpdateLocale         ErrorCodes = "Error: Database Update Locale"
	CodeErrUpdateProject          ErrorCodes = "Error: Database Update Project"
	CodeErrUpdateTranslation      ErrorCodes = "Error: Database Update Translation"
	CodeErrUpdateCategory         ErrorCodes = "Error: Database Update Category"
	CodeErrUpdateOrganization     ErrorCodes = "Error: Database Update Organization"
	CodeErrUpdateTranslationValue ErrorCodes = "Error: Database Update TranslationValue"
)
