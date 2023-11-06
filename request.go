package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
)

// Request represents the OAS definition of a request
type Request struct {
	// Ref is the OAS $ref name for the request
	//
	// If this is a non-empty string and the response is used by Method.Request, then a $ref to "#/components/requestBodies/" is used
	//
	// If the Request is used by Components.Requests this value is ignored
	Ref string
	// Description is the OAS description
	Description string
	// Required is the OAS flag determining if the request is required
	Required bool
	// ContentType is the OAS content type
	//
	// defaults to "application/json"
	ContentType string
	// AlternativeContentTypes is a map of alternative content types (where the key is the media type - e.g. "application/json")
	AlternativeContentTypes ContentTypes
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
	// Examples is the ordered list of examples for the request
	Examples Examples
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
	// Comment is any comment(s) to appear in the OAS spec yaml (not used with Ref)
	Comment string
}

func (r *Request) isArray() bool {
	return r.IsArray
}

func (r *Request) schema() any {
	return r.Schema
}

func (r *Request) schemaRef() string {
	return r.SchemaRef
}

func (r *Request) examples() Examples {
	return r.Examples
}

func (r *Request) alternatives() ContentTypes {
	return r.AlternativeContentTypes
}

func (r *Request) writeYaml(w yaml.Writer) {
	w.WriteTagStart(tagNameRequestBody)
	if r.Ref == "" {
		w.WriteComments(r.Comment).
			WriteTagValue(tagNameDescription, r.Description).
			WriteTagValue(tagNameRequired, r.Required)
		writeContent(r.ContentType, r, w)
		writeExtensions(r.Extensions, w)
		writeAdditional(r.Additional, r, w)
	} else {
		writeRef(tagNameRequestBodies, r.Ref, w)
	}
	w.WriteTagEnd()
}

func (r *Request) componentsWriteYaml(name string, w yaml.Writer) {
	w.WriteTagStart(name).
		WriteTagValue(tagNameDescription, r.Description).
		WriteTagValue(tagNameRequired, r.Required)
	writeContent(r.ContentType, r, w)
	writeExtensions(r.Extensions, w)
	writeAdditional(r.Additional, r, w)
	w.WriteTagEnd()
}
