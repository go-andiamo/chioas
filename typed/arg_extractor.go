package typed

import (
	"errors"
	"fmt"
	"github.com/go-andiamo/urit"
	"net/http"
	"reflect"
	"strings"
)

// ArgExtractor is a struct that can be passed as an option to NewTypedMethodsHandlerBuilder
// and contains a function that extracts the typed arg from the request
//
// For example, if your API had multiple places where an 'id' path param was used - you could alias the id type...
//
//	type Id string
//
// and then in the handler...
//
//	func DeletePerson(personId Id) {
//	  ...
//	}
//
// and create an ArgExtractor to get the id from the path param...
//
//	 idExtractor := &ArgExtractor[Id]{
//		  Extract: func(r *http.Request) (Id, error) {
//		    return Id(chi.URLParam(r, "id")), nil
//		  },
//		}
//
// and use the ArgExtractor...
//
//	var myApiDef = chioas.Definition{
//	  ...
//	  MethodHandlerBuilder: typed.NewTypedMethodsHandlerBuilder(idExtractor),
//	  ...
//	}
//
// Or, if the extractor for the given arg type never reads the request body, you can simply specify
// the extractor as a func - with the signature:
//
//	func(r *http.Request) (T, error)
//
// where T is the arg type
type ArgExtractor[T any] struct {
	// Extract is the function that extracts the arg value from the request
	Extract func(r *http.Request) (T, error)
	// ReadsBody denotes that the arg uses the request body
	ReadsBody bool
}

type argExtractors map[reflect.Type]*argExtractor

func (a argExtractors) add(ax *argExtractor) error {
	if _, ok := a[ax.forType]; ok {
		return fmt.Errorf("multiple mappings for arg type '%s'", ax.forType.String())
	}
	a[ax.forType] = ax
	return nil
}

func (a argExtractors) IsApplicable(argType reflect.Type, method string, path string) (is bool, readsBody bool) {
	if ax, ok := a[argType]; ok {
		return true, ax.readsBody
	}
	return false, false
}
func (a argExtractors) BuildValue(argType reflect.Type, request *http.Request, pathParams []urit.PathVar) (reflect.Value, error) {
	if ax, ok := a[argType]; ok {
		return ax.extract(request)
	}
	return reflect.Value{}, fmt.Errorf("no mapping for type '%s'", argType.String())
}

type argExtractor struct {
	readsBody bool
	forType   reflect.Type
	fn        reflect.Value
}

func (ax *argExtractor) extract(request *http.Request) (reflect.Value, error) {
	out := ax.fn.Call([]reflect.Value{reflect.ValueOf(request)})
	if out[1].IsNil() {
		return out[0], nil
	}
	return out[0], out[1].Interface().(error)
}

var argType = strings.SplitN(reflect.TypeOf(ArgExtractor[string]{}).String(), "[", 2)[0] + "["
var argPkg = reflect.TypeOf(ArgExtractor[string]{}).PkgPath()
var reqType = reflect.TypeOf(&http.Request{})

func isArgExtractor(v any) (*argExtractor, error) {
	if v != nil {
		vt := reflect.TypeOf(v)
		if vt.Kind() == reflect.Func {
			if vt.NumIn() == 1 && vt.NumOut() == 2 && vt.In(0) == reqType && vt.Out(1) == interfaceTypeError {
				return &argExtractor{
					readsBody: false,
					forType:   vt.Out(0),
					fn:        reflect.ValueOf(v),
				}, nil
			}
			return nil, errors.New("arg extractor func must have signature 'func(*http.Request) (T, error)'")
		} else {
			isPtr := vt.Kind() == reflect.Pointer
			if vt.Kind() == reflect.Struct || (isPtr && vt.Elem().Kind() == reflect.Struct) {
				vts := vt.String()
				var fn reflect.Value
				var readsBody bool
				is := false
				if is = strings.HasPrefix(vts, argType) && vt.PkgPath() == argPkg; is {
					vo := reflect.ValueOf(v)
					fn = vo.FieldByName("Extract")
					readsBody = vo.FieldByName("ReadsBody").Interface().(bool)
				} else if is = isPtr && strings.HasPrefix(vts, "*"+argType) && vt.Elem().PkgPath() == argPkg; is {
					vo := reflect.ValueOf(v).Elem()
					fn = vo.FieldByName("Extract")
					readsBody = vo.FieldByName("ReadsBody").Interface().(bool)
				}
				if is {
					if !fn.IsValid() || fn.IsNil() || fn.Type().Kind() != reflect.Func {
						return nil, errors.New("ArgExtractor[T] missing 'Extract' function")
					} else {
						return &argExtractor{
							forType:   fn.Type().Out(0),
							fn:        fn,
							readsBody: readsBody,
						}, nil
					}
				}
			}
		}
	}
	return nil, nil
}
