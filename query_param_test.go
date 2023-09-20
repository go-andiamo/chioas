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
  schema:
    type: "string"
- name: "bar"
  in: "query"
  required: true
  schema:
    type: "string"
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
  schema:
    type: "string"
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
  schema:
    type: "string"
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
  schema:
    type: "string"
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
  schema:
    type: "string"
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
  schema:
    type: "string"
  foo: "bar"
`,
		},
		{
			queryParam: QueryParam{
				Name:    "won't see this",
				Ref:     "foo",
				Comment: "won't see this",
			},
			expect: `- $ref: "#/components/parameters/foo"
`,
		},
		{
			queryParam: QueryParam{
				Name:    "foo",
				Comment: "test comment",
			},
			expect: `- name: "foo"
  #test comment
  in: "query"
  required: false
  schema:
    type: "string"
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
