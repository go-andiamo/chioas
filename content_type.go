package chioas

import "github.com/go-andiamo/chioas/yaml"

// ContentTypes is used by Response.AlternativeContentTypes and Request.AlternativeContentTypes to denote alternative content types
//
// The key is a media type - e.g. "application/json"
type ContentTypes map[string]ContentType

type ContentType struct {
	// Schema is the optional OAS Schema
	//
	// Only used if the value is non-nil - otherwise uses SchemaRef is used
	//
	// The value can be any of the following:
	//
	// * chioas.Schema (or *chioas.Schema)
	//
	// * a chioas.SchemaConverter
	//
	// * a chioas.SchemaWriter
	//
	// * a struct or ptr to struct (schema written is determined by examining struct fields)
	//
	// * a slice of structs (items schema written is determined by examining struct fields)
	Schema any
	// SchemaRef is the OAS schema reference
	//
	// Only used if value is a non-empty string - if both Schema is nil and SchemaRef is empty string, then an
	// empty object schema is written to the spec yaml, e.g.
	//   schema:
	//     type: "object"
	//
	// If the value does not contain a path (i.e. does not contain any "/") then the ref
	// path will be the value prefixed with components schemas path.  For example, specifying "foo"
	// will result in a schema ref:
	//   schema:
	//     $ref: "#/components/schemas/foo"
	SchemaRef string
	// IsArray indicates that request is an array of items
	IsArray bool
	// Examples is the ordered list of examples for the content type
	Examples Examples
}

func (ct ContentType) isArray() bool {
	return ct.IsArray
}

func (ct ContentType) schema() any {
	return ct.Schema
}

func (ct ContentType) schemaRef() string {
	return ct.SchemaRef
}

func (ct ContentType) examples() Examples {
	return ct.Examples
}

func (ct ContentType) alternatives() ContentTypes {
	panic("ContentType does not have alternatives")
}

type contentWritable interface {
	isArray() bool
	schema() any
	schemaRef() string
	examples() Examples
	alternatives() ContentTypes
}

func writeContent(contentType string, cw contentWritable, w yaml.Writer) {
	w.WriteTagStart(tagNameContent)
	writeContentType(contentType, cw, w)
	if alts := cw.alternatives(); alts != nil {
		for altCt, altCw := range alts {
			writeContentType(altCt, altCw, w)
		}
	}
	w.WriteTagEnd()
}

func writeContentType(contentType string, cw contentWritable, w yaml.Writer) {
	w.WriteTagStart(defValue(contentType, tagNameApplicationJson))
	w.WriteTagStart(tagNameSchema)
	isArray := cw.isArray()
	if schema := cw.schema(); schema != nil {
		writeSchema(schema, isArray, w)
	} else if schemaRef := cw.schemaRef(); schemaRef != "" {
		writeSchemaRef(schemaRef, isArray, w)
	} else {
		w.WriteTagValue(tagNameType, tagValueTypeObject)
	}
	w.WriteTagEnd()
	cw.examples().writeYaml(w)
	w.WriteTagEnd()
}
