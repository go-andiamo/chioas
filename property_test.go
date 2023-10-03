package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProperty_WriteYaml(t *testing.T) {
	testCases := []struct {
		property Property
		expect   string
	}{
		{
			property: Property{
				Name:        "foo",
				Description: "foo desc",
				Type:        "number",
				Example:     1,
				Additional:  &testAdditional{},
				Comment:     "test comment",
			},
			expect: `"foo":
  #test comment
  description: "foo desc"
  type: "number"
  example: 1
  foo: "bar"
`,
		},
		{
			property: Property{
				Name:        "foo",
				SchemaRef:   "foo",
				Description: "foo desc",
				Type:        "number",
				Example:     1,
			},
			expect: `"foo":
  $ref: "#/components/schemas/foo"
`,
		},
		{
			property: Property{
				Name:      "foo",
				SchemaRef: "foo",
				Type:      "array",
			},
			expect: `"foo":
  type: "array"
  items:
    $ref: "#/components/schemas/foo"
`,
		},
		{
			property: Property{
				Name:        "foo",
				Description: "foo desc",
			},
			expect: `"foo":
  description: "foo desc"
  type: "string"
`,
		},
		{
			property: Property{
				Name:   "foo",
				Format: "email",
				Example: map[string]any{
					"foo": "bar",
				},
			},
			expect: `"foo":
  type: "string"
  example:
    foo: bar
  format: "email"
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			tc.property.writeYaml(w)
			data, err := w.Bytes()
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
		})
	}
}
