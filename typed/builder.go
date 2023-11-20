package typed

import (
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

// BuildHandler is normally called from chioas when building handlers (i.e. it implements the chioas.MethodHandlerBuilder interface)
//
// It can also be called directly for building handlers for testing - for example:
//
//	func typedHandler(req *http.Request) (any, error) {
//	    return nil, errors.New("fooey")
//	}
//
//	func TestTypedHandler(t *testing.T) {
//	    mb := typed.NewTypedMethodsHandlerBuilder()
//	    // build the handler to be tested...
//	    hf, err := mb.BuildHandler("/", http.MethodGet, chioas.Method{Handler: typedHandler}, nil)
//	    if err != nil {
//	        t.Fatal(err)
//	    }
//	    // create test request and writer...
//	    request, _ := http.NewRequest(http.MethodGet, "/", nil)
//	    writer := httptest.NewRecorder()
//	    // test the handler...
//	    hf.ServeHTTP(writer, request)
//	    if writer.Code != http.StatusInternalServerError {
//	        t.Fatalf("expected status %d but got %d", http.StatusInternalServerError, writer.Code)
//	    }
//	}
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
		return b.zeroInHandler(path, method, thisApi, mf)
	}
	return b.ioHandler(path, method, thisApi, mf)
}

func (b *builder) zeroInHandler(path string, method string, thisApi any, mf reflect.Value) (http.HandlerFunc, error) {
	outs, err := newOutsBuilder(mf)
	if err != nil {
		return nil, fmt.Errorf("error building return args (path: %s, method: %s) - %s", path, method, err.Error())
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		outs.handleReturnArgs(mf.Call([]reflect.Value{}), b, thisApi, writer, request)
	}, nil
}

func (b *builder) ioHandler(path string, method string, thisApi any, mf reflect.Value) (http.HandlerFunc, error) {
	ins, err := newInsBuilder(mf, path, method, b)
	if err != nil {
		return nil, fmt.Errorf("error building in args (path: %s, method: %s) - %s", path, method, err.Error())
	}
	outs, err := newOutsBuilder(mf)
	if err != nil {
		return nil, fmt.Errorf("error building return args (path: %s, method: %s) - %s", path, method, err.Error())
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		if inArgs, err := ins.build(writer, request); err == nil {
			outs.handleReturnArgs(mf.Call(inArgs), b, thisApi, writer, request)
		} else {
			b.getErrorHandler(thisApi).HandleError(writer, request, err)
		}
	}, nil
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
