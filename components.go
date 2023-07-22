package chioas

import "github.com/go-andiamo/chioas/yaml"

type Components struct {
	Schemas    Schemas
	Additional Additional
}

func (c *Components) writeYaml(w yaml.Writer) {
	w.WriteTagStart(tagNameComponents)
	c.Schemas.writeYaml(w)
	if c.Additional != nil {
		c.Additional.Write(c, w)
	}
	w.WriteTagEnd()
}
