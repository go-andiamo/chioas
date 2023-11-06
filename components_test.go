package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComponents_WriteYaml(t *testing.T) {
	c := &Components{
		Schemas: Schemas{
			{
				Name: "test",
			},
		},
		Requests: CommonRequests{
			"foo": {
				Description: "foo request",
			},
		},
		Responses: CommonResponses{
			"bar": {
				Description: "bar response",
			},
		},
		Parameters: CommonParameters{
			"baz": {},
		},
		Examples: Examples{
			{
				Name: "foo",
			},
		},
		Additional: &testAdditional{},
		Extensions: Extensions{
			"foo": "bar",
		},
		Comment: "test comment",
	}
	w := yaml.NewWriter(nil)
	c.writeYaml(w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `#test comment
components:
  schemas:
    "test":
      type: object
  requestBodies:
    foo:
      description: "foo request"
      required: false
      content:
        "application/json":
          schema:
            type: object
  responses:
    bar:
      description: "bar response"
      content:
        "application/json":
          schema:
            type: object
  parameters:
    baz:
      name: baz
      in: query
      required: false
      schema:
        type: string
  examples:
    foo:
      value: null
  x-foo: bar
  foo: bar
`
	assert.Equal(t, expect, string(data))
}

func TestComponents_WriteYaml_ErrorOnDuplicates(t *testing.T) {
	c := &Components{
		Schemas: Schemas{
			{
				Name: "foo",
			},
			{
				Name: "foo",
			},
		},
	}
	w := yaml.NewWriter(nil)
	c.writeYaml(w)
	_, err := w.Bytes()
	assert.Error(t, err)

	c = &Components{
		Examples: Examples{
			{
				Name: "foo",
			},
			{
				Name: "foo",
			},
		},
	}
	w = yaml.NewWriter(nil)
	c.writeYaml(w)
	_, err = w.Bytes()
	assert.Error(t, err)
}
