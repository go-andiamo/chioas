package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/go-andiamo/urit"
	"github.com/go-chi/chi/v5"
	"sort"
	"strings"
)

type Paths map[string]Path

func (ps Paths) writeYaml(opts *DocOptions, autoHeads bool, context string, w yaml.Writer) {
	for _, p := range ps.flattenAndSort() {
		p.writeYaml(opts, autoHeads, context, w)
	}
}

type ApplyMiddlewares func(thisApi any) chi.Middlewares

type Path struct {
	Methods          Methods
	Paths            Paths
	Middlewares      chi.Middlewares // chi middlewares for path
	ApplyMiddlewares ApplyMiddlewares
	Tag              string
	PathParams       PathParams
	HideDocs         bool // hides this path (and descendants) from docs
}

type flatPath struct {
	path     string
	ancestry []Path
	def      Path
	tag      string
}

func (p flatPath) writeYaml(opts *DocOptions, autoHeads bool, context string, w yaml.Writer) {
	if p.def.Methods != nil && p.def.Methods.hasVisibleMethods(opts) {
		template, err := urit.NewTemplate(p.path)
		if err != nil {
			w.SetError(err)
			return
		}
		w.WritePathStart(context, template.Template(true))
		if p.def.Methods != nil {
			p.def.Methods.writeYaml(opts, autoHeads, template, p.getPathParams(), p.tag, w)
		}
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
			if !def.HideDocs {
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
