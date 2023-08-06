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
	// Comment is any comment(s) to appear in the OAS spec yaml (not used with Ref)
	Comment string
	// Ref is the OAS $ref name for the parameter
	//
	// If this is a non-empty string, then a $ref to "#/components/parameters/" is used
	Ref string
}

func (pp PathParam) writeYaml(name string, w yaml.Writer) {
	if pp.Ref == "" {
		w.WriteItemStart(tagNameName, name).
			WriteComments(pp.Comment).
			WriteTagValue(tagNameDescription, pp.Description).
			WriteTagValue(tagNameIn, tagValuePath).
			WriteTagValue(tagNameRequired, true).
			WriteTagValue(tagNameExample, pp.Example)
		writeExtensions(pp.Extensions, w)
		writeAdditional(pp.Additional, pp, w)
		w.WriteTagEnd()
	} else {
		w.WriteItemValue(tagNameRef, refPathParameters+pp.Ref)
	}
}
