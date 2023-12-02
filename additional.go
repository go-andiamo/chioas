package chioas

import "github.com/go-andiamo/chioas/yaml"

// AdditionalOasProperties is a type (map) of additional OAS properties that can be used for
// use on any .Additional property
type AdditionalOasProperties map[string]any

func (ap AdditionalOasProperties) Write(on any, w yaml.Writer) {
	for k, v := range ap {
		w.WriteTagValue(k, v)
	}
}

// Additional is an interface that can be supplied to many parts of the definition
// to write additional yaml to the OAS
type Additional interface {
	Write(on any, w yaml.Writer)
}

func writeAdditional(a Additional, on any, w yaml.Writer) {
	if a != nil {
		a.Write(on, w)
	}
}
