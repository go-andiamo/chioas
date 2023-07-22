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

type Server struct {
	Description string
	Additional  Additional
}

func (s Server) writeYaml(url string, w yaml.Writer) {
	w.WriteItemStart(tagNameUrl, url).
		WriteTagValue(tagNameDescription, s.Description)
	if s.Additional != nil {
		s.Additional.Write(s, w)
	}
	w.WriteTagEnd()
}
