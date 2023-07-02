package chioas

type Tags []Tag

type Tag struct {
	Name        string
	Description string
}

func (t Tag) writeYaml(w *yamlWriter) {
	w.writeItemStart(tagNameName, t.Name)
	w.writeTagValue(tagNameDescription, t.Description)
	w.writeTagEnd()
}

func defaultTag(parentTag string, tag string) string {
	if tag != "" {
		return tag
	}
	return parentTag
}
