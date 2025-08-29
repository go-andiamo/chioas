package chioas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"reflect"
	"strings"
)

// FromJson builds a Definition from an existing OAS spec JSON
//
// where all methods in the spec must have a `x-handler` property, e.g.
//
//	"paths": {
//	  "/": {
//	    "get": {
//	      "x-handler": "getRoot"
//	      ...
//
// where the key "getRoot" must be provided in the handlers arg
//
// or...
//
//	"paths": {
//	  "/": {
//	    "get": {
//	      "x-handler": ".GetRoot"
//	      ...
//
// where the "GetRoot" must be a http.HandlerFunc method on the supplied api arg
func FromJson(r io.Reader, opts *FromOptions) (result *Definition, err error) {
	useOptions := defaultedFromOptions(opts, true)
	result = &Definition{
		DocOptions: *useOptions.DocOptions,
	}
	useReader := r
	if result.DocOptions.ServeDocs {
		if result.DocOptions.specData, err = io.ReadAll(r); err == nil {
			useReader = bytes.NewReader(result.DocOptions.specData)
		}
	}
	if err == nil {
		d := json.NewDecoder(useReader)
		d.UseNumber()
		if err = d.Decode(result); err == nil {
			if useOptions.PathMiddlewares != nil {
				result.Middlewares = useOptions.PathMiddlewares("/")
				_ = result.WalkPaths(func(path string, pathDef *Path) (cont bool, err error) {
					pathDef.Middlewares = useOptions.PathMiddlewares(path)
					return true, nil
				})
			}
			err = result.WalkMethods(func(path string, method string, methodDef *Method) (contd bool, err error) {
				err = useOptions.setMethodHandler(path, method, methodDef)
				return true, err
			})
		}
	}
	return
}

// FromYaml builds a Definition from an existing OAS spec YAML
//
// where all methods in the spec must have a `x-handler` tag, e.g.
//
//	paths:
//	  "/":
//	    get:
//	      x-handler: "getRoot"
//
// where the key "getRoot" must be provided in the handlers arg
//
// or...
//
//	paths:
//	  "/":
//	    get:
//	      x-handler: ".GetRoot"
//
// where the "GetRoot" must be a http.HandlerFunc method on the supplied api arg
func FromYaml(r io.Reader, opts *FromOptions) (result *Definition, err error) {
	useOptions := defaultedFromOptions(opts, false)
	result = &Definition{
		DocOptions: *useOptions.DocOptions,
	}
	useReader := r
	if result.DocOptions.ServeDocs {
		if result.DocOptions.specData, err = io.ReadAll(r); err == nil {
			useReader = bytes.NewReader(result.DocOptions.specData)
		}
	}
	if err == nil {
		d := yaml.NewDecoder(useReader)
		if err = d.Decode(result); err == nil {
			if useOptions.PathMiddlewares != nil {
				result.Middlewares = useOptions.PathMiddlewares("/")
				_ = result.WalkPaths(func(path string, pathDef *Path) (cont bool, err error) {
					pathDef.Middlewares = useOptions.PathMiddlewares(path)
					return true, nil
				})
			}
			err = result.WalkMethods(func(path string, method string, methodDef *Method) (contd bool, err error) {
				err = useOptions.setMethodHandler(path, method, methodDef)
				return true, err
			})
		}
	}
	return
}

func (f *FromOptions) setMethodHandler(path string, method string, methodDef *Method) (err error) {
	if v, ok := methodDef.Extensions["handler"]; ok {
		if handlerName, ok := v.(string); ok {
			var hf http.HandlerFunc
			if hf, err = f.getMethodHandler(path, method, handlerName); err == nil {
				methodDef.Handler = hf
			}
		} else {
			err = fmt.Errorf("path '%s', method '%s' - 'x-handler' tag not a string", path, method)
		}
	} else if f.Strict {
		err = fmt.Errorf("path '%s', method '%s' - missing 'x-handler' tag", path, method)
	}
	return err
}

func (f *FromOptions) getMethodHandler(path string, method string, handlerName string) (http.HandlerFunc, error) {
	if strings.HasPrefix(handlerName, ".") && f.Api != nil {
		mfn := reflect.ValueOf(f.Api).MethodByName(handlerName[1:])
		if !mfn.IsValid() {
			return nil, fmt.Errorf("path '%s', method '%s' - 'x-handler' method '%s' does not exist", path, method, handlerName[1:])
		}
		if hf, ok := mfn.Interface().(func(http.ResponseWriter, *http.Request)); ok {
			return hf, nil
		}
		return nil, fmt.Errorf("path '%s', method '%s' - 'x-handler' method '%s' is not http.HandlerFunc", path, method, handlerName[1:])
	} else if f.Handlers != nil {
		if hf, ok := f.Handlers[handlerName]; ok && hf != nil {
			return hf, nil
		}
	}
	return nil, fmt.Errorf("path '%s', method '%s' - 'x-handler' no method or func found for '%s'", path, method, handlerName)
}

// Handlers is a lookup, by name, used by FromJson or FromYaml
type Handlers map[string]http.HandlerFunc

type PathMiddlewares func(path string) chi.Middlewares

// FromOptions is used by FromJson and FromYaml to control how the definition is
// built from a pre-existing OAS spec
type FromOptions struct {
	// DocOptions is the optional doc options for the generated Definition
	DocOptions *DocOptions
	// Api is the optional api object that provides methods where `x-handler` is specified as ".HandlerMethodName"
	Api any
	// Handlers is the optional look for handlers specified by `x-handler`
	Handlers Handlers
	// Strict when set, causes the FromJson / FromYaml to error if no `x-handler` tag is specified
	Strict bool
	// PathMiddlewares is an optional func that sets middlewares on paths found in the from spec
	PathMiddlewares PathMiddlewares
}

func defaultedFromOptions(opts *FromOptions, asJson bool) *FromOptions {
	if opts != nil {
		return &FromOptions{
			DocOptions:      defaultedDocsOptions(opts.DocOptions, asJson),
			Api:             opts.Api,
			Handlers:        opts.Handlers,
			Strict:          opts.Strict,
			PathMiddlewares: opts.PathMiddlewares,
		}
	}
	return &FromOptions{
		DocOptions: defaultedDocsOptions(nil, asJson),
		Strict:     true,
	}
}

func defaultedDocsOptions(opts *DocOptions, asJson bool) *DocOptions {
	if opts != nil {
		opts.AsJson = asJson
		return opts
	}
	return &DocOptions{
		AsJson: asJson,
	}
}
