package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCommonParameters_WriteYaml(t *testing.T) {
	w := yaml.NewWriter(nil)
	ps := CommonParameters{
		"Fooey": {
			Name:     "foo",
			Required: true,
		},
	}
	ps.writeYaml(w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `parameters:
  Fooey:
    name: "foo"
    in: "query"
    required: true
`
	assert.Equal(t, expect, string(data))
}

func TestCommonParameter_WriteYaml(t *testing.T) {
	testCases := []struct {
		param  CommonParameter
		expect string
	}{
		{
			param: CommonParameter{
				Name: "foo",
			},
			expect: `name: "foo"
  in: "query"
  required: false
`,
		},
		{
			param: CommonParameter{
				Name:        "foo",
				Description: "foo param",
			},
			expect: `name: "foo"
  description: "foo param"
  in: "query"
  required: false
`,
		},
		{
			param: CommonParameter{
				Name:        "foo",
				Description: "foo param",
				Required:    true,
			},
			expect: `name: "foo"
  description: "foo param"
  in: "query"
  required: true
`,
		},
		{
			param: CommonParameter{
				Name:        "foo",
				Description: "foo param",
				Example:     "foo example",
				Required:    true,
			},
			expect: `name: "foo"
  description: "foo param"
  in: "query"
  required: true
  example: "foo example"
`,
		},
		{
			param: CommonParameter{
				Name:      "foo",
				SchemaRef: "FooRef",
			},
			expect: `name: "foo"
  in: "query"
  required: false
  schema:
    $ref: "#/components/schemas/FooRef"
`,
		},
		{
			param: CommonParameter{
				Name: "foo",
				Schema: &Schema{
					Name: "won't see this",
					Type: "string",
				},
			},
			expect: `name: "foo"
  in: "query"
  required: false
  schema:
    type: "string"
`,
		},
		{
			param: CommonParameter{
				Name:       "foo",
				Additional: &testAdditional{},
			},
			expect: `name: "foo"
  in: "query"
  required: false
  foo: "bar"
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			tc.param.writeYaml("Test", w)
			data, err := w.Bytes()
			require.NoError(t, err)
			assert.Equal(t, "Test:\n  "+tc.expect, string(data))
		})
	}
}
