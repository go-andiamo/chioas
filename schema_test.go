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
    type: object
    required:
      - foo
`
	assert.Equal(t, expect, string(data))
}

func TestSchema_WriteYaml(t *testing.T) {
	testCases := []struct {
		schema    Schema
		withName  bool
		expect    string
		expectErr string
	}{
		{
			expect: `type: object
`,
		},
		{
			withName:  true,
			expectErr: "schema without name",
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
type: object
required:
  - foo
properties:
  "foo":
    type: string
  "bar":
    type: string
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
type: array
required:
  - foo
properties:
  "foo":
    type: string
  "bar":
    type: string
`,
		},
		{
			schema: Schema{
				Type: "string",
				Enum: []any{"foo", "bar", 0},
			},
			expect: `type: string
enum:
  - foo
  - bar
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
  type: object
  foo: bar
`,
		},
		{
			schema: Schema{
				Name:    "test",
				Comment: "test comment",
			},
			expect: `#test comment
type: object
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
  type: object
`,
		},
		{
			schema: Schema{
				Name: "test",
				Example: map[string]any{
					"foo": "bar",
				},
			},
			withName: true,
			expect: `"test":
  type: object
  example:
    foo: bar
`,
		},
		{
			schema: Schema{
				Name: "test",
				Properties: Properties{
					{
						Name:     "foo",
						Required: true,
					},
				},
			},
			expect: `type: object
required:
  - foo
properties:
  "foo":
    type: string
`,
		},
		{
			schema: Schema{
				Name:               "test",
				RequiredProperties: []string{"foo", "foo"},
				Properties: Properties{
					{
						Name:     "foo",
						Required: true,
					},
				},
			},
			expect: `type: object
required:
  - foo
properties:
  "foo":
    type: string
`,
		},
		{
			schema: Schema{
				Name:      "Foo",
				SchemaRef: "Foo2",
			},
			withName: true,
			expect: `"Foo":
  $ref: "#/components/schemas/Foo2"
`,
		},
		{
			schema: Schema{
				Name: "Foo",
				Discriminator: &Discriminator{
					PropertyName: "foo",
				},
			},
			withName: true,
			expect: `"Foo":
  type: object
  discriminator:
    propertyName: foo
`,
		},
		{
			schema: Schema{
				Name: "Foo",
				Ofs:  &Ofs{},
			},
			withName: true,
			expect: `"Foo":
  type: object
`,
		},
		{
			schema: Schema{
				Name: "Foo",
				Ofs: &Ofs{
					Of: []OfSchema{
						OfRef("foo"),
						OfRef("bar"),
					},
				},
			},
			withName: true,
			expect: `"Foo":
  type: object
  oneOf:
    - $ref: "#/components/schemas/foo"
    - $ref: "#/components/schemas/bar"
`,
		},
		{
			schema: Schema{
				Type:   "string",
				Format: "uuid",
			},
			expect: `type: string
format: uuid
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			tc.schema.writeYaml(tc.withName, w)
			data, err := w.Bytes()
			if tc.expectErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.expectErr, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expect, string(data))
			}
		})
	}
}
