package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"net/http"
	"sort"
	"strconv"
)

// Responses is a map of Response where the key is the http status code
type Responses map[int]Response

func (r Responses) writeYaml(isHead bool, w yaml.Writer) {
	if l := len(r); l > 0 {
		sortCodes := make([]int, 0, l)
		for rc := range r {
			sortCodes = append(sortCodes, rc)
		}
		sort.Ints(sortCodes)
		w.WriteTagStart(tagNameResponses)
		for _, sc := range sortCodes {
			r[sc].writeYaml(sc, isHead, w)
		}
		w.WriteTagEnd()
	}
}

// Response is the OAS definition of a response
type Response struct {
	// Ref is the OAS $ref name for the response
	//
	// If this is a non-empty string and the response is used by Method.Responses, then a $ref to "#/components/responses/" is used
	//
	// If the Response is used by Components.ResponseBodies this value is ignored
	Ref string
	// Description is the OAS description
	Description string
	// NoContent indicates that this response has not content
	//
	// This does not eed to set this when status code is http.StatusNoContent
	NoContent bool
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
	// Examples is the ordered list of examples for the response
	Examples Examples
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
	// Comment is any comment(s) to appear in the OAS spec yaml (not used with Ref)
	Comment string
}

func (r Response) isArray() bool {
	return r.IsArray
}

func (r Response) schema() any {
	return r.Schema
}

func (r Response) schemaRef() string {
	return r.SchemaRef
}

func (r Response) examples() Examples {
	return r.Examples
}

func (r Response) alternatives() ContentTypes {
	return r.AlternativeContentTypes
}

func (r Response) writeYaml(statusCode int, isHead bool, w yaml.Writer) {
	w.WriteTagStart(strconv.Itoa(statusCode))
	if r.Ref == "" {
		w.WriteComments(r.Comment)
		desc := r.Description
		if desc == "" {
			desc = http.StatusText(statusCode)
		}
		w.WriteTagValue(tagNameDescription, desc)
		if !isHead && !r.NoContent && statusCode != http.StatusNoContent {
			writeContent(r.ContentType, r, w)
		}
		writeExtensions(r.Extensions, w)
		writeAdditional(r.Additional, r, w)
	} else {
		writeRef(tagNameResponses, r.Ref, w)
	}
	w.WriteTagEnd()
}

func (r Response) componentsWriteYaml(name string, w yaml.Writer) {
	w.WriteTagStart(name)
	w.WriteTagValue(tagNameDescription, r.Description)
	if !r.NoContent {
		writeContent(r.ContentType, r, w)
	}
	writeExtensions(r.Extensions, w)
	writeAdditional(r.Additional, r, w)
	w.WriteTagEnd()
}
