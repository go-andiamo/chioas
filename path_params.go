package chioas

import "github.com/go-andiamo/urit"

type PathParams map[string]PathParam

type PathParam struct {
	Description string
	Example     any
}

func (p PathParams) writeYaml(path string, w *yamlWriter) {
	if path == "" {
		return
	}
	pt, err := urit.NewTemplate(path)
	if err != nil {
		w.err = err
		return
	}
	pvs := pt.Vars()
	if len(pvs) > 0 {
		w.writeTagStart(tagNameParameters)
		for _, pv := range pvs {
			if pp, ok := p[pv.Name]; ok {
				pp.writeYaml(pv.Name, w)
			} else {
				(PathParam{}).writeYaml(pv.Name, w)
			}
		}
		w.writeTagEnd()
	}
}

func (pp PathParam) writeYaml(name string, w *yamlWriter) {
	w.writeItemStart(tagNameName, name)
	w.writeTagValue(tagNameDescription, pp.Description)
	w.writeTagValue(tagNameIn, "path")
	w.writeTagValue(tagNameRequired, true)
	w.writeTagValue(tagNameExample, pp.Example)
	w.writeTagEnd()
}
