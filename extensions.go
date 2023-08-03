package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"strings"
)

// Extensions can be added to many OAS items and are written as `x-` yaml properties
type Extensions map[string]any

const extensionPrefix = "x-"

func writeExtensions(e Extensions, w yaml.Writer) {
	if e != nil {
		e.writeYaml(w)
	}
}

func (e Extensions) writeYaml(w yaml.Writer) {
	for k, v := range e {
		key := k
		if !strings.HasPrefix(key, extensionPrefix) {
			key = extensionPrefix + key
		}
		w.WriteTagValue(key, v)
	}
}
