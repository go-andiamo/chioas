package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/internal/values"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOf(t *testing.T) {
	var refOf OfSchema = OfRef("my-ref")
	var schemaOf OfSchema = &Schema{}
	assert.True(t, refOf.IsRef())
	assert.False(t, schemaOf.IsRef())
	assert.Equal(t, "my-ref", refOf.Ref())
	assert.NotNil(t, schemaOf.Schema())
	assert.Panics(t, func() {
		_ = refOf.Schema()
	})
	assert.Panics(t, func() {
		_ = schemaOf.Ref()
	})
}

func TestOf_WriteYaml(t *testing.T) {
	testCases := []struct {
		ofs    *Ofs
		expect string
	}{
		{
			ofs: &Ofs{
				Of: []OfSchema{
					OfRef("my-schema"),
				},
			},
			expect: `oneOf:
  - $ref: "#/components/schemas/my-schema"
`,
		},
		{
			ofs: &Ofs{
				OfType: OneOf,
				Of: []OfSchema{
					OfRef("my-schema"),
				},
			},
			expect: `oneOf:
  - $ref: "#/components/schemas/my-schema"
`,
		},
		{
			ofs: &Ofs{
				Of: []OfSchema{
					&Schema{},
				},
			},
			expect: `oneOf:
  - type: object
`,
		},
		{
			ofs: &Ofs{
				Of: []OfSchema{

					&Schema{
						RequiredProperties: []string{"foo"},
						Properties: Properties{
							{
								Name: "foo",
								Type: values.TypeObject,
								Properties: Properties{
									{
										Name: "bar",
										Type: values.TypeBoolean,
									},
								},
							},
						},
					},
				},
			},
			expect: `oneOf:
  - type: object
    required:
      - foo
    properties:
      "foo":
        type: object
        properties:
          "bar":
            type: boolean
`,
		},
		{
			ofs: &Ofs{
				Of: []OfSchema{
					OfRef("my-schema1"),
					OfRef("my-schema2"),
				},
			},
			expect: `oneOf:
  - $ref: "#/components/schemas/my-schema1"
  - $ref: "#/components/schemas/my-schema2"
`,
		},
		{
			ofs: &Ofs{
				OfType: AllOf,
				Of: []OfSchema{
					OfRef("my-schema"),
					&Schema{
						RequiredProperties: []string{"foo"},
						Properties: Properties{
							{
								Name: "foo",
								Type: values.TypeObject,
								Properties: Properties{
									{
										Name: "bar",
										Type: values.TypeBoolean,
									},
								},
							},
						},
					},
				},
			},
			expect: `allOf:
  - $ref: "#/components/schemas/my-schema"
  - type: object
    required:
      - foo
    properties:
      "foo":
        type: object
        properties:
          "bar":
            type: boolean
`,
		},
		{
			ofs: &Ofs{
				OfType: AnyOf,
				Of: []OfSchema{
					OfRef("my-schema"),
					&Schema{
						RequiredProperties: []string{"foo"},
						Properties: Properties{
							{
								Name: "foo",
								Type: values.TypeObject,
								Properties: Properties{
									{
										Name: "bar",
										Type: values.TypeBoolean,
									},
								},
							},
						},
					},
				},
			},
			expect: `anyOf:
  - $ref: "#/components/schemas/my-schema"
  - type: object
    required:
      - foo
    properties:
      "foo":
        type: object
        properties:
          "bar":
            type: boolean
`,
		},
		{
			ofs: &Ofs{
				Of: []OfSchema{
					&Schema{
						Type: "string",
						Enum: []any{"a", "b"},
					},
				},
			},
			expect: `oneOf:
  - type: string
    enum:
      - a
      - b
`,
		},
		{
			ofs: &Ofs{
				OfType: OneOf,
				Of: []OfSchema{
					&Of{
						SchemaRef: "foo",
					},
				},
			},
			expect: `oneOf:
  - $ref: "#/components/schemas/foo"
`,
		},
		{
			ofs: &Ofs{
				OfType: OneOf,
				Of: []OfSchema{
					&Of{
						SchemaDef: &Schema{},
					},
				},
			},
			expect: `oneOf:
  - type: object
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			tc.ofs.writeYaml(w)
			data, err := w.Bytes()
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
		})
	}
}
