package chioas

import "github.com/go-andiamo/chioas/yaml"

// Properties is an ordered collection of Property
type Properties []Property

// Property represents the OAS definition of a property
type Property struct {
	// Name is the OAS name of the property
	Name string
	// Description is the OAS description of the property
	Description string
	// Type is the OAS type of the property
	//
	// Should be one of "string", "object", "array", "boolean", "integer", "number" or "null"
	Type string
	// ItemType is the OAS type of array items
	//
	// only used if Type = "array"
	ItemType string
	// Example is the OAS example for the property
	Example any
	// SchemaRef is the OAS schema reference
	//
	// Only used if value is a non-empty string
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
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
}

func (p Property) writeYaml(w yaml.Writer) {
	w.WriteTagStart("\"" + p.Name + "\"").
		WriteComments(p.Comment)
	if p.SchemaRef != "" {
		writeSchemaRef(p.SchemaRef, p.Type == tagValueTypeArray, w)
	} else if p.Type == tagValueTypeArray {
		w.WriteTagValue(tagNameDescription, p.Description).
			WriteTagValue(tagNameType, tagValueTypeArray).
			WriteTagStart(tagNameItems).
			WriteTagValue(tagNameType, defValue(p.ItemType, tagValueTypeString)).
			WriteTagValue(tagNameExample, p.Example).
			WriteTagEnd()
	} else {
		w.WriteTagValue(tagNameDescription, p.Description).
			WriteTagValue(tagNameType, defValue(p.Type, tagValueTypeString)).
			WriteTagValue(tagNameExample, p.Example)
	}
	writeExtensions(p.Extensions, w)
	writeAdditional(p.Additional, p, w)
	w.WriteTagEnd()
}
