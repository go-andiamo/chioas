package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPathParam_WriteYaml(t *testing.T) {
	w := yaml.NewWriter(nil)
	pp := PathParam{
		Description: "test desc",
		Example:     "fooey",
		Additional:  &testAdditional{},
	}
	pp.writeYaml("foo", w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `- name: "foo"
  description: "test desc"
  in: "path"
  required: true
  example: "fooey"
  foo: "bar"
`
	assert.Equal(t, expect, string(data))
}
