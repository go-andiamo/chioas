package chioas

type Tags []Tag

type Tag struct {
	Name        string
	Description string
}

func (t Tag) writeYaml(w *yamlWriter) {
	w.writeItemStart("name", t.Name)
	w.writeTagValue("description", t.Description)
	w.writeTagEnd()
}

func defaultTag(parentTag string, tag string) string {
	if tag != "" {
		return tag
	}
	return parentTag
}
