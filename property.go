package chioas

import "github.com/go-andiamo/chioas/yaml"

type Property struct {
	Name        string
	Description string
	Type        string
	ItemType    string // only used if Type = "array"
	Example     any
	SchemaRef   string
	Additional  Additional
}

func (p Property) writeYaml(w yaml.Writer) {
	w.WriteTagStart("\"" + p.Name + "\"")
	if p.SchemaRef != "" {
		writeSchemaRef(p.SchemaRef, p.Type == tagValueTypeArray, w)
	} else if p.Type == tagValueTypeArray {
		w.WriteTagValue(tagNameDescription, p.Description).
			WriteTagValue(tagNameType, tagValueTypeArray).
			WriteTagStart(tagNameItems).
			WriteTagValue(tagNameType, defValue(p.ItemType, tagValueTypeString)).
			WriteTagValue(tagNameExample, p.Example).
			WriteTagEnd()
	} else {
		w.WriteTagValue(tagNameDescription, p.Description).
			WriteTagValue(tagNameType, defValue(p.Type, tagValueTypeString)).
			WriteTagValue(tagNameExample, p.Example)
	}
	if p.Additional != nil {
		p.Additional.Write(p, w)
	}
	w.WriteTagEnd()
}
