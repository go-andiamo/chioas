package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/internal/tags"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteContentType(t *testing.T) {
	testCases := []struct {
		contentType string
		ct          ContentType
		expect      string
	}{
		{
			expect: `"application/json":
  schema:
    type: object
`,
		},
		{
			ct: ContentType{
				IsArray: true,
			},
			expect: `"application/json":
  schema:
    type: object
`,
		},
		{
			ct: ContentType{
				SchemaRef: "foo",
			},
			expect: `"application/json":
  schema:
    $ref: "#/components/schemas/foo"
`,
		},
		{
			ct: ContentType{
				SchemaRef: "foo",
				IsArray:   true,
			},
			expect: `"application/json":
  schema:
    type: array
    items:
      $ref: "#/components/schemas/foo"
`,
		},
		{
			ct: ContentType{
				Schema: &mySchemaConverter{},
			},
			expect: `"application/json":
  schema:
    type: object
    required:
      - foo
    properties:
      "foo":
        type: string
`,
		},
		{
			ct: ContentType{
				Schema:  &mySchemaConverter{},
				IsArray: true,
			},
			expect: `"application/json":
  schema:
    type: array
    items:
      type: object
      required:
        - foo
      properties:
        "foo":
          type: string
`,
		},
		{
			ct: ContentType{
				Schema: &mySchemaWriter{},
			},
			expect: `"application/json":
  schema:
    type: object
    properties:
      "foo":
        type: string
`,
		},
		{
			ct: ContentType{
				Schema: myTestResponse{},
			},
			expect: `"application/json":
  schema:
    type: object
    required:
      - bar
    properties:
      "foo":
        type: string
      "bar":
        type: boolean
        example: false
      "baz":
        type: integer
      "float":
        type: number
      "slice":
        type: array
        items:
          type: string
      "map":
        type: object
`,
		},
		{
			ct: ContentType{
				Schema: &myTestResponse{},
			},
			expect: `"application/json":
  schema:
    type: object
    required:
      - bar
    properties:
      "foo":
        type: string
      "bar":
        type: boolean
        example: false
      "baz":
        type: integer
      "float":
        type: number
      "slice":
        type: array
        items:
          type: string
      "map":
        type: object
`,
		},
		{
			ct: ContentType{
				Schema: []myTestResponse{},
			},
			expect: `"application/json":
  schema:
    type: array
    items:
      type: object
      required:
        - bar
      properties:
        "foo":
          type: string
        "bar":
          type: boolean
        "baz":
          type: integer
        "float":
          type: number
        "slice":
          type: array
          items:
            type: string
        "map":
          type: object
`,
		},
		{
			ct: ContentType{
				SchemaRef: "foo",
				Examples: Examples{
					{
						Name: "eg",
					},
				},
			},
			expect: `"application/json":
  schema:
    $ref: "#/components/schemas/foo"
  examples:
    eg:
      value: null
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			writeContentType(tc.contentType, tc.ct, w)
			data, err := w.Bytes()
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
		})
	}
}

func TestContentType_Alternatives_Panics(t *testing.T) {
	ct := ContentType{}
	assert.Panics(t, func() {
		_ = ct.alternatives()
	})
}

type mySchemaConverter struct {
}

func (m *mySchemaConverter) ToSchema() *Schema {
	return &Schema{
		Name:               "test",
		RequiredProperties: []string{"foo"},
		Properties: []Property{
			{
				Name: "foo",
				Type: "string",
			},
		},
	}
}

type mySchemaWriter struct {
}

func (m *mySchemaWriter) WriteSchema(w yaml.Writer) {
	w.WriteTagValue(tags.Type, "object").
		WriteTagStart(tags.Properties).
		WriteTagStart(`"foo"`).
		WriteTagValue(tags.Type, "string").
		WriteTagEnd().WriteTagEnd()
}

type myTestResponse struct {
	Foo   string         `json:"foo,omitempty"`
	Bar   bool           `json:"bar"`
	Baz   *int           `json:"baz"`
	Float *float32       `json:"float"`
	Slice []string       `json:"slice,omitempty"`
	Map   map[string]any `json:"map,omitempty"`
}
