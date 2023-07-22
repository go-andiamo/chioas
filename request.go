package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
)

type Request struct {
	Description string
	Required    bool
	ContentType string // defaults to "application/json"
	Schema      any
	SchemaRef   string
	IsArray     bool // indicates that Schema or SchemaRef is an array of items
	Additional  Additional
}

func (r *Request) writeYaml(w yaml.Writer) {
	w.WriteTagStart(tagNameRequestBody).
		WriteTagValue(tagNameDescription, r.Description).
		WriteTagValue(tagNameRequired, r.Required)
	writeContent(r.ContentType, r.Schema, r.SchemaRef, r.IsArray, w)
	if r.Additional != nil {
		r.Additional.Write(r, w)
	}
	w.WriteTagEnd()
}
