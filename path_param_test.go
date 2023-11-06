package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPathParam_WriteYaml(t *testing.T) {
	w := yaml.NewWriter(nil)
	pp := PathParam{
		Description: "test desc",
		Example:     "fooey",
		Additional:  &testAdditional{},
		Comment:     "test comment",
	}
	pp.writeYaml("foo", w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `- name: foo
  #test comment
  description: "test desc"
  in: path
  required: true
  example: fooey
  schema:
    type: string
  foo: bar
`
	assert.Equal(t, expect, string(data))
}

func TestPathParam_WriteYaml_Schema(t *testing.T) {
	w := yaml.NewWriter(nil)
	pp := PathParam{
		Description: "test desc",
		Schema: &Schema{
			Type: "object",
		},
	}
	pp.writeYaml("foo", w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `- name: foo
  description: "test desc"
  in: path
  required: true
  schema:
    type: object
`
	assert.Equal(t, expect, string(data))
}

func TestPathParam_WriteYaml_SchemaRef(t *testing.T) {
	w := yaml.NewWriter(nil)
	pp := PathParam{
		Description: "test desc",
		SchemaRef:   "foo",
	}
	pp.writeYaml("foo", w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `- name: foo
  description: "test desc"
  in: path
  required: true
  schema:
    $ref: "#/components/schemas/foo"
`
	assert.Equal(t, expect, string(data))
}

func TestPathParam_WriteYaml_Refd(t *testing.T) {
	w := yaml.NewWriter(nil)
	pp := PathParam{
		Ref:         "foo",
		Description: "won't see this",
		Example:     "won't see this either",
		Additional:  &testAdditional{},
		Comment:     "won't see",
	}
	pp.writeYaml("foo", w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `- $ref: "#/components/parameters/foo"
`
	assert.Equal(t, expect, string(data))
}
