package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSecurityScheme_WriteYaml(t *testing.T) {
	testCases := []struct {
		security   SecurityScheme
		asSecurity bool
		expect     string
	}{
		{
			security: SecurityScheme{
				Name: "test",
			},
			expect: `test:
  type: http
`,
		},
		{
			security: SecurityScheme{
				Name: "test",
			},
			asSecurity: true,
			expect: `- test: []
`,
		},
		{
			security: SecurityScheme{
				Name:   "test",
				Type:   "oauth2",
				Scopes: []string{"read:foo"},
			},
			asSecurity: true,
			expect: `- test:
  - "read:foo"
`,
		},
		{
			security: SecurityScheme{
				Name:        "test",
				Description: "test desc",
				Type:        "apiKey",
				Scheme:      "basic",
				In:          "header",
				ParamName:   "X-API-KEY",
				Additional:  &testAdditional{},
				Extensions:  Extensions{"foo": "bar"},
				Comment:     "test comment",
			},
			expect: `test:
  #test comment
  description: "test desc"
  type: apiKey
  scheme: basic
  in: header
  name: X-API-KEY
  x-foo: bar
  foo: bar
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			tc.security.writeYaml(w, tc.asSecurity)
			data, err := w.Bytes()
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
		})
	}
}
