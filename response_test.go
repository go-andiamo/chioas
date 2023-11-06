package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestResponses_WriteYaml(t *testing.T) {
	w := yaml.NewWriter(nil)
	rs := Responses{}
	rs.writeYaml(false, w)
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
			Comment:   "test comment",
		},
	}
	rs.writeYaml(false, w)
	data, err = w.Bytes()
	assert.NoError(t, err)
	const expect = `responses:
  200:
    #test comment
    description: OK
    content:
      "application/json":
        schema:
          $ref: "#/components/schemas/test_ok"
  201:
    description: Created
    content:
      "application/json":
        schema:
          $ref: "#/components/schemas/test_created"
`
	assert.Equal(t, expect, string(data))
}

func TestResponses_WriteYaml_Refd(t *testing.T) {
	w := yaml.NewWriter(nil)
	rs := Responses{}
	rs.writeYaml(false, w)
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
			Comment:   "won't see this",
		},
	}
	rs.writeYaml(false, w)
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

func TestResponse_WriteYaml(t *testing.T) {
	iptr := 16
	fptr := float32(16.16)
	testCases := []struct {
		response Response
		expect   string
	}{
		{
			response: Response{
				Description: "req desc",
				Additional:  &testAdditional{},
			},
			expect: `200:
  description: "req desc"
  content:
    "application/json":
      schema:
        type: object
  foo: bar
`,
		},
		{
			response: Response{
				Description: "req desc",
				NoContent:   true,
				Additional:  &testAdditional{},
			},
			expect: `200:
  description: "req desc"
  foo: bar
`,
		},
		{
			response: Response{
				Description: "req desc",
				ContentType: "text/csv",
				SchemaRef:   "req_ref",
				IsArray:     true,
			},
			expect: `200:
  description: "req desc"
  content:
    "text/csv":
      schema:
        type: array
        items:
          $ref: "#/components/schemas/req_ref"
`,
		},
		{
			response: Response{
				Description: "req desc",
				SchemaRef:   "/my/req_ref.yaml",
			},
			expect: `200:
  description: "req desc"
  content:
    "application/json":
      schema:
        $ref: "/my/req_ref.yaml"
`,
		},
		{
			response: Response{
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
			},
			expect: `200:
  description: "req desc"
  content:
    "application/json":
      schema:
        type: array
        items:
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
			response: Response{
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
			},
			expect: `200:
  description: "req desc"
  content:
    "application/json":
      schema:
        type: array
        items:
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
			response: Response{
				Description: "req desc",
				IsArray:     true,
				Schema:      myTestResponse{},
			},
			expect: `200:
  description: "req desc"
  content:
    "application/json":
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
			response: Response{
				Description: "req desc",
				IsArray:     true,
				Schema: &myTestResponse{
					Foo:   "foo eg",
					Bar:   true,
					Baz:   &iptr,
					Float: &fptr,
				},
			},
			expect: `200:
  description: "req desc"
  content:
    "application/json":
      schema:
        type: array
        items:
          type: object
          required:
            - bar
          properties:
            "foo":
              type: string
              example: "foo eg"
            "bar":
              type: boolean
              example: true
            "baz":
              type: integer
              example: 16
            "float":
              type: number
              example: 16.16
            "slice":
              type: array
              items:
                type: string
            "map":
              type: object
`,
		},
		{
			response: Response{
				Description: "req desc",
				IsArray:     true,
				Schema:      &mySchemaConverter{},
			},
			expect: `200:
  description: "req desc"
  content:
    "application/json":
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
			response: Response{
				Description: "req desc",
				IsArray:     true,
				Schema:      &mySchemaWriter{},
			},
			expect: `200:
  description: "req desc"
  content:
    "application/json":
      schema:
        type: array
        items:
          type: object
          properties:
            "foo":
              type: string
`,
		},
		{
			response: Response{
				Description: "req desc",
				IsArray:     true,
				Schema:      false,
			},
			expect: `200:
  description: "req desc"
  content:
    "application/json":
      schema:
        type: null
`,
		},
		{
			response: Response{
				Description: "req desc",
				Schema:      []myTestResponse{},
			},
			expect: `200:
  description: "req desc"
  content:
    "application/json":
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
			response: Response{
				Description: "req desc",
				Schema: []myTestResponse{
					{
						Foo:   "foo eg",
						Bar:   true,
						Baz:   &iptr,
						Float: &fptr,
					},
				},
			},
			expect: `200:
  description: "req desc"
  content:
    "application/json":
      schema:
        type: array
        items:
          type: object
          required:
            - bar
          properties:
            "foo":
              type: string
              example: "foo eg"
            "bar":
              type: boolean
              example: true
            "baz":
              type: integer
              example: 16
            "float":
              type: number
              example: 16.16
            "slice":
              type: array
              items:
                type: string
            "map":
              type: object
`,
		},
		{
			response: Response{
				SchemaRef: "foo",
				AlternativeContentTypes: ContentTypes{
					"application/xml": {
						SchemaRef: "foo",
					},
				},
			},
			expect: `200:
  description: OK
  content:
    "application/json":
      schema:
        $ref: "#/components/schemas/foo"
    "application/xml":
      schema:
        $ref: "#/components/schemas/foo"
`,
		},
		{
			response: Response{
				SchemaRef: "foo",
				Examples: Examples{
					{
						Name: "eg",
					},
				},
			},
			expect: `200:
  description: OK
  content:
    "application/json":
      schema:
        $ref: "#/components/schemas/foo"
      examples:
        eg:
          value: null
`,
		},
		{
			response: Response{
				SchemaRef: "foo",
				Examples: Examples{
					{
						Name: "eg",
					},
				},
				AlternativeContentTypes: ContentTypes{
					"application/xml": {
						Examples: Examples{
							{
								Name: "egXml",
							},
						},
					},
				},
			},
			expect: `200:
  description: OK
  content:
    "application/json":
      schema:
        $ref: "#/components/schemas/foo"
      examples:
        eg:
          value: null
    "application/xml":
      schema:
        type: object
      examples:
        egXml:
          value: null
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			tc.response.writeYaml(http.StatusOK, false, w)
			data, err := w.Bytes()
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
		})
	}
}
