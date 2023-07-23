package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestPaths_WriteYaml(t *testing.T) {
	opts := &DocOptions{}
	paths := Paths{
		"/foo": {}, // won't be seen - because has no methods
		"/bar": {
			Tag: "tests",
			Methods: Methods{
				http.MethodGet: {
					Description: "this is a test",
				},
			},
		},
	}
	w := yaml.NewWriter(nil)
	paths.writeYaml(opts, false, "", w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expected = `"/bar":
  get:
    description: "this is a test"
    tags:
      - "tests"
    responses:
      200:
        description: "OK"
        content:
          application/json:
            schema:
              type: "object"
`
	assert.Equal(t, expected, string(data))
}

func TestPaths_WriteYaml_WithHidden(t *testing.T) {
	opts := &DocOptions{
		HideHeadMethods: true,
	}
	paths := Paths{
		"/foo": {}, // won't be seen - because has no methods
		"/bar": {
			Tag: "tests",
			Methods: Methods{
				http.MethodGet: {
					Description: "this is a test",
				},
			},
		},
		"/baz": {
			Methods: Methods{
				http.MethodGet:  {HideDocs: true},
				http.MethodHead: {},
			},
		},
		"/buzz": {
			Paths: Paths{
				"/foo": {}, // no methods
				"/bar": { // no visible methods
					Methods: Methods{
						http.MethodHead: {},
					},
				},
				"/baz": { // no visible methods
					Methods: Methods{
						http.MethodGet: {HideDocs: true},
					},
				},
				"/buzz": {
					Methods: Methods{
						http.MethodHead: {},
					},
					Paths: Paths{
						"/buzz": {
							HideDocs: true,
							Methods: Methods{
								http.MethodGet: {},
							},
						},
						"/foo": {
							Methods: Methods{
								http.MethodGet:  {},
								http.MethodHead: {},
							},
						},
					},
				},
			},
		},
	}
	w := yaml.NewWriter(nil)
	paths.writeYaml(opts, false, "", w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expected = `"/bar":
  get:
    description: "this is a test"
    tags:
      - "tests"
    responses:
      200:
        description: "OK"
        content:
          application/json:
            schema:
              type: "object"
"/buzz/buzz/foo":
  get:
    responses:
      200:
        description: "OK"
        content:
          application/json:
            schema:
              type: "object"
`
	assert.Equal(t, expected, string(data))
}

func TestFlatPath_writeYaml_WithBadUrl(t *testing.T) {
	w := yaml.NewWriter(nil)
	p := flatPath{
		path: "badpath/{unclosed",
		def: Path{
			Methods: Methods{
				http.MethodGet: {},
			},
		},
	}
	p.writeYaml(nil, false, "", w)
	err := w.Errored()
	assert.Error(t, err)
}
