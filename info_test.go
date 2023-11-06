package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInfo_WriteYaml(t *testing.T) {
	testCases := []struct {
		info   Info
		expect string
	}{
		{
			expect: `info:
  title: "API Documentation"
  version: "1.0.0"
`,
		},
		{
			info: Info{
				Title:       "foo",
				Description: "bar",
				Version:     "1.0.1",
			},
			expect: `info:
  title: foo
  description: bar
  version: "1.0.1"
`,
		},
		{
			info: Info{
				Contact: &Contact{},
			},
			expect: `info:
  title: "API Documentation"
  version: "1.0.0"
`,
		},
		{
			info: Info{
				Contact: &Contact{
					Name:  "me",
					Url:   "https://example.com",
					Email: "me@example.com",
				},
			},
			expect: `info:
  title: "API Documentation"
  version: "1.0.0"
  contact:
    name: me
    url: "https://example.com"
    email: "me@example.com"
`,
		},
		{
			info: Info{
				License: &License{},
			},
			expect: `info:
  title: "API Documentation"
  version: "1.0.0"
`,
		},
		{
			info: Info{
				License: &License{
					Name: "",
				},
			},
			expect: `info:
  title: "API Documentation"
  version: "1.0.0"
`,
		},
		{
			info: Info{
				License: &License{
					Name: "Apache 2.0",
					Url:  "https://example.com",
				},
			},
			expect: `info:
  title: "API Documentation"
  version: "1.0.0"
  license:
    name: "Apache 2.0"
    url: "https://example.com"
`,
		},
		{
			info: Info{
				ExternalDocs: &ExternalDocs{},
			},
			expect: `info:
  title: "API Documentation"
  version: "1.0.0"
`,
		},
		{
			info: Info{
				ExternalDocs: &ExternalDocs{
					Url:         "https://example.com",
					Description: "foo",
				},
			},
			expect: `info:
  title: "API Documentation"
  version: "1.0.0"
externalDocs:
  description: foo
  url: "https://example.com"
`,
		},
		{
			info: Info{
				Additional: &testAdditional{},
			},
			expect: `info:
  title: "API Documentation"
  version: "1.0.0"
  foo: bar
`,
		},
		{
			info: Info{
				Contact: &Contact{
					Name:    "me",
					Comment: "test comment",
				},
				License: &License{
					Name:    "test",
					Comment: "test comment",
				},
				Comment: "test comment",
				ExternalDocs: &ExternalDocs{
					Url:     "http://example.com",
					Comment: "test comment",
				},
			},
			expect: `info:
  #test comment
  title: "API Documentation"
  version: "1.0.0"
  contact:
    #test comment
    name: me
  license:
    #test comment
    name: test
externalDocs:
  #test comment
  url: "http://example.com"
`,
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			w := yaml.NewWriter(nil)
			tc.info.writeYaml(w)
			data, err := w.Bytes()
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, string(data))
		})
	}
}
