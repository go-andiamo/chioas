package chioas

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

// GetHandler is a function that can be set on Method.Handler - and is called to obtain the http.HandlerFunc
type GetHandler func(path string, method string, thisApi any) (http.HandlerFunc, error)

type MethodHandlerBuilder interface {
	BuildHandler(path string, method string, mdef Method, thisApi any) (http.HandlerFunc, error)
}

var defaultMethodHandlerBuilder MethodHandlerBuilder = &methodHandlerBuilder{}

func getMethodHandlerBuilder(builder MethodHandlerBuilder) MethodHandlerBuilder {
	if builder != nil {
		return builder
	}
	return defaultMethodHandlerBuilder
}

type methodHandlerBuilder struct {
}

func (m *methodHandlerBuilder) BuildHandler(path string, method string, mdef Method, thisApi any) (http.HandlerFunc, error) {
	if mdef.Handler == nil {
		return nil, fmt.Errorf("handler not set (path: %s, method: %s)", path, method)
	}
	switch hf := mdef.Handler.(type) {
	case http.HandlerFunc:
		return hf, nil
	case func(http.ResponseWriter, *http.Request):
		return hf, nil
	case func(string, string, any) (http.HandlerFunc, error):
		return hf(path, method, thisApi)
	case string:
		if thisApi == nil {
			return nil, fmt.Errorf("method by name '%s' can only be used when 'thisApi' arg is passed to Definition.SetupRoutes (path: %s, method: %s)", hf, path, method)
		}
		mfn := reflect.ValueOf(thisApi).MethodByName(hf)
		if !mfn.IsValid() {
			return nil, fmt.Errorf("method name '%s' does not exist (path: %s, method: %s)", hf, path, method)
		}
		if hf, ok := mfn.Interface().(func(http.ResponseWriter, *http.Request)); ok {
			return hf, nil
		}
		return nil, fmt.Errorf("method name '%s' is not http.HandlerFunc (path: %s, method: %s)", hf, path, method)
	default:
		if mf, ok := isMethodExpression(mdef.Handler, thisApi); ok {
			return mf, nil
		}
	}
	return nil, fmt.Errorf("invalid handler type (path: %s, method: %s)", path, method)
}

func isMethodExpression(m any, thisApi any) (http.HandlerFunc, bool) {
	if thisApi != nil {
		mv := reflect.ValueOf(m)
		mt := mv.Type()
		// check it's a func with 3 in args and no return...
		if mt.Kind() == reflect.Func && mt.NumIn() == 3 && mt.NumOut() == 0 {
			apiv := reflect.ValueOf(thisApi)
			// for it to be a potential method expression, the first in arg must be receiver type...
			if mt.In(0) == apiv.Type() {
				// check that the function name indicates it is actually a method expression...
				// (i.e. rather than just a regular function whose first arg type happens to be receiver type)
				mn := runtime.FuncForPC(mv.Pointer()).Name()
				if strings.Contains(mn, ".(") && strings.Contains(mn, ").") {
					mn = parseMethodName(mn)
					if mfn := apiv.MethodByName(mn); mfn.IsValid() {
						if hf, ok := mfn.Interface().(func(http.ResponseWriter, *http.Request)); ok {
							return hf, true
						}
					}
				}
			}
		}
	}
	return nil, false
}

func parseMethodName(methodName string) string {
	parts := strings.Split(methodName, ".")
	return parts[len(parts)-1]
}
