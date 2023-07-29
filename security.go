package chioas

import "github.com/go-andiamo/chioas/yaml"

type SecuritySchemes []SecurityScheme

func (ss SecuritySchemes) writeYaml(w yaml.Writer, asSecurity bool) {
	if len(ss) > 0 {
		if asSecurity {
			for _, s := range ss {
				s.writeYaml(w, true)
			}
		} else {
			w.WriteTagStart(tagNameSecuritySchemes)
			for _, s := range ss {
				s.writeYaml(w, false)
			}
			w.WriteTagEnd()
		}
	}
}

type SecurityScheme struct {
	Name        string
	Description string
	Type        string
	Scheme      string
	ParamName   string
	In          string
	Scopes      []string
}

func (s SecurityScheme) writeYaml(w yaml.Writer, asSecurity bool) {
	if asSecurity {
		if (s.Type == "oauth2" || s.Type == "openIdConnect") && len(s.Scopes) > 0 {
			w.WriteLines("- " + s.Name + ":")
			for _, l := range s.Scopes {
				w.WriteLines(`  - "` + l + `"`)
			}
		} else {
			w.WriteLines("- " + s.Name + ": []")
		}
	} else {
		w.WriteTagStart(s.Name).
			WriteTagValue(tagNameDescription, s.Description).
			WriteTagValue(tagNameType, s.Type).
			WriteTagValue(tagNameScheme, s.Scheme).
			WriteTagValue(tagNameIn, s.In).
			WriteTagValue(tagNameName, s.ParamName).
			WriteTagEnd()
	}
}
