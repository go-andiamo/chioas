package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueryParams_WriteYaml(t *testing.T) {
	w := yaml.NewWriter(nil)
	ps := QueryParams{
		{
			Name: "foo",
		},
		{
			Name:     "bar",
			Required: true,
		},
	}
	ps.writeYaml(w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `- name: "foo"
  in: "query"
  required: false
- name: "bar"
  in: "query"
  required: true
`
	assert.Equal(t, expect, string(data))
}

func TestQueryParam_WriteYaml(t *testing.T) {
	testCases := []struct {
		queryParam QueryParam
		expect     string
	}{
		{
			queryParam: QueryParam{
				Name: "foo",
			},
			expect: `- name: "foo"
  in: "query"
  required: false
`,
		},
		{
			queryParam: QueryParam{
				Name:        "foo",
				Description: "foo param",
			},
			expect: `- name: "foo"
  description: "foo param"
  in: "query"
  required: false
`,
		},
		{
			queryParam: QueryParam{
				Name:        "foo",
				Description: "foo param",
				Required:    true,
			},
			expect: `- name: "foo"
  description: "foo param"
  in: "query"
  required: true
`,
		},
		{
			queryParam: QueryParam{
				Name:        "foo",
				Description: "foo param",
				Example:     "foo example",
				Required:    true,
			},
			expect: `- name: "foo"
  description: "foo param"
  in: "query"
  required: true
  example: "foo example"
`,
		},
		{
			queryParam: QueryParam{
				Name:      "foo",
				SchemaRef: "FooRef",
			},
			expect: `- name: "foo"
  in: "query"
  required: false
  schema:
    $ref: "#/components/schemas/FooRef"
`,
		},
		{
			queryParam: QueryParam{
				Name: "foo",
				Schema: &Schema{
					Name: "won't see this",
					Type: "string",
				},
			},
			expect: `- name: "foo"
  in: "query"
  required: false
  schema:
    type: "string"
`,
		},
		{
			queryParam: QueryParam{
				Name:       "foo",
				Additional: &testAdditional{},
			},
			expect: `- name: "foo"
  in: "query"
  required: false
  foo: "bar"
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			tc.queryParam.writeYaml(w)
			data, err := w.Bytes()
			require.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
		})
	}
}