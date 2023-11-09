package typed

import (
	"fmt"
	"net/http"
)

// ApiError is an error interface that can be returned from typed method/func handlers
// allowing the status code for the error to be set in the response
//
// Implementations of this interface can also be used by ResponseMarshaler.Marshal and JsonResponse.Error
type ApiError interface {
	error
	StatusCode() int
	Wrapped() error
}

type apiError struct {
	statusCode int
	msg        string
	err        error
}

func (err *apiError) Error() string {
	if err.err != nil && err.msg == "" {
		if err.err.Error() == "" {
			return http.StatusText(err.statusCode)
		}
		return err.err.Error()
	}
	return err.msg
}

func (err *apiError) StatusCode() int {
	return err.statusCode
}

func (err *apiError) Wrapped() error {
	return err.err
}

// NewApiError creates a new ApiError with the specified status code and error message
//
// If the message is an empty string, the actual message is set from the status code using http.StatusText
func NewApiError(statusCode int, msg string) ApiError {
	if statusCode < http.StatusContinue {
		statusCode = http.StatusInternalServerError
	}
	if msg == "" {
		msg = http.StatusText(statusCode)
	}
	return &apiError{
		statusCode: statusCode,
		msg:        msg,
	}
}

// NewApiErrorf creates a new ApiError with the specified status code and error format + args
//
// If the message is an empty string, the actual message is set from the status code using http.StatusText
func NewApiErrorf(statusCode int, format string, a ...any) ApiError {
	if statusCode < http.StatusContinue {
		statusCode = http.StatusInternalServerError
	}
	msg := fmt.Sprintf(format, a...)
	if msg == "" {
		msg = http.StatusText(statusCode)
	}
	return &apiError{
		statusCode: statusCode,
		msg:        msg,
	}
}

// WrapApiError creates a new ApiError by wrapping the error and using the provided status code
//
// Note: If the provided error is nil then nil is returned
//
// Note: If the provided status code is less than 100 (http.StatusContinue) the status code http.StatusInternalServerError is used
func WrapApiError(statusCode int, err error) ApiError {
	if err == nil {
		return nil
	}
	if statusCode < http.StatusContinue {
		statusCode = http.StatusInternalServerError
	}
	return &apiError{
		statusCode: statusCode,
		err:        err,
	}
}

// WrapApiErrorMsg creates a new ApiError by wrapping the error - using the provided status code and optionally overriding the message
//
// Note: If the provided error is nil then nil is returned
//
// Note: If the provided status code is less than 100 (http.StatusContinue) the status code http.StatusInternalServerError is used
func WrapApiErrorMsg(statusCode int, err error, msg string) ApiError {
	if err == nil {
		return nil
	}
	if statusCode < http.StatusContinue {
		statusCode = http.StatusInternalServerError
	}
	return &apiError{
		statusCode: statusCode,
		msg:        msg,
		err:        err,
	}
}
