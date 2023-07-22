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

type Schema struct {
	Name               string
	Description        string
	Type               string
	RequiredProperties []string
	Properties         []Property
	Default            any
	Enum               []any
	Additional         Additional
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
	if s.Additional != nil {
		s.Additional.Write(s, w)
	}
	if withName {
		w.WriteTagEnd()
	}
}
