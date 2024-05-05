package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/go-andiamo/urit"
	"github.com/go-chi/chi/v5"
	"sort"
	"strings"
)

type Paths map[string]Path

func (ps Paths) writeYaml(opts *DocOptions, autoHeads bool, autoOptions bool, context string, w yaml.Writer) {
	for _, p := range ps.flattenAndSort() {
		p.writeYaml(opts, autoHeads, autoOptions, context, w)
	}
}

// Path represents a path for both the router and the OAS spec
type Path struct {
	// Methods is the methods on the path
	Methods Methods
	// Paths is the sub-paths of the path
	Paths Paths
	// Middlewares is any chi.Middlewares for the path
	Middlewares chi.Middlewares
	// ApplyMiddlewares is an optional function that returns chi.Middlewares for the path
	ApplyMiddlewares ApplyMiddlewares
	// Tag is the OAS tag of the path
	//
	// If this is an empty string and any ancestor Path.Tag is set then that ancestor tag is used
	//
	// The final tag is used by Method
	Tag string
	// PathParams is the OAS information about path params on this path
	//
	// Any path params introduced in the path are descended down the sub-paths and methods - any
	// path params that are not documented will still be seen in the OAS spec for methods
	PathParams PathParams
	// HideDocs if set to true, hides this path (and descendants) from docs
	HideDocs bool
	// Disabled is an optional DisablerFunc that, when called, returns whether this path is to be disabled
	//
	// When a path is disabled it does not appear in docs and is not registered on the router
	Disabled DisablerFunc
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
	// AutoOptionsMethod when set to true, automatically adds OPTIONS method for the path (and because Chioas knows the methods for each path can correctly set the Allow header)
	//
	// Note: If an OPTIONS method is already defined for the path then no OPTIONS method is automatically added
	AutoOptionsMethod bool
}

type flatPath struct {
	path     string
	ancestry []Path
	def      Path
	tag      string
}

func (p flatPath) writeYaml(opts *DocOptions, autoHeads bool, autoOptions bool, context string, w yaml.Writer) {
	if p.def.Methods != nil && p.def.Methods.hasVisibleMethods(opts) {
		template, err := urit.NewTemplate(p.path)
		if err != nil {
			w.SetError(err)
			return
		}
		w.WritePathStart(context, template.Template(true)).
			WriteComments(p.def.Comment)
		if p.def.Methods != nil {
			p.def.Methods.writeYaml(opts, autoHeads, autoOptions || p.def.AutoOptionsMethod, template, p.getPathParams(), p.tag, w)
		}
		writeExtensions(p.def.Extensions, w)
		writeAdditional(p.def.Additional, p.def, w)
		w.WriteTagEnd()
	}
}

func (p flatPath) getPathParams() PathParams {
	result := PathParams{}
	for _, a := range p.ancestry {
		if a.PathParams != nil {
			for k, pp := range a.PathParams {
				result[k] = pp
			}
		}
	}
	if p.def.PathParams != nil {
		for k, pp := range p.def.PathParams {
			result[k] = pp
		}
	}
	return result
}

func (ps Paths) flattenAndSort() []flatPath {
	result := pathsTraverse(nil, nil, nil, "", ps)
	sort.Slice(result, func(i, j int) bool {
		return result[i].path < result[j].path
	})
	return result
}

func pathsTraverse(collected []flatPath, parentPaths []string, ancestry []Path, tag string, paths Paths) []flatPath {
	if paths != nil {
		for path, def := range paths {
			hidden := def.HideDocs
			if !hidden && def.Disabled != nil {
				hidden = def.Disabled()
			}
			if !hidden {
				pps := append(parentPaths, path)
				useTag := defaultTag(tag, def.Tag)
				fp := flatPath{
					path:     strings.Join(pps, ""),
					ancestry: ancestry,
					def:      def,
					tag:      useTag,
				}
				collected = append(collected, fp)
				collected = pathsTraverse(collected, pps, append(ancestry, def), useTag, def.Paths)
			}
		}
	}
	return collected
}
