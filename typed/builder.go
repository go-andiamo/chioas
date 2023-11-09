package typed

import (
	"encoding/json"
	"fmt"
	"github.com/go-andiamo/chioas"
	"net/http"
	"reflect"
)

const (
	hdrContentType  = "Content-Type"
	contentTypeJson = "application/json"
)

// NewTypedMethodsHandlerBuilder creates a new handler for use on chioas.Definition and provides
// capability to have typed methods/funcs for API endpoints.
//
// the options arg can be any of types ErrorHandler, ArgBuilder or Unmarshaler
//
// if no Unmarshaler is passed then a default JSON unmarshaler is used - and if multiple Unmarshaler are passed then only the last one is used
//
// For a complete example, see package docs
func NewTypedMethodsHandlerBuilder(options ...any) chioas.MethodHandlerBuilder {
	result := &typedMethodsHandlerBuilder{
		errorHandler: nil,
		argBuilders:  make([]ArgBuilder, 0, len(options)),
		unmarshaler:  defaultUnmarshaler,
	}
	for _, o := range options {
		if o != nil {
			if eh, ok := o.(ErrorHandler); ok {
				result.errorHandler = eh
			} else if vb, ok := o.(ArgBuilder); ok {
				result.argBuilders = append(result.argBuilders, vb)
			} else if um, ok := o.(Unmarshaler); ok {
				result.unmarshaler = um
			}
		}
	}
	return result
}

type typedMethodsHandlerBuilder struct {
	errorHandler ErrorHandler
	argBuilders  []ArgBuilder
	unmarshaler  Unmarshaler
}

func (t *typedMethodsHandlerBuilder) BuildHandler(path string, method string, mdef chioas.Method, thisApi any) (http.HandlerFunc, error) {
	if mdef.Handler == nil {
		return nil, fmt.Errorf("handler not set (path: %s, method: %s)", path, method)
	} else if hf, ok := mdef.Handler.(http.HandlerFunc); ok {
		return hf, nil
	} else if hf, ok := mdef.Handler.(func(http.ResponseWriter, *http.Request)); ok {
		return hf, nil
	} else if gh, ok := mdef.Handler.(func(string, string, any) (http.HandlerFunc, error)); ok {
		return gh(path, method, thisApi)
	} else if mn, ok := mdef.Handler.(string); ok {
		if thisApi != nil {
			return t.buildFromMethodName(path, method, thisApi, mn)
		} else {
			return nil, fmt.Errorf("method by name '%s' can only be used when 'thisApi' arg is passed to Definition.SetupRoutes (path: %s, method: %s)", mn, path, method)
		}
	} else if mfn := reflect.ValueOf(mdef.Handler); mfn.IsValid() && mfn.Kind() == reflect.Func {
		if hf, err := t.handlerFor(path, method, thisApi, mfn); err == nil {
			return hf, nil
		} else {
			return nil, err
		}
	}
	return nil, fmt.Errorf("invalid handler type (path: %s, method: %s)", path, method)
}

func (t *typedMethodsHandlerBuilder) buildFromMethodName(path string, method string, thisApi any, methodName string) (http.HandlerFunc, error) {
	mf := reflect.ValueOf(thisApi).MethodByName(methodName)
	if !mf.IsValid() {
		return nil, fmt.Errorf("method name '%s' does not exist (path: %s, method: %s)", methodName, path, method)
	}
	if hf, ok := mf.Interface().(func(http.ResponseWriter, *http.Request)); ok {
		return hf, nil
	}
	return t.handlerFor(path, method, thisApi, mf)
}

func (t *typedMethodsHandlerBuilder) handlerFor(path string, method string, thisApi any, mf reflect.Value) (http.HandlerFunc, error) {
	mft := mf.Type()
	ins, outs := mft.NumIn(), mft.NumOut()
	_ = outs
	if ins == 0 && outs == 0 {
		return func(http.ResponseWriter, *http.Request) {
			mf.Call([]reflect.Value{})
		}, nil
	} else if ins == 0 {
		return t.zeroInHandler(thisApi, mf), nil
	}
	return t.ioHandler(path, method, thisApi, mf)
}

func (t *typedMethodsHandlerBuilder) zeroInHandler(thisApi any, mf reflect.Value) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		t.handleReturnArgs(mf.Call([]reflect.Value{}), thisApi, writer, request)
	}
}

func (t *typedMethodsHandlerBuilder) ioHandler(path string, method string, thisApi any, mf reflect.Value) (http.HandlerFunc, error) {
	inb, err := newInsBuilder(mf, path, method, t.unmarshaler, t.argBuilders...)
	if err != nil {
		return nil, fmt.Errorf("error building in args (path: %s, method: %s) - %s", path, method, err.Error())
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		if inArgs, err := inb.build(writer, request); err == nil {
			t.handleReturnArgs(mf.Call(inArgs), thisApi, writer, request)
		} else {
			t.getErrorHandler(thisApi).HandleError(writer, request, err)
		}
	}, nil
}

func (t *typedMethodsHandlerBuilder) handleReturnArgs(retArgs []reflect.Value, thisApi any, writer http.ResponseWriter, request *http.Request) {
	// see if any return args are an error...
	for _, retArg := range retArgs {
		if retArg.IsValid() {
			if err, ok := retArg.Interface().(error); ok {
				t.getErrorHandler(thisApi).HandleError(writer, request, err)
				return
			}
		}
	}
	// no errors - find the first valid return arg to write as the response...
	for _, retArg := range retArgs {
		if retArg.IsValid() {
			t.handleResponseArg(retArg.Interface(), thisApi, writer, request)
			return
		}
	}
}

func (t *typedMethodsHandlerBuilder) handleResponseArg(res any, thisApi any, writer http.ResponseWriter, request *http.Request) {
	if rm, ok := res.(ResponseMarshaler); ok {
		if data, sc, hdrs, err := rm.Marshal(request); err == nil {
			if sc < http.StatusContinue && len(data) == 0 {
				sc = http.StatusNoContent
			} else if sc < http.StatusContinue {
				sc = http.StatusOK
			}
			writer.WriteHeader(sc)
			for _, hd := range hdrs {
				writer.Header().Set(hd[0], hd[1])
			}
			_, _ = writer.Write(data)
		} else {
			t.getErrorHandler(thisApi).HandleError(writer, request, err)
		}
		return
	}
	switch rt := res.(type) {
	case JsonResponse:
		if err := rt.write(writer); err != nil {
			t.getErrorHandler(thisApi).HandleError(writer, request, err)
		}
	case *JsonResponse:
		if rt != nil {
			if err := rt.write(writer); err != nil {
				t.getErrorHandler(thisApi).HandleError(writer, request, err)
			}
		} else {
			writer.WriteHeader(http.StatusNoContent)
		}
	case []byte:
		if len(rt) == 0 {
			writer.WriteHeader(http.StatusNoContent)
		} else {
			writer.WriteHeader(http.StatusOK)
		}
		_, _ = writer.Write(rt)
	default:
		if data, err := json.Marshal(res); err == nil {
			writer.WriteHeader(http.StatusOK)
			writer.Header().Set(hdrContentType, contentTypeJson)
			_, _ = writer.Write(data)
		} else {
			t.getErrorHandler(thisApi).HandleError(writer, request, err)
		}
	}
}

func (t *typedMethodsHandlerBuilder) getErrorHandler(thisApi any) ErrorHandler {
	if t.errorHandler != nil {
		return t.errorHandler
	} else if thisApi != nil {
		if eh, ok := thisApi.(ErrorHandler); ok {
			return eh
		}
	}
	return defaultErrorHandler
}
