package chioas

import "github.com/go-andiamo/chioas/yaml"

// Tags is an ordered collection of Tag
type Tags []Tag

// Tag represents the OAS definition of a tag
type Tag struct {
	// Name is the OAS name of the tag
	Name string
	// Description is the OAS description
	Description string
	// ExternalDocs is the optional OAS external docs
	ExternalDocs *ExternalDocs
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
}

func (t Tag) writeYaml(w yaml.Writer) {
	w.WriteItemStart(tagNameName, t.Name).
		WriteComments(t.Comment).
		WriteTagValue(tagNameDescription, t.Description)
	if t.ExternalDocs != nil {
		t.ExternalDocs.writeYaml(w)
	}
	writeExtensions(t.Extensions, w)
	writeAdditional(t.Additional, t, w)
	w.WriteTagEnd()
}

func defaultTag(parentTag string, tag string) string {
	if tag != "" {
		return tag
	}
	return parentTag
}
