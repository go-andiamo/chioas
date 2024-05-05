package chioas

import (
	"bufio"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strings"
)

// Definition is the overall definition of an api
type Definition struct {
	// DocOptions is the documentation options for the spec
	DocOptions DocOptions
	// Info is the OAS info for the spec
	Info Info
	// Servers is the OAS servers for the spec
	Servers Servers
	// Tags is the OAS tags for the spec
	Tags Tags
	// Methods is any methods on the root api
	Methods Methods
	// Middlewares is any chi.Middlewares for api root
	Middlewares chi.Middlewares
	// ApplyMiddlewares is an optional function that returns chi.Middlewares for api root
	ApplyMiddlewares ApplyMiddlewares
	// Paths is the api paths to be setup (each path can have sub-paths)
	Paths Paths // descendant paths
	// Components is the OAS components
	Components *Components
	// Security is the OAS security for the api
	Security SecuritySchemes
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
	// AutoHeadMethods when set to true, automatically adds HEAD methods for GET methods (where HEAD method not explicitly specified)
	//
	// If you don't want these automatically added HEAD methods to appear in the OAS spec - then set DocOptions.HideAutoOptionsMethods
	AutoHeadMethods bool
	// AutoOptionsMethods when set to true, automatically adds OPTIONS methods for each path (and because Chioas knows the methods for each path can correctly set the Allow header)
	//
	// Note: If an OPTIONS method is already defined for the path then no OPTIONS method is automatically added
	AutoOptionsMethods bool
	// OptionsMethodPayloadBuilder is an optional implementation of OptionsMethodPayloadBuilder that can provide body payloads for the automatically created OPTIONS methods
	OptionsMethodPayloadBuilder OptionsMethodPayloadBuilder
	// RootAutoOptionsMethod when set to true, automatically adds OPTIONS method for the root path (and because Chioas knows the methods for each path can correctly set the Allow header)
	//
	// Note: If an OPTIONS method is already defined for the root path then no OPTIONS method is automatically added
	RootAutoOptionsMethod bool
	// AutoMethodNotAllowed when set to true, automatically adds a method not allowed (405) handler for each path (and because Chioas knows the methods for each path can correctly set the Allow header)
	AutoMethodNotAllowed bool
	// MethodHandlerBuilder is an optional MethodHandlerBuilder which is called to build the
	// http.HandlerFunc for the method
	//
	// If MethodHandlerBuilder is nil then the default method handler builder is used
	MethodHandlerBuilder MethodHandlerBuilder
}

// SetupRoutes sets up the API routes on the supplied chi.Router
//
// Pass the thisApi arg if any of the methods use method by name
func (d *Definition) SetupRoutes(router chi.Router, thisApi any) error {
	if err := d.DocOptions.SetupRoutes(d, router); err != nil {
		return err
	}
	subRoute := chi.NewRouter()
	middlewares := d.Middlewares
	if d.ApplyMiddlewares != nil {
		middlewares = append(middlewares, d.ApplyMiddlewares(thisApi)...)
	}
	subRoute.Use(middlewares...)
	if err := d.setupMethods(root, nil, d.Methods, d.RootAutoOptionsMethod, subRoute, thisApi); err != nil {
		return err
	}
	if err := d.setupPaths(nil, d.Paths, subRoute, thisApi); err != nil {
		return err
	}
	if d.AutoMethodNotAllowed {
		subRoute.MethodNotAllowed(d.methodNotAllowedHandler(d.Methods))
	}
	router.Mount(root, subRoute)
	return nil
}

func (d *Definition) setupPaths(ancestry []string, paths Paths, route chi.Router, thisApi any) error {
	if paths != nil {
		for p, pDef := range paths {
			disabled := false
			if pDef.Disabled != nil {
				disabled = pDef.Disabled()
			}
			if !disabled {
				newAncestry := append(ancestry, p)
				subRoute := chi.NewRouter()
				middlewares := pDef.Middlewares
				if pDef.ApplyMiddlewares != nil {
					middlewares = append(middlewares, pDef.ApplyMiddlewares(thisApi)...)
				}
				if d.AutoMethodNotAllowed {
					subRoute.MethodNotAllowed(d.methodNotAllowedHandler(pDef.Methods))
				}
				subRoute.Use(middlewares...)
				if err := d.setupMethods(strings.Join(newAncestry, ""), &pDef, pDef.Methods, d.AutoOptionsMethods || pDef.AutoOptionsMethod, subRoute, thisApi); err != nil {
					return err
				}
				if err := d.setupPaths(newAncestry, pDef.Paths, subRoute, thisApi); err != nil {
					return err
				}
				route.Mount(p, subRoute)
			}
		}
	}
	return nil
}

