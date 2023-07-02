package chioas

import (
	"sort"
	"strings"
)

type Paths map[string]Path

type Path struct {
	Methods    Methods
	Paths      Paths
	Tag        string
	PathParams PathParams
}

func (ps Paths) writeYaml(context string, w *yamlWriter) {
	for _, p := range ps.flattenAndSort() {
		p.writeYaml(context, w)
	}
}

type flatPath struct {
	path     string
	ancestry []Path
	def      Path
	tag      string
}

func (p flatPath) writeYaml(context string, w *yamlWriter) {
	w.writePathStart(context, p.path)
	if p.def.Methods != nil {
		p.def.Methods.writeYaml(p.tag, w)
	}
	w.writeTagEnd()
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
	return collected
}
