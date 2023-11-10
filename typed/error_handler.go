package typed

import (
	"encoding/json"
	"net/http"
)

// ErrorHandler is an interface that can be used to write an error to the response
//
// an ErrorHandler can be passed as an option to NewTypedMethodsHandlerBuilder
//
// The original *http.Request is passed so that responses can be written in the required content type (i.e. according to the `Accept` header)
type ErrorHandler interface {
	HandleError(writer http.ResponseWriter, request *http.Request, err error)
}

var defaultErrorHandler ErrorHandler = &errorHandler{}

type errorHandler struct {
}

func (eh *errorHandler) HandleError(writer http.ResponseWriter, request *http.Request, err error) {
	switch et := err.(type) {
	case ApiError:
		writer.WriteHeader(et.StatusCode())
	default:
		writer.WriteHeader(http.StatusInternalServerError)
	}
	if em, ok := err.(json.Marshaler); ok {
		if data, mErr := em.MarshalJSON(); mErr == nil {
			writer.Header().Set(hdrContentType, contentTypeJson)
			_, _ = writer.Write(data)
		} else {
			_, _ = writer.Write([]byte(err.Error() + "\n" + mErr.Error()))
		}
		return
	}
	_, _ = writer.Write([]byte(err.Error()))
}
