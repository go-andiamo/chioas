package chioas

import (
	"bufio"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

type Definition struct {
	This        any
	Context     string
	Title       string
	Description string
	Version     string
	Tags        Tags
	Methods     Methods
	Paths       Paths
}

func (d *Definition) SetupRoutes(router chi.Router) {
	router.Route("/", func(r chi.Router) {
		d.setupMethods(d.Methods, r)
		d.setupPaths(d.Paths, r)
	})
}

func (d *Definition) setupPaths(paths Paths, route chi.Router) {
	if paths != nil {
		for p, pDef := range paths {
			route.Route(p, func(r chi.Router) {
				d.setupMethods(pDef.Methods, r)
				d.setupPaths(pDef.Paths, r)
			})
		}
	}
}

func (d *Definition) setupMethods(methods Methods, route chi.Router) {
	if methods != nil {
		for m, mDef := range methods {
			route.MethodFunc(m, "/", d.getHandler(mDef))
		}
	}
}

func (d *Definition) getHandler(method Method) http.HandlerFunc {
	return method.getHandler(d.This)
}

func (d *Definition) WriteYaml(w io.Writer) error {
	yw := newYamlWriter(bufio.NewWriter(w))
	return d.writeYaml(yw)
}

func (d *Definition) AsYaml() ([]byte, error) {
	w := newYamlWriter(nil)
	_ = d.writeYaml(w)
	return w.bytes()
}

func (d *Definition) writeYaml(w *yamlWriter) error {
	w.writeTagValue(tagNameOpenApi, OasVersion)
	w.writeTagStart(tagNameInfo)
	w.writeTagValue(tagNameTitle, d.Title)
	w.writeTagValue(tagNameDescription, d.Description)
	w.writeTagValue(tagNameVersion, d.Version)
	w.writeTagEnd()
	if d.Tags != nil && len(d.Tags) > 0 {
		w.writeTagStart(tagNameTags)
		for _, t := range d.Tags {
			t.writeYaml(w)
		}
		w.writeTagEnd()
	}
	w.writeTagStart(tagNamePaths)
	if d.Methods != nil && len(d.Methods) > 0 {
		w.writePathStart(d.Context, "/")
		d.Methods.writeYaml("", nil, "", w)
		w.writeTagEnd()
	}
	if d.Paths != nil {
		d.Paths.writeYaml(d.Context, w)
	}
	w.writeTagEnd()
	return w.err
}
