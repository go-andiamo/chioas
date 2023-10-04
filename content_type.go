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
}

func writeContent(contentType string, schema any, schemaRef string, isArray bool, contentTypes ContentTypes, w yaml.Writer) {
	w.WriteTagStart(tagNameContent)
	writeContentType(contentType, schema, schemaRef, isArray, w)
	if contentTypes != nil {
		for k, v := range contentTypes {
			writeContentType(k, v.Schema, v.SchemaRef, v.IsArray, w)
		}
	}
	w.WriteTagEnd()
}

func writeContentType(contentType string, schema any, schemaRef string, isArray bool, w yaml.Writer) {
	w.WriteTagStart(defValue(contentType, tagNameApplicationJson))
	w.WriteTagStart(tagNameSchema)
	if schema != nil {
		writeSchema(schema, isArray, w)
	} else if schemaRef != "" {
		writeSchemaRef(schemaRef, isArray, w)
	} else {
		w.WriteTagValue(tagNameType, tagValueTypeObject)
	}
	w.WriteTagEnd().WriteTagEnd()
}
