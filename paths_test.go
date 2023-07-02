package chioas

import (
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
	w := newYamlWriter(nil)
	paths.writeYaml("", w)
	data, err := w.bytes()
	require.NoError(t, err)
	const expected = `"/bar":
  get:
    description: "this is a test"
    tags:
      - "tests"
`
	assert.Equal(t, expected, string(data))
}
