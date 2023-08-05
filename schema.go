package chioas

import "github.com/go-andiamo/chioas/yaml"

type Schemas []Schema

func (ss Schemas) writeYaml(w yaml.Writer) {
	if len(ss) > 0 {
		w.WriteTagStart(tagNameSchemas)
		for _, s := range ss {
			s.writeYaml(true, w)
		}
		w.WriteTagEnd()
	}
}

type SchemaConverter interface {
	ToSchema() *Schema
}

type SchemaWriter interface {
	WriteSchema(w yaml.Writer)
}

// Schema represents the OAS definition of a schema
type Schema struct {
	// Name is the OAS name of the schema
	Name string
	// Description is the OAS description
	Description string
	// Type is the OAS type
	//
	// Should be one of "string", "object", "array", "boolean", "integer", "number" or "null"
	Type string
	// RequiredProperties is the ordered collection of required properties
	RequiredProperties []string
	// Properties is the ordered collection of properties
	Properties Properties
	// Default is the OAS default
	Default any
	// Enum is the OAS enum
	Enum []any
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
}

func (s Schema) writeYaml(withName bool, w yaml.Writer) {
	if withName {
		w.WriteTagStart("\"" + s.Name + "\"")
	}
	w.WriteTagValue(tagNameDescription, s.Description)
	if s.Type != "" {
		w.WriteTagValue(tagNameType, s.Type)
	} else {
		w.WriteTagValue(tagNameType, tagValueTypeObject)
	}
	if len(s.RequiredProperties) > 0 {
		w.WriteTagStart(tagNameRequired)
		for _, rp := range s.RequiredProperties {
			w.WriteItem(rp)
		}
		w.WriteTagEnd()
	}
	if len(s.Properties) > 0 {
		w.WriteTagStart(tagNameProperties)
		for _, p := range s.Properties {
			p.writeYaml(w)
		}
		w.WriteTagEnd()
	}
	w.WriteTagValue(tagNameDefault, s.Default)
	if len(s.Enum) > 0 {
		w.WriteTagStart(tagNameEnum)
		for _, e := range s.Enum {
			w.WriteItem(e)
		}
		w.WriteTagEnd()
	}
	writeExtensions(s.Extensions, w)
	writeAdditional(s.Additional, s, w)
	if withName {
		w.WriteTagEnd()
	}
}