func (d *Definition) setupMethods(path string, pathDef *Path, methods Methods, pathAutoOptions bool, route chi.Router, thisApi any) error {
	if methods != nil && len(methods) > 0 {
		for m, mDef := range methods {
			if h, err := getMethodHandlerBuilder(d.MethodHandlerBuilder).BuildHandler(path, m, mDef, thisApi); err == nil {
				route.MethodFunc(m, root, h)
			} else {
				return err
			}
		}
		if (d.AutoOptionsMethods || pathAutoOptions) && !methods.hasOptions() {
			route.MethodFunc(http.MethodOptions, root, d.optionsHandler(methods, path, pathDef))
		}
		if d.AutoHeadMethods {
			if mDef, ok := methods.getWithoutHead(); ok {
				h, _ := getMethodHandlerBuilder(d.MethodHandlerBuilder).BuildHandler(path, http.MethodHead, mDef, thisApi)
				route.MethodFunc(http.MethodHead, root, h)
			}
		}
	} else if pathAutoOptions {
		route.MethodFunc(http.MethodOptions, root, d.optionsHandler(Methods{}, path, pathDef))
	}
	return nil
}

func (d *Definition) optionsHandler(methods Methods, path string, pathDef *Path) http.HandlerFunc {
	add := []string{http.MethodOptions}
	if _, hasGet := methods.getWithoutHead(); hasGet && d.AutoHeadMethods {
		add = append(add, http.MethodHead)
	}
	allow := strings.Join(methods.sorted(add...), ", ")
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set(hdrAllow, allow)
		payloadData := make([]byte, 0)
		if d.OptionsMethodPayloadBuilder != nil {
			var addHdrs map[string]string
			payloadData, addHdrs = d.OptionsMethodPayloadBuilder.BuildPayload(path, pathDef, d)
			for k, v := range addHdrs {
				writer.Header().Set(k, v)
			}
		}
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(payloadData)
	}
}

func (d *Definition) methodNotAllowedHandler(methods Methods) http.HandlerFunc {
	add := make([]string, 0)
	if _, hasGet := methods.getWithoutHead(); hasGet && d.AutoHeadMethods {
		add = append(add, http.MethodHead)
	}
	if d.AutoOptionsMethods {
		add = append(add, http.MethodOptions)
	}
	allow := strings.Join(methods.sorted(add...), ", ")
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set(hdrAllow, allow)
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// WriteYaml writes the definition as YAML to the provided io.Writer
func (d *Definition) WriteYaml(w io.Writer) error {
	yw := yaml.NewWriter(bufio.NewWriter(w))
	err := d.writeYaml(yw)
	if err == nil {
		err = yw.Flush()
	}
	return err
}

// WriteJson writes the definition as JSON to the provided io.Writer
func (d *Definition) WriteJson(writer io.Writer) (err error) {
	w := yaml.NewWriter(nil)
	_ = d.writeYaml(w)
	var data []byte
	if data, err = w.Bytes(); err == nil {
		if data, err = yaml2Json(data); err == nil {
			_, err = writer.Write(data)
		}
	}
	return
}

// AsYaml returns the spec as YAML data
func (d *Definition) AsYaml() ([]byte, error) {
	w := yaml.NewWriter(nil)
	_ = d.writeYaml(w)
	return w.Bytes()
}

// AsJson returns the spec as JSON data
func (d *Definition) AsJson() (data []byte, err error) {
	w := yaml.NewWriter(nil)
	_ = d.writeYaml(w)
	if data, err = w.Bytes(); err == nil {
		data, err = yaml2Json(data)
	}
	return
}

func (d *Definition) writeYaml(w yaml.Writer) error {
	if d.DocOptions.CheckRefs {
		w.RefChecker(d)
	}
	w.WriteComments(d.Comment)
	w.WriteTagValue(tagNameOpenApi, OasVersion)
	d.Info.writeYaml(w)
	if d.Servers != nil {
		d.Servers.writeYaml(w)
	}
	if d.Tags != nil && len(d.Tags) > 0 {
		w.WriteTagStart(tagNameTags)
		for _, t := range d.Tags {
			t.writeYaml(w)
		}
		w.WriteTagEnd()
	}
	w.WriteTagStart(tagNamePaths)
	if d.Methods != nil && len(d.Methods) > 0 {
		w.WritePathStart(d.DocOptions.Context, root)
		d.Methods.writeYaml(&d.DocOptions, d.AutoHeadMethods, d.AutoOptionsMethods, nil, nil, "", w)
		w.WriteTagEnd()
	}
	if d.Paths != nil {
		d.Paths.writeYaml(&d.DocOptions, d.AutoHeadMethods, d.AutoOptionsMethods, d.DocOptions.Context, w)
	}
	w.WriteTagEnd()
	if d.Components != nil {
		d.Components.writeYaml(w)
	}
	if len(d.Security) > 0 {
		w.WriteTagStart(tagNameSecurity)
		d.Security.writeYaml(w, true)
		w.WriteTagEnd()
	}
	writeExtensions(d.Extensions, w)
	writeAdditional(d.Additional, d, w)
	return w.Errored()
}
