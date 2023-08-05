package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestResponses_WriteYaml(t *testing.T) {
	w := yaml.NewWriter(nil)
	rs := Responses{}
	rs.writeYaml(w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, ``, string(data))

	w = yaml.NewWriter(nil)
	rs = Responses{
		http.StatusCreated: {
			SchemaRef: "test_created",
		},
		http.StatusOK: {
			SchemaRef: "test_ok",
		},
	}
	rs.writeYaml(w)
	data, err = w.Bytes()
	assert.NoError(t, err)
	const expect = `responses:
  200:
    description: "OK"
    content:
      application/json:
        schema:
          $ref: "#/components/schemas/test_ok"
  201:
    description: "Created"
    content:
      application/json:
        schema:
          $ref: "#/components/schemas/test_created"
`
	assert.Equal(t, expect, string(data))
}

func TestResponses_WriteYaml_Refd(t *testing.T) {
	w := yaml.NewWriter(nil)
	rs := Responses{}
	rs.writeYaml(w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, ``, string(data))

	w = yaml.NewWriter(nil)
	rs = Responses{
		http.StatusCreated: {
			Ref:       "foo",
			SchemaRef: "test_created",
		},
		http.StatusOK: {
			Ref:       "bar",
			SchemaRef: "test_ok",
		},
	}
	rs.writeYaml(w)
	data, err = w.Bytes()
	assert.NoError(t, err)
	const expect = `responses:
  200:
    $ref: "#/components/responses/bar"
  201:
    $ref: "#/components/responses/foo"
`
	assert.Equal(t, expect, string(data))
}

