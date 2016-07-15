package jsonapi

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
type Errors []Errors

// ErrorRecordNotFound creating Error for record not found behaviour
func ErrorRecordNotFound() Error {
	return Error{
		Status: "404",
		Title:  "Record Not Found",
		Detail: "The record you are looking for does not exist",
	}
}

// ErrorPageNotFound creating Error for page not found behaviour
func ErrorPageNotFound() Error {
	return Error{
		Status: "404",
		Title:  "Page Not Found",
		Detail: "The page you are looking for does not exist",
	}
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
		Source: &ErrorSource{Pointer: pointer},
		Title:  "Invalid Attribute",
		Detail: details,
	}
}

// ErrorInprocessible creating Error for inprocessible entries
func ErrorInprocessible(details string) Error {
	return Error{
		Status: "400",
		Title:  "Inprocessible Entry",
		Detail: details,
	}
}

// ErrorUnauthorized creating Error for unauthorized request
func ErrorUnauthorized() Error {
	return Error{
		Status: "401",
		Title:  "Unauthorized Request",
		Detail: "You are forbidden from accessing this page",
	}
}
