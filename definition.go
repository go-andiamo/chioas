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
	DocOptions      DocOptions
	AutoHeadMethods bool // set to true to automatically add HEAD methods for GET methods (and where HEAD method not explicitly specified)
	Info            Info
	Servers         Servers
	Tags            Tags
	Methods         Methods         // methods on api root
	Middlewares     chi.Middlewares // chi middlewares for api root
	Paths           Paths           // descendant paths
	Components      *Components
	Additional      Additional
}

// SetupRoutes sets up the API routes on the supplied chi.Router
//
// Pass the thisApi arg if any of the methods use method by name
func (d *Definition) SetupRoutes(router chi.Router, thisApi any) error {
	subRoute := chi.NewRouter()
	subRoute.Use(d.Middlewares...)
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
			subRoute.Use(pDef.Middlewares...)
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
	if d.Additional != nil {
		d.Additional.Write(d, w)
	}
	return w.Errored()
}
