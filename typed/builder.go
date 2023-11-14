package typed

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-andiamo/chioas"
	"net/http"
	"reflect"
)

const (
	hdrContentType  = "Content-Type"
	hdrAccept       = "Accept"
	contentTypeJson = "application/json"
)

// NewTypedMethodsHandlerBuilder creates a new handler for use on chioas.Definition and provides
// capability to have typed methods/funcs for API endpoints.
//
// the options arg can be any of types ErrorHandler, ArgBuilder, PathParamArgBuilder or Unmarshaler
//
// if no Unmarshaler is passed then a default JSON unmarshaler is used - and if multiple Unmarshaler are passed then only the last one is used
//
// For a complete example, see package docs
func NewTypedMethodsHandlerBuilder(options ...any) chioas.MethodHandlerBuilder {
	result := &builder{
		errorHandler: nil,
		argBuilders:  make([]ArgBuilder, 0, len(options)),
		unmarshaler:  defaultUnmarshaler,
	}
	for _, o := range options {
		if o != nil {
			switch ot := o.(type) {
			case ErrorHandler:
				result.errorHandler = ot
			case ArgBuilder:
				result.argBuilders = append(result.argBuilders, ot)
			case Unmarshaler:
				result.unmarshaler = ot
			default:
				if result.initErr == nil {
					result.initErr = errors.New("invalid option passed to NewTypedMethodsHandlerBuilder")
				}
			}
		}
	}
	return result
}

type builder struct {
	errorHandler ErrorHandler
	argBuilders  []ArgBuilder
	unmarshaler  Unmarshaler
	initErr      error
}

func (b *builder) BuildHandler(path string, method string, mdef chioas.Method, thisApi any) (http.HandlerFunc, error) {
	if b.initErr != nil {
		return nil, b.initErr
	} else if mdef.Handler == nil {
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
		if thisApi != nil {
			return b.buildFromMethodName(path, method, thisApi, hf)
		} else {
			return nil, fmt.Errorf("method by name '%s' can only be used when 'thisApi' arg is passed to Definition.SetupRoutes (path: %s, method: %s)", hf, path, method)
		}
	default:
		if mfn := reflect.ValueOf(mdef.Handler); mfn.IsValid() && mfn.Kind() == reflect.Func {
			if hf, err := b.handlerFor(path, method, thisApi, mfn); err == nil {
				return hf, nil
			} else {
				return nil, err
			}
		}
	}
	return nil, fmt.Errorf("invalid handler type (path: %s, method: %s)", path, method)
}

func (b *builder) buildFromMethodName(path string, method string, thisApi any, methodName string) (http.HandlerFunc, error) {
	mf := reflect.ValueOf(thisApi).MethodByName(methodName)
	if !mf.IsValid() {
		return nil, fmt.Errorf("method name '%s' does not exist (path: %s, method: %s)", methodName, path, method)
	}
	if hf, ok := mf.Interface().(func(http.ResponseWriter, *http.Request)); ok {
		return hf, nil
	}
	return b.handlerFor(path, method, thisApi, mf)
}

func (b *builder) handlerFor(path string, method string, thisApi any, mf reflect.Value) (http.HandlerFunc, error) {
	mft := mf.Type()
	ins, outs := mft.NumIn(), mft.NumOut()
	_ = outs
	if ins == 0 && outs == 0 {
		return func(http.ResponseWriter, *http.Request) {
			mf.Call([]reflect.Value{})
		}, nil
	} else if ins == 0 {
		return b.zeroInHandler(thisApi, mf), nil
	}
	return b.ioHandler(path, method, thisApi, mf)
}

func (b *builder) zeroInHandler(thisApi any, mf reflect.Value) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		b.handleReturnArgs(mf.Call([]reflect.Value{}), thisApi, writer, request)
	}
}

func (b *builder) ioHandler(path string, method string, thisApi any, mf reflect.Value) (http.HandlerFunc, error) {
	inb, err := newInsBuilder(mf, path, method, b)
	if err != nil {
		return nil, fmt.Errorf("error building in args (path: %s, method: %s) - %s", path, method, err.Error())
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		if inArgs, err := inb.build(writer, request); err == nil {
			b.handleReturnArgs(mf.Call(inArgs), thisApi, writer, request)
		} else {
			b.getErrorHandler(thisApi).HandleError(writer, request, err)
		}
	}, nil
}

func (b *builder) handleReturnArgs(retArgs []reflect.Value, thisApi any, writer http.ResponseWriter, request *http.Request) {
	// see if any return args are an error...
	for _, retArg := range retArgs {
		if retArg.IsValid() {
			if err, ok := retArg.Interface().(error); ok {
				b.getErrorHandler(thisApi).HandleError(writer, request, err)
				return
			}
		}
	}
	// no errors - find the first valid return arg to write as the response...
	for _, retArg := range retArgs {
		if retArg.IsValid() {
			b.handleResponseArg(retArg.Interface(), thisApi, writer, request)
			return
		}
	}
}

func (b *builder) handleResponseArg(res any, thisApi any, writer http.ResponseWriter, request *http.Request) {
	if rm, ok := res.(ResponseMarshaler); ok {
		if data, sc, hdrs, err := rm.Marshal(request); err == nil {
			if len(data) == 0 {
				sc = defaultStatusCode(sc, http.StatusNoContent)
			} else {
				sc = defaultStatusCode(sc, http.StatusOK)
			}
			for _, hd := range hdrs {
				writer.Header().Set(hd[0], hd[1])
			}
			writer.WriteHeader(sc)
			_, _ = writer.Write(data)
		} else {
			b.getErrorHandler(thisApi).HandleError(writer, request, err)
		}
		return
	}
	switch rt := res.(type) {
	case JsonResponse:
		if err := rt.write(writer); err != nil {
			b.getErrorHandler(thisApi).HandleError(writer, request, err)
		}
	case *JsonResponse:
		if rt != nil {
			if err := rt.write(writer); err != nil {
				b.getErrorHandler(thisApi).HandleError(writer, request, err)
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
			writer.Header().Set(hdrContentType, contentTypeJson)
			writer.WriteHeader(http.StatusOK)
			_, _ = writer.Write(data)
		} else {
			b.getErrorHandler(thisApi).HandleError(writer, request, err)
		}
	}
}

func (b *builder) getErrorHandler(thisApi any) ErrorHandler {
	if b.errorHandler != nil {
		return b.errorHandler
	} else if thisApi != nil {
		if eh, ok := thisApi.(ErrorHandler); ok {
			return eh
		}
	}
	return defaultErrorHandler
}