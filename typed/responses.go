package typed

import (
	"encoding/json"
	"encoding/xml"
	"gopkg.in/yaml.v3"
	"net/http"
	"strings"
)

func defaultStatusCode(sc int, defaultSc int) int {
	if sc < http.StatusContinue {
		return defaultSc
	}
	return sc
}

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
	// StatusCode is the status code for the response (if less than 100 Continue then 200 OK is assumed)
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
			if len(bt) == 0 {
				sc = defaultStatusCode(sc, http.StatusNoContent)
			}
		default:
			if data, err = json.Marshal(jr.Body); err != nil {
				return err
			}
		}
	} else {
		sc = defaultStatusCode(sc, http.StatusNoContent)
	}
	sc = defaultStatusCode(sc, http.StatusOK)
	writer.Header().Set(hdrContentType, contentTypeJson)
	for _, hd := range jr.Headers {
		writer.Header().Set(hd[0], hd[1])
	}
	writer.WriteHeader(sc)
	_, _ = writer.Write(data)
	return nil
}

var _ ResponseMarshaler = &MultiResponseMarshaler{}

// MultiResponseMarshaler is an implementation of ResponseMarshaler
// that supports marshalling a response value as json, yaml or xml - according to the request `Accept` header
//
// If the `Accept` header does not indicate json, yaml or xml then an ApiError indicating status code 406 Not Acceptable
// (unless a FallbackContentType is supplied)
type MultiResponseMarshaler struct {
	// Error is the optional error that can be returned
	Error error
	// StatusCode is the status code for the response (if less than 100 Continue then 200 OK is assumed)
	StatusCode int
	// Headers is any headers to be set on the response (note the `Content-Type` header will be set automatically but can be overridden by these)
	Headers [][2]string
	// Body is any payload to be marshaled for the response body
	//
	// If Body is nil the status code will default 204 No Content (i.e. if StatusCode is less than 100)
	//
	// Note: It is up to you to ensure that the Body can be reliably marshalled into all json, yaml and xml - for example, map[string]any will NOT marshal into xml
	Body any
	// FallbackContentType is the content type to assume if the `Accept` header is not one of json, yaml or xml
	//
	// If this value is empty (or not one of "application/json", "application/yaml" or "application/xml") then
	// no fallback is used
	FallbackContentType string
	// ExcludeJson if set, prevents MultiResponseMarshaler supporting json
	ExcludeJson bool
	// ExcludeYaml if set, prevents MultiResponseMarshaler supporting yaml
	ExcludeYaml bool
	// ExcludeXml if set, prevents MultiResponseMarshaler supporting xml
	ExcludeXml bool
}

func (m *MultiResponseMarshaler) Marshal(request *http.Request) (data []byte, statusCode int, hdrs [][2]string, err error) {
	at := request.Header.Get(hdrAccept)
	if cAt := strings.IndexByte(at, ';'); cAt != -1 {
		at = at[:cAt]
	}
	if pAt := strings.IndexByte(at, '+'); pAt != -1 {
		at = at[:pAt]
	}
	statusCode = defaultStatusCode(m.StatusCode, http.StatusOK)
	for i, ct := range []string{at, m.FallbackContentType} {
		switch ct {
		case contentTypeJson:
			if m.ExcludeJson {
				err = NewApiError(http.StatusNotAcceptable, "")
				return
			}
			hdrs = m.finalHeaders(ct)
			if m.Body == nil {
				statusCode = defaultStatusCode(m.StatusCode, http.StatusNoContent)
			} else {
				data, err = json.Marshal(m.Body)
			}
			return
		case contentTypeYaml, contentTypeYamlX, contentTypeYamlTxt:
			if m.ExcludeYaml {
				err = NewApiError(http.StatusNotAcceptable, "")
				return
			}
			hdrs = m.finalHeaders(ct)
			if m.Body == nil {
				statusCode = defaultStatusCode(m.StatusCode, http.StatusNoContent)
			} else {
				data, err = yaml.Marshal(m.Body)
			}
			return
		case contentTypeXml, contentTypeXmlTxt:
			if m.ExcludeXml {
				err = NewApiError(http.StatusNotAcceptable, "")
				return
			}
			hdrs = m.finalHeaders(ct)
			if m.Body == nil {
				statusCode = defaultStatusCode(m.StatusCode, http.StatusNoContent)
			} else {
				data, err = xml.Marshal(m.Body)
			}
			return
		default:
			if i > 0 || m.FallbackContentType == "" {
				err = NewApiError(http.StatusNotAcceptable, "")
			}
		}
		if err != nil {
			break
		}
	}
	return
}

func (m *MultiResponseMarshaler) finalHeaders(contentType string) [][2]string {
	for _, hdr := range m.Headers {
		if hdr[0] == hdrContentType {
			return m.Headers
		}
	}
	return append(m.Headers, [2]string{hdrContentType, contentType})
}
