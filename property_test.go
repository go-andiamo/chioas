package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProperty_WriteYaml(t *testing.T) {
	p := Property{
		Name:        "foo",
		Description: "foo desc",
		Type:        "number",
		Example:     1,
		Additional:  &testAdditional{},
		Comment:     "test comment",
	}
	w := yaml.NewWriter(nil)
	p.writeYaml(w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `"foo":
  #test comment
  description: "foo desc"
  type: "number"
  example: 1
  foo: "bar"
`
	assert.Equal(t, expect, string(data))
}

func TestProperty_WriteYaml_SchemaRef(t *testing.T) {
	p := Property{
		Name:        "foo",
		SchemaRef:   "foo",
		Description: "foo desc",
		Type:        "number",
		Example:     1,
	}
	w := yaml.NewWriter(nil)
	p.writeYaml(w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `"foo":
  $ref: "#/components/schemas/foo"
`
	assert.Equal(t, expect, string(data))
}

func TestProperty_WriteYaml_SchemaRefArray(t *testing.T) {
	p := Property{
		Name:      "foo",
		SchemaRef: "foo",
		Type:      "array",
	}
	w := yaml.NewWriter(nil)
	p.writeYaml(w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `"foo":
  type: "array"
  items:
    $ref: "#/components/schemas/foo"
`
	assert.Equal(t, expect, string(data))
}

func TestProperty_WriteYaml_DefaultType(t *testing.T) {
	p := Property{
		Name:        "foo",
		Description: "foo desc",
	}
	w := yaml.NewWriter(nil)
	p.writeYaml(w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `"foo":
  description: "foo desc"
  type: "string"
`
	assert.Equal(t, expect, string(data))
}
