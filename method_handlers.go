package chioas

import (
	"fmt"
	"net/http"
	"reflect"
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
	}
	return nil, fmt.Errorf("invalid handler type (path: %s, method: %s)", path, method)
}
