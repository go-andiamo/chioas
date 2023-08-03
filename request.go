package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
)

// Request represents the OAS definition of a request
type Request struct {
	// Description is the OAS description
	Description string
	// Required is the OAS flag determining if the request is required
	Required bool
	// ContentType is the OAS content type
	//
	// defaults to "application/json"
	ContentType string
	// Schema is the optional OAS Schema
	Schema any
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
	// IsArray indicates that request is an array of items
	IsArray bool
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
}

func (r *Request) writeYaml(w yaml.Writer) {
	w.WriteTagStart(tagNameRequestBody).
		WriteTagValue(tagNameDescription, r.Description).
		WriteTagValue(tagNameRequired, r.Required)
	writeContent(r.ContentType, r.Schema, r.SchemaRef, r.IsArray, w)
	writeExtensions(r.Extensions, w)
	writeAdditional(r.Additional, r, w)
	w.WriteTagEnd()
}
