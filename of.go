package chioas

import (
	"github.com/go-andiamo/chioas/internal/tags"
	"github.com/go-andiamo/chioas/yaml"
)

type OfType uint

const (
	OneOf OfType = iota
	AnyOf
	AllOf
)

func (ot OfType) TagName() string {
	switch ot {
	case AllOf:
		return tags.AllOf
	case AnyOf:
		return tags.AnyOf
	}
	return tags.OneOf
}

// Ofs is a representation of OAS oneOf, allOf or anyOf
type Ofs struct {
	// OfType is the type - can be OneOf (default), AnyOf or AllOf
	OfType OfType
	// Of is the ordered slice of OfSchema items
	//
	// an OfSchema can be either a *Schema or an OfRef
	//
	// Note: the Ofs will not be written if the Of is empty!
	Of []OfSchema
}

func (ofs *Ofs) writeYaml(w yaml.Writer) {
	if len(ofs.Of) > 0 {
		w.WriteTagStart(ofs.OfType.TagName())
		for _, of := range ofs.Of {
			if of.IsRef() {
				writeItemRef(tags.Schemas, of.Ref(), w)
			} else {
				of.Schema().writeOfYaml(w)
			}
		}
		w.WriteTagEnd()
	}
}

type Of struct {
	SchemaRef string
	SchemaDef *Schema
}

func (o *Of) IsRef() bool {
	return o.SchemaRef != ""
}

func (o *Of) Ref() string {
	return o.SchemaRef
}

func (o *Of) Schema() *Schema {
	return o.SchemaDef
}

var _ OfSchema = &Of{}

type OfSchema interface {
	IsRef() bool
	Ref() string
	Schema() *Schema
}

type OfRef string

func (o OfRef) Ref() string {
	return string(o)
}

func (o OfRef) Schema() *Schema {
	panic("OfRef does not contain schema")
}

func (o OfRef) IsRef() bool {
	return true
}
