package chioas

import "github.com/go-andiamo/chioas/yaml"

type QueryParams []QueryParam

type QueryParam struct {
	Name        string
	Description string
	Required    bool
	In          string // defaults to "query" (use "query" or "header")
	Example     any
	Schema      *Schema
	SchemaRef   string
	Additional  Additional
}

func (qp QueryParams) writeYaml(w yaml.Writer) {
	for _, p := range qp {
		p.writeYaml(w)
	}
}

func (p QueryParam) writeYaml(w yaml.Writer) {
	w.WriteItemStart(tagNameName, p.Name).
		WriteTagValue(tagNameDescription, p.Description).
		WriteTagValue(tagNameIn, defValue(p.In, tagValueQuery)).
		WriteTagValue(tagNameRequired, p.Required).
		WriteTagValue(tagNameExample, p.Example)
	if p.Schema != nil {
		w.WriteTagStart(tagNameSchema)
		p.Schema.writeYaml(false, w)
		w.WriteTagEnd()
	} else if p.SchemaRef != "" {
		w.WriteTagStart(tagNameSchema)
		writeSchemaRef(p.SchemaRef, false, w)
		w.WriteTagEnd()
	}
	if p.Additional != nil {
		p.Additional.Write(p, w)
	}
	w.WriteTagEnd()
}
