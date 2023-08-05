package chioas

import "github.com/go-andiamo/chioas/yaml"

// CommonRequests is a map of Request, where the key is the name (that can be referenced by Request.Ref)
type CommonRequests map[string]Request

// CommonResponses is a map of Response, where the key is the name (that can be referenced by Response.Ref)
type CommonResponses map[string]Response

// Components represents the OAS components
type Components struct {
	// Schemas is the OAS reusable schemas
	Schemas Schemas
	// Requests is the OAS reusable requests
	//
	// To reference one of these, use Method.Request.Ref with the name
	Requests CommonRequests
	// Responses is the OAS reusable responses
	//
	// To reference one of these, use Method.Responses.Ref with the name
	Responses CommonResponses
	// Parameters is the OAS reusable parameters
	Parameters CommonParameters
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
	if c.Requests != nil {
		c.Requests.writeYaml(w)
	}
	if c.Responses != nil {
		c.Responses.writeYaml(w)
	}
	if c.Parameters != nil {
		c.Parameters.writeYaml(w)
	}
	c.SecuritySchemes.writeYaml(w, false)
	writeExtensions(c.Extensions, w)
	writeAdditional(c.Additional, c, w)
	w.WriteTagEnd()
}

func (r CommonRequests) writeYaml(w yaml.Writer) {
	if len(r) > 0 {
		w.WriteTagStart(tagNameRequestBodies)
		for name, rr := range r {
			rr.componentsWriteYaml(name, w)
		}
		w.WriteTagEnd()
	}
}

func (r CommonResponses) writeYaml(w yaml.Writer) {
	if len(r) > 0 {
		w.WriteTagStart(tagNameResponses)
		for name, rr := range r {
			rr.componentsWriteYaml(name, w)
		}
		w.WriteTagEnd()
	}
}
