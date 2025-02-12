package errdefs

import (
	"errors"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type AppError struct {
	Message    string
	Code       string
	Extensions map[string]interface{}
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code, message string, extensions map[string]interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Extensions: extensions,
	}
}

func HandleError(err error) *gqlerror.Error {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return &gqlerror.Error{
			Message: appErr.Message,
			Extensions: map[string]interface{}{
				"code":    appErr.Code,
				"details": appErr.Extensions,
			},
		}
	}

	return &gqlerror.Error{
		Message: "Unexpected error occurred",
		Extensions: map[string]interface{}{
			"code": "INTERNAL_SERVER_ERROR",
		},
	}
}
