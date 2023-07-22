package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRequest_WriteYaml(t *testing.T) {
	w := yaml.NewWriter(nil)
	r := &Request{
		Description: "desc",
		Required:    true,
		SchemaRef:   "foo",
		Additional:  &testAdditional{},
	}

	r.writeYaml(w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `requestBody:
  description: "desc"
  required: true
  content:
    application/json:
      schema:
        $ref: "#/components/schemas/foo"
  foo: "bar"
`
	assert.Equal(t, expect, string(data))
}

func TestRequest_WriteYaml_IsArray(t *testing.T) {
	w := yaml.NewWriter(nil)
	r := &Request{
		Description: "desc",
		Required:    true,
		SchemaRef:   "foo",
		IsArray:     true,
	}

	r.writeYaml(w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `requestBody:
  description: "desc"
  required: true
  content:
    application/json:
      schema:
        type: "array"
        items:
          $ref: "#/components/schemas/foo"
`
	assert.Equal(t, expect, string(data))
}
