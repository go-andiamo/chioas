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
	// AutoHeadMethods when set to true, automatically adds HEAD methods for GET methods (where HEAD method not explicitly specified)
	AutoHeadMethods bool
	// AutoOptionsMethods when set to true, automatically adds OPTIONS methods for each path (and because Chioas knows the methods for each path can correctly set the Allow header)
	//
	// Note: If an OPTIONS method is already defined for the path then no OPTIONS method is automatically added
	AutoOptionsMethods bool
	// MethodHandlerBuilder is an optional MethodHandlerBuilder which is called to build the
	// http.HandlerFunc for the method
	//
	// If MethodHandlerBuilder is nil then the default method handler builder is used
	MethodHandlerBuilder MethodHandlerBuilder
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
	if err := d.setupMethods(root, d.Methods, subRoute, thisApi); err != nil {
		return err
	}
	if err := d.setupPaths(nil, d.Paths, subRoute, thisApi); err != nil {
		return err
	}
	router.Mount(root, subRoute)
	return nil
}

func (d *Definition) setupPaths(ancestry []string, paths Paths, route chi.Router, thisApi any) error {
	if paths != nil {
		for p, pDef := range paths {
			newAncestry := append(ancestry, p)
			subRoute := chi.NewRouter()
			middlewares := pDef.Middlewares
			if pDef.ApplyMiddlewares != nil {
				middlewares = append(middlewares, pDef.ApplyMiddlewares(thisApi)...)
			}
			subRoute.Use(middlewares...)
			if err := d.setupMethods(strings.Join(newAncestry, ""), pDef.Methods, subRoute, thisApi); err != nil {
				return err
			}
			if err := d.setupPaths(newAncestry, pDef.Paths, subRoute, thisApi); err != nil {
				return err
			}
			route.Mount(p, subRoute)
		}
	}
	return nil
}

func (d *Definition) setupMethods(path string, methods Methods, route chi.Router, thisApi any) error {
	if methods != nil {
		for m, mDef := range methods {
			if h, err := getMethodHandlerBuilder(d.MethodHandlerBuilder).BuildHandler(path, m, mDef, thisApi); err == nil {
				route.MethodFunc(m, root, h)
			} else {
				return err
			}
		}
		if d.AutoOptionsMethods && !methods.hasOptions() {
			route.MethodFunc(http.MethodOptions, root, optionsHandler(methods, d.AutoHeadMethods))
		}
		if d.AutoHeadMethods {
			if mDef, ok := methods.getWithoutHead(); ok {
				h, _ := getMethodHandlerBuilder(d.MethodHandlerBuilder).BuildHandler(path, http.MethodHead, mDef, thisApi)
				route.MethodFunc(http.MethodHead, root, h)
			}
		}
	}
	return nil
}

func optionsHandler(methods Methods, addHead bool) http.HandlerFunc {
	add := []string{http.MethodOptions}
	if addHead {
		add = append(add, http.MethodHead)
	}
	allow := strings.Join(methods.sorted(add...), ", ")
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set(hdrAllow, allow)
		writer.WriteHeader(http.StatusOK)
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
