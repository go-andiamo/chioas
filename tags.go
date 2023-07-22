package chioas

import "github.com/go-andiamo/chioas/yaml"

type Tags []Tag

type Tag struct {
	Name         string
	Description  string
	ExternalDocs *ExternalDocs
	Additional   Additional
}

func (t Tag) writeYaml(w yaml.Writer) {
	w.WriteItemStart(tagNameName, t.Name).
		WriteTagValue(tagNameDescription, t.Description)
	if t.ExternalDocs != nil {
		t.ExternalDocs.writeYaml(w)
	}
	if t.Additional != nil {
		t.Additional.Write(t, w)
	}
	w.WriteTagEnd()
}

func defaultTag(parentTag string, tag string) string {
	if tag != "" {
		return tag
	}
	return parentTag
}
