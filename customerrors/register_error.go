package customerrors

// ErrorWrapper wraps an error with an HTTP status code and a user-facing message.
type ErrorWrapper struct {
	Message string `json:"message"`
	Code    int    `json:"errCode"`
	Err     error  `json:"-"`
}

// Error returns the underlying error message, falling back to the wrapper message.
func (err ErrorWrapper) Error() string {
	if err.Err != nil {
		return err.Err.Error()
	}
	return err.Message
}

// Unwrap returns the underlying error for use with errors.Is and errors.As.
func (err ErrorWrapper) Unwrap() error {
	return err.Err
}

// Wrap creates an ErrorWrapper with the given HTTP status code, cause, and message.
func Wrap(code int, err error, message string) error {
	return ErrorWrapper{
		Message: message,
		Code:    code,
		Err:     err,
	}
}
