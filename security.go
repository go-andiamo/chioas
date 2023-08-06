package chioas

import "github.com/go-andiamo/chioas/yaml"

// SecuritySchemes is an ordered collection of SecurityScheme
//
// Used by:
//
// * Components.SecuritySchemes to define the security schemes used by the api
//
// * Definition.Security to define the security requirements across the entire api
//
// * Method.Security to define the security requirement for a particular method
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

// SecurityScheme represents the OAS definition of a security scheme
type SecurityScheme struct {
	// Name is the OAS name of the security scheme
	Name string
	// Description is the OAS description
	Description string
	// Type is the OAS security scheme type
	//
	// Valid values are: "apiKey", "http", "mutualTLS", "oauth2", "openIdConnect"
	//
	// Defaults to "http"
	Type string
	// Scheme is the OAS HTTP Authorization scheme
	Scheme string
	// ParamName is the OAS param name (in query, header or cookie)
	ParamName string
	// In is the OAS definition of where the API key param is found
	//
	// Valid values are: "query", "header" or "cookie"
	In string
	// Scopes is the security requirement scopes
	//
	// Only used when the SecurityScheme is part of Definition.Security and the
	// Type is either "oauth2" or "openIdConnect"
	Scopes []string
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
}

func (s SecurityScheme) writeYaml(w yaml.Writer, asSecurity bool) {
	if asSecurity {
		if (s.Type == "oauth2" || s.Type == "openIdConnect") && len(s.Scopes) > 0 {
			w.WriteItemStart(s.Name, nil)
			for _, l := range s.Scopes {
				w.WriteItem(l)
			}
			w.WriteTagEnd()
		} else {
			w.WriteItemValue(s.Name, yaml.LiteralValue{Value: "[]"})
		}
	} else {
		w.WriteTagStart(s.Name).
			WriteComments(s.Comment).
			WriteTagValue(tagNameDescription, s.Description).
			WriteTagValue(tagNameType, defValue(s.Type, "http")).
			WriteTagValue(tagNameScheme, s.Scheme).
			WriteTagValue(tagNameIn, s.In).
			WriteTagValue(tagNameName, s.ParamName)
		writeExtensions(s.Extensions, w)
		writeAdditional(s.Additional, s, w)
		w.WriteTagEnd()
	}
}
