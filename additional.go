package chioas

import "github.com/go-andiamo/chioas/yaml"

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
