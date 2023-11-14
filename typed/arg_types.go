package typed

import (
	"github.com/go-andiamo/urit"
	"net/http"
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

// Headers is a type that can be used as a handler method/func arg to receive request headers
type Headers map[string][]string

// PathParams is a type that can be used as a handler method/func arg to receive request path params
//
// Another way to receive request path params (in order) is to use either []string or ...string (varadic)
// examples:
//
//	func getSomething(pathParams []string) (json.RawMessage, error)
//
//	func getSomething(pathParams ..string) (json.RawMessage, error)
type PathParams map[string][]string

// QueryParams is a type that can be used as a handler method/func arg to receive request query params
type QueryParams map[string][]string

// RawQuery is a type that can be used as a handler method/func arg to receive request raw query
type RawQuery string
