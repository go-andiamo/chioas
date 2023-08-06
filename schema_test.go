package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchemas_WriteYaml(t *testing.T) {
	w := yaml.NewWriter(nil)

	s := Schemas{
		{
			Name:               "test",
			Description:        "test desc",
			RequiredProperties: []string{"foo"},
		},
	}
	s.writeYaml(w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `schemas:
  "test":
    description: "test desc"
    type: "object"
    required:
      - "foo"
`
	assert.Equal(t, expect, string(data))
}

func TestSchema_WriteYaml(t *testing.T) {
	testCases := []struct {
		schema   Schema
		withName bool
		expect   string
	}{
		{
			expect: `type: "object"
`,
		},
		{
			withName: true,
			expect: `"":
  type: "object"
`,
		},
		{
			schema: Schema{
				Name:               "test",
				Description:        "test desc",
				RequiredProperties: []string{"foo"},
				Properties: []Property{
					{
						Name: "foo",
					},
					{
						Name: "bar",
					},
				},
			},
			expect: `description: "test desc"
type: "object"
required:
  - "foo"
properties:
  "foo":
    type: "string"
  "bar":
    type: "string"
`,
		},
		{
			schema: Schema{
				Name:               "test",
				Type:               "array",
				Description:        "test desc",
				RequiredProperties: []string{"foo"},
				Properties: []Property{
					{
						Name: "foo",
					},
					{
						Name: "bar",
					},
				},
			},
			expect: `description: "test desc"
type: "array"
required:
  - "foo"
properties:
  "foo":
    type: "string"
  "bar":
    type: "string"
`,
		},
		{
			schema: Schema{
				Type: "string",
				Enum: []any{"foo", "bar", 0},
			},
			expect: `type: "string"
enum:
  - "foo"
  - "bar"
  - 0
`,
		},
		{
			schema: Schema{
				Name:       "test",
				Additional: &testAdditional{},
			},
			withName: true,
			expect: `"test":
  type: "object"
  foo: "bar"
`,
		},
		{
			schema: Schema{
				Name:    "test",
				Comment: "test comment",
			},
			expect: `#test comment
type: "object"
`,
		},
		{
			schema: Schema{
				Name:    "test",
				Comment: "test comment",
			},
			withName: true,
			expect: `"test":
  #test comment
  type: "object"
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			tc.schema.writeYaml(tc.withName, w)
			data, err := w.Bytes()
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
		})
	}
}
