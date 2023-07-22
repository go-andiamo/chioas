package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestPaths_WriteYaml(t *testing.T) {
	paths := Paths{
		"/foo": {},
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
	paths.writeYaml(nil, "", w)
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
	p.writeYaml(nil, "", w)
	err := w.Errored()
	assert.Error(t, err)
}
