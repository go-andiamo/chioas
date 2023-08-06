package chioas

import "github.com/go-andiamo/chioas/yaml"

type Servers map[string]Server

func (s Servers) writeYaml(w yaml.Writer) {
	if len(s) > 0 {
		w.WriteTagStart(tagNameServers)
		for url, svr := range s {
			svr.writeYaml(url, w)
		}
		w.WriteTagEnd()
	}
}

// Server represents the OAS definition of a server
type Server struct {
	// Description is the OAS description
	Description string
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
}

func (s Server) writeYaml(url string, w yaml.Writer) {
	w.WriteItemStart(tagNameUrl, url).
		WriteComments(s.Comment).
		WriteTagValue(tagNameDescription, s.Description)
	writeExtensions(s.Extensions, w)
	writeAdditional(s.Additional, s, w)
	w.WriteTagEnd()
}
