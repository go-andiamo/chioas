package typed

import (
	"github.com/go-andiamo/urit"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
)

// ArgBuilder is an interface that can be passed as an option to NewTypedMethodsHandlerBuilder, allowing support
// for additional typed handler arg types
type ArgBuilder interface {
	// IsApplicable determines whether this ArgBuilder can handle the given arg reflect.Type
	//
	// If it is applicable, this method should return true - and return readsBody true if it intends to read the request body (as only one arg can read the request body)
	//
	// The method and path are provided for information purposes
	IsApplicable(argType reflect.Type, method string, path string) (is bool, readsBody bool)
	// BuildValue builds the final arg reflect.Value that will be used to call the typed handler
	//
	// If no error is returned, then the reflect.Value returned MUST match the arg type (failure to do so will result in an error response)
	BuildValue(argType reflect.Type, request *http.Request, params []urit.PathVar) (reflect.Value, error)
}

// Headers is a type that can be used as a typed handler arg to receive request headers
type Headers map[string][]string

// PathParams is a type that can be used as a typed handler arg to receive request path params
//
// Another way to receive request path params (in order) is to use either []string or ...string (varadic)
// examples:
//
//	func getSomething(pathParams []string) (json.RawMessage, error)
//
//	func getSomething(pathParams ..string) (json.RawMessage, error)
type PathParams map[string][]string

// QueryParams is a type that can be used as a typed handler arg to receive request query params
type QueryParams map[string][]string

// RawQuery is a type that can be used as a typed handler arg to receive request raw query
type RawQuery string

var multipartFormType = reflect.TypeOf(&multipart.Form{})

// NewMultipartFormArgSupport creates an arg type builder - for use as an option passed to NewTypedMethodsHandlerBuilder(options ...any)
//
// By adding this as an option to NewTypedMethodsHandlerBuilder, any typed handler with an arg of type *multipart.Form will be supported
//
// If a typed handler has an arg type of *multipart.Form but the request is not `Content-Type=multipart/form-data`, or the request body is nil (or any other error from http.Request.ParseMultipartForm)
// then the typed handler is not called and an error response of 400 Bad Request is served - unless noAuthError is set, in which case such
// errors result in a nil *multipart.Form being passed to the typed handler
func NewMultipartFormArgSupport(maxMemory int64, noAutoError bool) ArgBuilder {
	return &multipartFormArgBuilder{maxMemory: maxMemory, noAutoError: noAutoError}
}

type multipartFormArgBuilder struct {
	maxMemory   int64
	noAutoError bool
}

func (ab *multipartFormArgBuilder) IsApplicable(argType reflect.Type, method string, path string) (is bool, readsBody bool) {
	return argType == multipartFormType, true
}

func (ab *multipartFormArgBuilder) BuildValue(argType reflect.Type, request *http.Request, params []urit.PathVar) (reflect.Value, error) {
	err := request.ParseMultipartForm(ab.maxMemory)
	if err != nil {
		if ab.noAutoError {
			var empty *multipart.Form
			return reflect.ValueOf(empty), nil
		}
		return reflect.Value{}, WrapApiError(http.StatusBadRequest, err)
	}
	return reflect.ValueOf(request.MultipartForm), nil
}

// PostForm is a type that can be used as a typed handler arg to receive request PostForm values
//
// Note: If this arg type is used for a typed handler that does not handle http methods POST, PUT or PATH - then
// the value will be empty (and the request body will not have been read)
type PostForm url.Values

// BasicAuth is a type that can be used as a typed handler arg to receive request BasicAuth
//
// If type *typed.BasicAuth is used as the typed handler arg, then the value will be nil if no Authorization header is present on the request
// or if the Authorization header is not basic auth (i.e. value does not start with "Basic ")
type BasicAuth struct {
	Username string
	Password string
	Ok       bool
}

// NamedQueryParam is an interface that enables a type to be used as a typed handler arg to receive a named query param
//
// Note: The underlying type must be a string or implement encoding.TextUnmarshaler
//
// Example:
//
//	// declare type and name...
//	type MyQueryParam string
//	func (MyQueryParam) QueryParamName() string {
//	    return "my" // tells chioas typed the query param name
//	}
//
// and then use that type as a handler arg...
//
//	func requestHandler(myQp MyQueryParam)
//
// or if the query param is optional...
//
//	func requestHandler(myQp *MyQueryParam)
//
// or if the query param has multiple values...
//
//	func requestHandler(myQps []MyQueryParam)
type NamedQueryParam interface {
	QueryParamName() string
}

// NamedPathParam is an interface that enables a type to be used as a typed handler arg to receive a named path param
//
// Note: The underlying type must be a string or implement encoding.TextUnmarshaler
//
// Example:
//
//	// declare type and name...
//	type PathParamID string
//	func (PathParamID) PathParamName() string {
//	    return "id" // tells chioas typed the path param name
//	}
//
// and then use that type as a handler arg...
//
//	func requestHandler(id PathParamID)
//
// or if the path param has multiple values...
//
//	func requestHandler(ids []PathParamID)
type NamedPathParam interface {
	PathParamName() string
}

// NamedHeader is an interface that enables a type to be used as a typed handler arg to receive a named header
//
// Note: The underlying type must be a string
//
// Example:
//
//	// declare type and name...
//	type Accept string
//	func (Accept) HeaderName() string {
//	    return "Accept" // tells chioas typed the header name
//	}
//
// and then use that type as a handler arg...
//
//	func requestHandler(accept Accept)
//
// or if the header is optional...
//
//	func requestHandler(accept *Accept)
type NamedHeader interface {
	HeaderName() string
}

// NamedCookie is an interface that enables a type to be used as a typed handler arg to receive a named cookie
//
// Note: The underlying type must be http.Cookie and usage must be a pointer
//
// Example:
//
//	// declare type and name...
//	type CSession http.Cookie
//	func (CSession) CookieName() string {
//	    return "session" // tells chioas typed the cookie name
//	}
//
// and then use that type as a handler arg...
//
//	func requestHandler(sess *CSession)
type NamedCookie interface {
	CookieName() string
}
