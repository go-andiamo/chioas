package chioas

import "github.com/go-andiamo/chioas/yaml"

// Components represents the OAS components
type Components struct {
	// Schemas is the OAS common schemas
	Schemas Schemas
	// SecuritySchemes is the OAS security schemes
	SecuritySchemes SecuritySchemes
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
}

func (c *Components) writeYaml(w yaml.Writer) {
	w.WriteTagStart(tagNameComponents)
	c.Schemas.writeYaml(w)
	c.SecuritySchemes.writeYaml(w, false)
	writeExtensions(c.Extensions, w)
	writeAdditional(c.Additional, c, w)
	w.WriteTagEnd()
}
