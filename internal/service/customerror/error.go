package customerror

import "fmt"

type (
	CustomError struct {
		Code    int
		Message string
		Err     error
	}
)

const (
	Internal int = iota
	Unprocessable
	Payment
	Unauthorized
	Conflict
)

func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Error code: %d, message: %s, original error: %s", e.Code, e.Message, e.Err.Error())
	}
	return fmt.Sprintf("Error code: %d, message: %s", e.Code, e.Message)
}

func NewCustomError(code int, message string, err error) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
