package chioas

import (
	"errors"
	"github.com/go-andiamo/chioas/yaml"
)

// Discriminator is a representation of the OAS discriminator object
type Discriminator struct {
	// PropertyName is the OAS property name for the discriminator
	PropertyName string
	// Mapping holds mappings between payload values (of the specified property) and schema names or references
	Mapping map[string]string
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
}

func (d *Discriminator) writeYaml(w yaml.Writer) {
	if d.PropertyName == "" {
		w.SetError(errors.New("discriminator without property name"))
		return
	}
	w.WriteTagStart(tagNameDiscriminator).
		WriteComments(d.Comment).
		WriteTagValue(tagNamePropertyName, d.PropertyName)
	if d.Mapping != nil && len(d.Mapping) > 0 {
		w.WriteTagStart(tagNameMapping)
		for k, v := range d.Mapping {
			w.WriteTagValue(k, refCheck(tagNameSchemas, v, w))
		}
		w.WriteTagEnd()
	}
	writeExtensions(d.Extensions, w)
	writeAdditional(d.Additional, d, w)
	w.WriteTagEnd()
}
