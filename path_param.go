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
	// Schema is the optional OAS Schema
	Schema *Schema
	// SchemaRef is the OAS schema reference
	//
	// Only used if value is a non-empty string - if both Schema is nil and SchemaRef is empty string, then an
	// empty object schema is written to the spec yaml, e.g.
	//   schema:
	//     type: "string"
	//
	// If the value does not contain a path (i.e. does not contain any "/") then the ref
	// path will be the value prefixed with components schemas path.  For example, specifying "foo"
	// will result in a schema ref:
	//   schema:
	//     $ref: "#/components/schemas/foo"
	SchemaRef string
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
		w.WriteTagStart(tagNameSchema)
		if pp.Schema != nil {
			pp.Schema.writeYaml(false, w)
		} else if pp.SchemaRef != "" {
			writeSchemaRef(pp.SchemaRef, false, w)
		} else {
			w.WriteTagValue(tagNameType, "string")
		}
		w.WriteTagEnd()
		writeExtensions(pp.Extensions, w)
		writeAdditional(pp.Additional, pp, w)
		w.WriteTagEnd()
	} else {
		writeItemRef(tagNameParameters, pp.Ref, w)
	}
}
