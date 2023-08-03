package chioas

import "github.com/go-andiamo/chioas/yaml"

// PathParams is a map of PathParam where the key is the param name
type PathParams map[string]PathParam

// PathParam represents the OAS definition of a path param
type PathParam struct {
	// Description is the OAS description
	Description string
	// Example is the OAS example
	Example any
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
}

func (pp PathParam) writeYaml(name string, w yaml.Writer) {
	w.WriteItemStart(tagNameName, name).
		WriteTagValue(tagNameDescription, pp.Description).
		WriteTagValue(tagNameIn, tagValuePath).
		WriteTagValue(tagNameRequired, true).
		WriteTagValue(tagNameExample, pp.Example)
	writeExtensions(pp.Extensions, w)
	writeAdditional(pp.Additional, pp, w)
	w.WriteTagEnd()
}
