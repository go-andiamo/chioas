package chioas

import "github.com/go-andiamo/chioas/yaml"

type PathParams map[string]PathParam

type PathParam struct {
	Description string
	Example     any
	Additional  Additional
}

func (pp PathParam) writeYaml(name string, w yaml.Writer) {
	w.WriteItemStart(tagNameName, name).
		WriteTagValue(tagNameDescription, pp.Description).
		WriteTagValue(tagNameIn, tagValuePath).
		WriteTagValue(tagNameRequired, true).
		WriteTagValue(tagNameExample, pp.Example)
	if pp.Additional != nil {
		pp.Additional.Write(pp, w)
	}
	w.WriteTagEnd()
}
