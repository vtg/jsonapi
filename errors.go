package jsonapi

import "strings"

var (
	// ErrorRecordNotFound returns Error for record not found behaviour
	ErrorRecordNotFound = Error{
		Status: "404",
		Title:  "Record Not Found",
		Detail: "The record you are looking for does not exist",
	}
	// ErrorPageNotFound returns Error for page not found behaviour
	ErrorPageNotFound = Error{
		Status: "404",
		Title:  "Page Not Found",
		Detail: "The page you are looking for does not exist",
	}

	// ErrorUnauthorized returns Error for unauthorized request
	ErrorUnauthorized = Error{
		Status: "401",
		Title:  "Unauthorized Request",
		Detail: "You are forbidden from accessing this page",
	}
)

// ErrorSource type
type ErrorSource struct {
	Pointer string `json:"pointer"`
}

// Error type
type Error struct {
	Code   string       `json:"code,omitempty"`
	Status string       `json:"status,omitempty"`
	Source *ErrorSource `json:"source,omitempty"`
	Title  string       `json:"title,omitempty"`
	Detail string       `json:"detail,omitempty"`
}

// Error returns Detail to implement error interface
func (e Error) Error() string {
	return e.Detail
}

// Errors type
type Errors struct {
	Errors []Error `json:"errors,omitempty"`
}

// HasErrors adds Error to errors
func (e Errors) HasErrors() bool {
	return len(e.Errors) > 0
}

// AddError adds Error to errors
func (e *Errors) AddError(err error) {
	if err == nil {
		return
	}

	switch err.(type) {
	case Error:
		e.Errors = append(e.Errors, err.(Error))
	case Errors:
		e.Errors = append(e.Errors, err.(Errors).Errors...)
	default:
		e.Errors = append(e.Errors, ErrorInternal(err.Error()))
	}
}

// Error returns Detail to implement error interface
func (e Errors) Error() string {
	msgs := make([]string, 0, len(e.Errors))
	for k := range e.Errors {
		msgs = append(msgs, e.Errors[k].Detail)
	}
	return strings.Join(msgs, ",")
}

// ErrorInternal creating Error for internal error
func ErrorInternal(details string) Error {
	return Error{
		Status: "500",
		Title:  "Internal Server Error",
		Detail: details,
	}
}

// ErrorInvalidAttribute creating Error for invalid attributes
func ErrorInvalidAttribute(pointer, details string) Error {
	return Error{
		Status: "422",
		Source: &ErrorSource{Pointer: "/data/attributes/" + pointer},
		Title:  "Invalid Attribute",
		Detail: details,
	}
}

// ErrorBadRequest creating Error for inprocessible entries
func ErrorBadRequest(details string) Error {
	return Error{
		Status: "400",
		Title:  "Bad Request",
		Detail: details,
	}
}
