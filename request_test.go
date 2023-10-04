package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRequest_WriteYaml(t *testing.T) {
	testCases := []struct {
		request *Request
		expect  string
	}{
		{
			request: &Request{
				Description: "desc",
				Required:    true,
				SchemaRef:   "foo",
				Additional:  &testAdditional{},
				Comment:     "test comment",
			},
			expect: `requestBody:
  #test comment
  description: "desc"
  required: true
  content:
    application/json:
      schema:
        $ref: "#/components/schemas/foo"
  foo: "bar"
`,
		},
		{
			request: &Request{
				Ref:     "foo",
				Comment: "won't see this",
			},
			expect: `requestBody:
  $ref: "#/components/requestBodies/foo"
`,
		},
		{
			request: &Request{
				Description: "desc",
				Required:    true,
				SchemaRef:   "foo",
				IsArray:     true,
			},
			expect: `requestBody:
  description: "desc"
  required: true
  content:
    application/json:
      schema:
        type: "array"
        items:
          $ref: "#/components/schemas/foo"
`,
		},
		{
			request: &Request{
				SchemaRef: "foo",
				AlternativeContentTypes: ContentTypes{
					"application/xml": {
						SchemaRef: "foo",
					},
				},
			},
			expect: `requestBody:
  required: false
  content:
    application/json:
      schema:
        $ref: "#/components/schemas/foo"
    application/xml:
      schema:
        $ref: "#/components/schemas/foo"
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			tc.request.writeYaml(w)
			data, err := w.Bytes()
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
		})
	}
}
