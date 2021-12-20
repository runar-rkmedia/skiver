package requestContext

import "fmt"

type ErrorCodes string
type Error struct {
	Code    ErrorCodes
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (%s)", e.Message, e.Code)
}

func NewError(msg string, code ErrorCodes) Error {
	return Error{code, msg}
}

var (
	ErrIDNonValid = NewError("Id was not valid", CodeErrIDNonValid)
	ErrIDTooLong  = NewError("Id was too long", CodeErrIDTooLong)
	ErrIDEmpty    = NewError("Id was empty", CodeErrIDEmpty)
)

const (
	CodeErrMethodNotAllowed ErrorCodes = "Error: HTTP-Method is not allowed"
	CodeErrNoRoute          ErrorCodes = "Error: No route matched for this http-path"
	CodeErrReadBody         ErrorCodes = "Error: Failed to read body"
	CodeErrMarhal           ErrorCodes = "Error: Failed to marshal"
	CodeErrUnmarshal        ErrorCodes = "Error: Failed to unmarshal"
	CodeErrJmesPath         ErrorCodes = "Error: JmesPath"
	CodeErrJmesPathMarshal  ErrorCodes = "Error: JmesPathMarshal"

	CodeErrRequestEntityTooLarge ErrorCodes = "Error: Request Entity too large"
	CodeErrInputValidation       ErrorCodes = "Error: General input validation"
	CodeErrIDNonValid            ErrorCodes = "Error: ID not valid"
	CodeErrIDTooLong             ErrorCodes = "Error: ID is too long"
	CodeErrIDEmpty               ErrorCodes = "Error: ID was Empty"

	CodeErrAuthenticationRequired ErrorCodes = "Error: Authentication required"
	CodeErrAuthoriziationFailed   ErrorCodes = "Error: Authorization failed"

	CodeErrLocale ErrorCodes = "Error: Locale error"

	CodeErrNotFoundLocale ErrorCodes = "Error: Locale not found"
	CodeErrNotFoundUser   ErrorCodes = "Error: User not found"

	CodeErrDBCreateLocale ErrorCodes = "Error: Database Create Locale"
)
