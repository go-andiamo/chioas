package chioas

import "github.com/go-andiamo/chioas/yaml"

// QueryParams represents an ordered collection of QueryParam
type QueryParams []QueryParam

// QueryParam represents the OAS definition of a query param (or header param)
type QueryParam struct {
	// Name is the name of the param
	Name string
	// Description is the OAS description
	Description string
	// Required is the OAS required flag
	Required bool
	// In is the OAS field defining whether the param is a "query" or "header" param
	//
	// Defaults to "query"
	In string
	// Example is the OAS example for the param
	Example any
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

func (qp QueryParams) writeYaml(w yaml.Writer) {
	for _, p := range qp {
		p.writeYaml(w)
	}
}

func (p QueryParam) writeYaml(w yaml.Writer) {
	if p.Ref == "" {
		w.WriteItemStart(tagNameName, p.Name).
			WriteComments(p.Comment).
			WriteTagValue(tagNameDescription, p.Description).
			WriteTagValue(tagNameIn, defValue(p.In, tagValueQuery)).
			WriteTagValue(tagNameRequired, p.Required).
			WriteTagValue(tagNameExample, p.Example)
		w.WriteTagStart(tagNameSchema)
		if p.Schema != nil {
			p.Schema.writeYaml(false, w)
		} else if p.SchemaRef != "" {
			writeSchemaRef(p.SchemaRef, false, w)
		} else {
			w.WriteTagValue(tagNameType, "string")
		}
		w.WriteTagEnd()
		writeExtensions(p.Extensions, w)
		writeAdditional(p.Additional, p, w)
		w.WriteTagEnd()
	} else {
		writeItemRef(tagNameParameters, p.Ref, w)
	}
}
