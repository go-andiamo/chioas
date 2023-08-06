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
	subRoute := chi.NewRouter()
	middlewares := d.Middlewares
	if d.ApplyMiddlewares != nil {
		middlewares = append(middlewares, d.ApplyMiddlewares(thisApi)...)
	}
	subRoute.Use(middlewares...)
	if err := d.DocOptions.setupRoutes(d, subRoute); err != nil {
		return err
	}
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
			if h, err := mDef.getHandler(path, m, thisApi); err == nil {
				route.MethodFunc(m, root, h)
			} else {
				return err
			}
		}
		if d.AutoHeadMethods {
			if getM, ok := methods.getWithoutHead(); ok {
				h, _ := getM.getHandler(path, http.MethodHead, thisApi)
				route.MethodFunc(http.MethodHead, root, h)
			}
		}
	}
	return nil
}

func (d *Definition) WriteYaml(w io.Writer) error {
	yw := yaml.NewWriter(bufio.NewWriter(w))
	err := d.writeYaml(yw)
	if err == nil {
		err = yw.Flush()
	}
	return err
}

func (d *Definition) AsYaml() ([]byte, error) {
	w := yaml.NewWriter(nil)
	_ = d.writeYaml(w)
	return w.Bytes()
}

func (d *Definition) writeYaml(w yaml.Writer) error {
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
		d.Methods.writeYaml(&d.DocOptions, d.AutoHeadMethods, nil, nil, "", w)
		w.WriteTagEnd()
	}
	if d.Paths != nil {
		d.Paths.writeYaml(&d.DocOptions, d.AutoHeadMethods, d.DocOptions.Context, w)
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
