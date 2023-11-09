package typed

import (
	"encoding/json"
	"net/http"
)

// ResponseMarshaler is an interface that can be implemented by objects returned from a typed method/func
//
// The body of the response is written by the data provided from Marshal, the response status
// code is also set from that returned and so are any headers
type ResponseMarshaler interface {
	// Marshal marshals the response for writing
	//
	// the *http.Request is passed as an arg so that the marshalling can take account of request type (i.e. `Accept` header)
	Marshal(request *http.Request) (data []byte, statusCode int, hdrs [][2]string, err error)
}

// JsonResponse is a struct that can be returned from a typed method/func
//
// The response error, status code, headers and body are determined by the properties
type JsonResponse struct {
	// Error is the optional error that can be returned
	Error error
	// StatusCode is the status code for the response
	StatusCode int
	// Headers is any headers to be set on the response
	Headers [][2]string
	// Body is any payload for the response body (marshalled to JSON)
	Body any
}

func (jr *JsonResponse) write(writer http.ResponseWriter) error {
	if jr.Error != nil {
		return jr.Error
	}
	var data []byte
	var err error
	sc := jr.StatusCode
	if jr.Body != nil {
		switch bt := jr.Body.(type) {
		case json.RawMessage:
			data = bt
			if len(bt) == 0 && sc < http.StatusContinue {
				sc = http.StatusNoContent
			}
		default:
			if data, err = json.Marshal(jr.Body); err != nil {
				return err
			}
		}
	} else if sc < http.StatusContinue {
		sc = http.StatusNoContent
	}
	if sc < http.StatusContinue {
		sc = http.StatusOK
	}
	writer.WriteHeader(sc)
	writer.Header().Set(hdrContentType, contentTypeJson)
	for _, hd := range jr.Headers {
		writer.Header().Set(hd[0], hd[1])
	}
	_, _ = writer.Write(data)
	return nil
}
