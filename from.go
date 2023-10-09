package chioas

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"golang.org/x/exp/slices"
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
	useReader := r
	if useOptions.DocOptions.ServeDocs {
		if useOptions.DocOptions.specData, err = io.ReadAll(r); err == nil {
			useReader = bytes.NewReader(useOptions.DocOptions.specData)
		}
	}
	if err == nil {
		d := json.NewDecoder(useReader)
		d.UseNumber()
		obj := map[string]any{}
		if err = d.Decode(&obj); err == nil {
			result, err = useOptions.definitionFrom(obj)
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
	useReader := r
	if useOptions.DocOptions.ServeDocs {
		if useOptions.DocOptions.specData, err = io.ReadAll(r); err == nil {
			useReader = bytes.NewReader(useOptions.DocOptions.specData)
		}
	}
	if err == nil {
		d := yaml.NewDecoder(useReader)
		rootNode := &node{}
		if err = d.Decode(rootNode); err == nil {
			rootObj, ok := rootNode.Value.(map[string]any)
			if !ok {
				err = errors.New("bad yaml")
			} else {
				result, err = useOptions.definitionFrom(rootObj)
			}
		}
	}
	return
}

func (f *FromOptions) definitionFrom(obj map[string]any) (result *Definition, err error) {
	result = &Definition{
		DocOptions:  *f.DocOptions,
		Methods:     Methods{},
		Paths:       Paths{},
		Middlewares: f.getPathMiddlewares("/"),
	}
	if paths, ok := getMap(obj, tagNamePaths); ok {
		err = f.processPaths(result, paths)
	} else {
		err = errors.New("no paths defined")
	}
	return
}

func (f *FromOptions) processPaths(def *Definition, paths map[string]any) error {
	for path, rawV := range paths {
		if v, ok := rawV.(map[string]any); ok {
			if path == "/" {
				if err := f.processMethods(path, v, def.Methods); err != nil {
					return err
				}
			} else if pObj, err := f.getPath(def, path); err == nil {
				if err := f.processMethods(path, v, pObj.Methods); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			return fmt.Errorf("path '%s' not a map", path)
		}
	}
	return nil
}

func (f *FromOptions) getPath(def *Definition, path string) (result Path, err error) {
	parts := strings.Split(path, "/")
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if l := len(parts); l > 0 && !slices.Contains(parts, "") {
		curr := def.Paths
		for i, pt := range parts {
			if rp, ok := curr["/"+pt]; ok {
				result = rp
			} else {
				result = Path{
					Paths:       Paths{},
					Methods:     Methods{},
					Middlewares: f.getPathMiddlewares("/" + strings.Join(parts[:i+1], "/")),
				}
				curr["/"+pt] = result
			}
			if i == l-1 {
				return
			}
			curr = result.Paths
		}
	} else {
		err = fmt.Errorf("invalid path '%s'", path)
	}
	return
}

func (f *FromOptions) getPathMiddlewares(path string) chi.Middlewares {
	if f.PathMiddlewares != nil {
		return f.PathMiddlewares(path)
	}
	return nil
}

var chiMethods = map[string]bool{
	http.MethodConnect: true,
	http.MethodDelete:  true,
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodOptions: true,
	http.MethodPatch:   true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodTrace:   true,
}

func (f *FromOptions) processMethods(path string, obj map[string]any, methods Methods) error {
	for k, rawV := range obj {
		m := strings.ToUpper(k)
		if chiMethods[m] {
			if v, ok := rawV.(map[string]any); ok {
				if method, err := f.getMethod(path, m, v); err == nil && method != nil {
					methods[m] = *method
				} else if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("method '%s' on path '%s' not a map", m, path)
			}
		}
	}
	return nil
}

func (f *FromOptions) getMethod(path string, method string, m map[string]any) (result *Method, err error) {
	if rawV, ok := m["x-handler"]; ok {
		if handlerName, ok := rawV.(string); ok {
			var hf http.HandlerFunc
			if hf, err = f.getMethodHandler(path, method, handlerName); err == nil {
				result = &Method{Handler: hf}
			}
		} else {
			err = fmt.Errorf("path '%s', method '%s' - 'x-handler' tag not a string", path, method)
		}
	} else if f.Strict {
		err = fmt.Errorf("path '%s', method '%s' - missing 'x-handler' tag", path, method)
	}
	return
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

func getMap(obj map[string]any, tagName string) (map[string]any, bool) {
	if raw, ok := obj[tagName]; ok {
		if m, ok := raw.(map[string]any); ok {
			return m, true
		}
	}
	return nil, false
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
