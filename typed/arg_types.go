package typed

import (
	"github.com/go-andiamo/urit"
	"net/http"
	"reflect"
)

// ArgBuilder is an interface that can be passed as an option to NewTypedMethodsHandlerBuilder, allowing support
// for additional types
type ArgBuilder interface {
	IsApplicable(argType reflect.Type, method string, path string) (is bool, readsBody bool)
	BuildValue(argType reflect.Type, writer http.ResponseWriter, request *http.Request, params []urit.PathVar) (reflect.Value, error)
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
