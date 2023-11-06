package chioas

import (
	"errors"
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
)

type Examples []Example

func (egs Examples) writeYaml(w yaml.Writer) {
	if len(egs) > 0 {
		// check for duplicates...
		m := map[string]struct{}{}
		for _, eg := range egs {
			if eg.Name != "" {
				if _, ok := m[eg.Name]; ok {
					w.SetError(fmt.Errorf("duplicate example name '%s' in components", eg.Name))
					return
				}
				m[eg.Name] = struct{}{}
			}
		}
		w.WriteTagStart(tagNameExamples)
		for _, eg := range egs {
			eg.writeYaml(w)
		}
		w.WriteTagEnd()
	}
}

type Example struct {
	// Name is the name of the example (must be specified)
	Name string
	// Summary is the OAS summary of the example
	Summary string
	// Description is the OAS description of the example
	Description string
	// Value os the OAS value for the example
	Value any
	// ExampleRef is the OAS example reference
	//
	// Only used if value is a non-empty string
	//
	// If the value does not contain a path (i.e. does not contain any "/") then the ref
	// path will be the value prefixed with components example path.  For example, specifying "foo"
	// will result in a schema ref:
	//   "my example":
	//     $ref: "#/components/examples/foo"
	ExampleRef string
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
	// Comment is any comment(s) to appear in the OAS spec yaml (not used with Ref)
	Comment string
}

func (eg *Example) writeYaml(w yaml.Writer) {
	if eg.Name == "" {
		w.SetError(errors.New("example without name"))
		return
	}
	w.WriteTagStart(eg.Name).
		WriteComments(eg.Comment)
	if eg.ExampleRef != "" {
		writeRef(tagNameExamples, eg.ExampleRef, w)
	} else {
		w.WriteTagValue(tagNameSummary, eg.Summary).
			WriteTagValue(tagNameDescription, eg.Description)
		if eg.Value != nil {
			w.WriteTagValue(tagNameValue, eg.Value)
		} else {
			w.WriteTagValue(tagNameValue, yaml.LiteralValue{Value: "null"})
		}
		writeExtensions(eg.Extensions, w)
		writeAdditional(eg.Additional, eg, w)
	}
	w.WriteTagEnd()
}