func TestResponse_WriteYaml_Basic(t *testing.T) {
	w := yaml.NewWriter(nil)
	r := Response{
		Description: "req desc",
		Additional:  &testAdditional{},
	}

	r.writeYaml(http.StatusOK, w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `200:
  description: "req desc"
  content:
    application/json:
      schema:
        type: "object"
  foo: "bar"
`
	assert.Equal(t, expect, string(data))
}

func TestResponse_WriteYaml_NoContent(t *testing.T) {
	w := yaml.NewWriter(nil)
	r := Response{
		Description: "req desc",
		NoContent:   true,
		Additional:  &testAdditional{},
	}

	r.writeYaml(http.StatusOK, w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `200:
  description: "req desc"
  foo: "bar"
`
	assert.Equal(t, expect, string(data))
}

func TestResponse_WriteYaml_TypeAndSchemaRefAndArray(t *testing.T) {
	w := yaml.NewWriter(nil)
	r := Response{
		Description: "req desc",
		ContentType: "text/csv",
		SchemaRef:   "req_ref",
		IsArray:     true,
	}

	r.writeYaml(http.StatusOK, w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `200:
  description: "req desc"
  content:
    text/csv:
      schema:
        type: "array"
        items:
          $ref: "#/components/schemas/req_ref"
`
	assert.Equal(t, expect, string(data))
}

func TestResponse_WriteYaml_SchemaRefPath(t *testing.T) {
	w := yaml.NewWriter(nil)
	r := Response{
		Description: "req desc",
		SchemaRef:   "/my/req_ref.yaml",
	}

	r.writeYaml(http.StatusOK, w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `200:
  description: "req desc"
  content:
    application/json:
      schema:
        $ref: "/my/req_ref.yaml"
`
	assert.Equal(t, expect, string(data))
}

func TestResponse_WriteYaml_Schema(t *testing.T) {
	w := yaml.NewWriter(nil)
	r := Response{
		Description: "req desc",
		IsArray:     true,
		Schema: Schema{
			Name:               "should not see this",
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
	}

	r.writeYaml(http.StatusOK, w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `200:
  description: "req desc"
  content:
    application/json:
      schema:
        type: "array"
        items:
          type: "object"
          required:
            - "foo"
          properties:
            "foo":
              type: "string"
            "bar":
              type: "string"
`
	assert.Equal(t, expect, string(data))
}

func TestResponse_WriteYaml_SchemaPtr(t *testing.T) {
	w := yaml.NewWriter(nil)
	r := Response{
		Description: "req desc",
		IsArray:     true,
		Schema: &Schema{
			Name:               "should not see this",
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
	}

	r.writeYaml(http.StatusOK, w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `200:
  description: "req desc"
  content:
    application/json:
      schema:
        type: "array"
        items:
          type: "object"
          required:
            - "foo"
          properties:
            "foo":
              type: "string"
            "bar":
              type: "string"
`
	assert.Equal(t, expect, string(data))
}

func TestResponse_WriteYaml_Struct(t *testing.T) {
	w := yaml.NewWriter(nil)
	r := Response{
		Description: "req desc",
		IsArray:     true,
		Schema:      myTestResponse{},
	}

	r.writeYaml(http.StatusOK, w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `200:
  description: "req desc"
  content:
    application/json:
      schema:
        type: "array"
        items:
          type: "object"
          required:
            - "bar"
          properties:
            "foo":
              type: "string"
            "bar":
              type: "boolean"
              example: false
            "baz":
              type: "integer"
            "float":
              type: "number"
            "slice":
              type: "array"
              items:
                type: "string"
            "map":
              type: "object"
`
	assert.Equal(t, expect, string(data))
}

func TestResponse_WriteYaml_StructPtr(t *testing.T) {
	w := yaml.NewWriter(nil)
	iptr := 16
	fptr := float32(16.16)
	r := Response{
		Description: "req desc",
		IsArray:     true,
		Schema: &myTestResponse{
			Foo:   "foo eg",
			Bar:   true,
			Baz:   &iptr,
			Float: &fptr,
		},
	}

	r.writeYaml(http.StatusOK, w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `200:
  description: "req desc"
  content:
    application/json:
      schema:
        type: "array"
        items:
          type: "object"
          required:
            - "bar"
          properties:
            "foo":
              type: "string"
              example: "foo eg"
            "bar":
              type: "boolean"
              example: true
            "baz":
              type: "integer"
              example: 16
            "float":
              type: "number"
              example: 16.160000
            "slice":
              type: "array"
              items:
                type: "string"
            "map":
              type: "object"
`
	assert.Equal(t, expect, string(data))
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

func TestResponse_WriteYaml_SchemaConverter(t *testing.T) {
	w := yaml.NewWriter(nil)
	r := Response{
		Description: "req desc",
		IsArray:     true,
		Schema:      &mySchemaConverter{},
	}

	r.writeYaml(http.StatusOK, w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `200:
  description: "req desc"
  content:
    application/json:
      schema:
        type: "array"
        items:
          type: "object"
          required:
            - "foo"
          properties:
            "foo":
              type: "string"
`
	assert.Equal(t, expect, string(data))
}

type mySchemaWriter struct {
}

func (m *mySchemaWriter) WriteSchema(w yaml.Writer) {
	w.WriteTagValue(tagNameType, "object").
		WriteTagStart(tagNameProperties).
		WriteTagStart(`"foo"`).
		WriteTagValue(tagNameType, "string").
		WriteTagEnd().WriteTagEnd()
}

func TestResponse_WriteYaml_SchemaWriter(t *testing.T) {
	w := yaml.NewWriter(nil)
	r := Response{
		Description: "req desc",
		IsArray:     true,
		Schema:      &mySchemaWriter{},
	}

	r.writeYaml(http.StatusOK, w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `200:
  description: "req desc"
  content:
    application/json:
      schema:
        type: "array"
        items:
          type: "object"
          properties:
            "foo":
              type: "string"
`
	assert.Equal(t, expect, string(data))
}

func TestResponse_WriteYaml_UnknownSchema(t *testing.T) {
	w := yaml.NewWriter(nil)
	r := Response{
		Description: "req desc",
		IsArray:     true,
		Schema:      false,
	}

	r.writeYaml(http.StatusOK, w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `200:
  description: "req desc"
  content:
    application/json:
      schema:
        type: "null"
`
	assert.Equal(t, expect, string(data))
}

func TestResponse_WriteYaml_StructArray(t *testing.T) {
	w := yaml.NewWriter(nil)
	r := Response{
		Description: "req desc",
		Schema:      []myTestResponse{},
	}

	r.writeYaml(http.StatusOK, w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `200:
  description: "req desc"
  content:
    application/json:
      schema:
        type: "array"
        items:
          type: "object"
          required:
            - "bar"
          properties:
            "foo":
              type: "string"
            "bar":
              type: "boolean"
            "baz":
              type: "integer"
            "float":
              type: "number"
            "slice":
              type: "array"
              items:
                type: "string"
            "map":
              type: "object"
`
	assert.Equal(t, expect, string(data))
}

func TestResponse_WriteYaml_StructArray_WithExampleElement(t *testing.T) {
	w := yaml.NewWriter(nil)
	iptr := 16
	fptr := float32(16.16)
	r := Response{
		Description: "req desc",
		Schema: []myTestResponse{
			{
				Foo:   "foo eg",
				Bar:   true,
				Baz:   &iptr,
				Float: &fptr,
			},
		},
	}

	r.writeYaml(http.StatusOK, w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `200:
  description: "req desc"
  content:
    application/json:
      schema:
        type: "array"
        items:
          type: "object"
          required:
            - "bar"
          properties:
            "foo":
              type: "string"
              example: "foo eg"
            "bar":
              type: "boolean"
              example: true
            "baz":
              type: "integer"
              example: 16
            "float":
              type: "number"
              example: 16.160000
            "slice":
              type: "array"
              items:
                type: "string"
            "map":
              type: "object"
`
	assert.Equal(t, expect, string(data))
}

type myTestResponse struct {
	Foo   string         `json:"foo,omitempty"`
	Bar   bool           `json:"bar"`
	Baz   *int           `json:"baz"`
	Float *float32       `json:"float"`
	Slice []string       `json:"slice,omitempty"`
	Map   map[string]any `json:"map,omitempty"`
}
