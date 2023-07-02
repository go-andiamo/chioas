package chioas

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPathParams_WriteYaml(t *testing.T) {
	w := newYamlWriter(nil)
	pps := PathParams{
		"barId": {
			Description: "bar desc",
		},
		"fooId": {
			Description: "foo desc",
			Example:     "foo eg",
		},
	}
	path := "/foo/{fooId: [a-z]*}/bar/{barId}/{year}-{month}"
	pps.writeYaml(path, w)
	data, err := w.bytes()
	require.NoError(t, err)
	const expect = `parameters:
  - name: "fooId"
    description: "foo desc"
    in: "path"
    required: true
    example: "foo eg"
  - name: "barId"
    description: "bar desc"
    in: "path"
    required: true
  - name: "year"
    in: "path"
    required: true
  - name: "month"
    in: "path"
    required: true
`
	assert.Equal(t, expect, string(data))
}

func TestPathParams_WriteYaml_BadPath(t *testing.T) {
	w := newYamlWriter(nil)
	pps := PathParams{}
	path := "/foo/{fooId"
	pps.writeYaml(path, w)
	_, err := w.bytes()
	require.Error(t, err)
}

func TestPathParam_WriteYaml(t *testing.T) {
	w := newYamlWriter(nil)
	pp := PathParam{
		Description: "test desc",
		Example:     "fooey",
	}
	pp.writeYaml("foo", w)
	data, err := w.bytes()
	require.NoError(t, err)
	const expect = `- name: "foo"
  description: "test desc"
  in: "path"
  required: true
  example: "fooey"
`
	assert.Equal(t, expect, string(data))
}
