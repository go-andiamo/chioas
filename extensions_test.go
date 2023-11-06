package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExtensions_WriteYaml(t *testing.T) {
	e := Extensions{
		"foo":   "bar",
		"x-bar": 1,
	}
	w := yaml.NewWriter(nil)
	writeExtensions(e, w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	assert.Contains(t, string(data), "x-foo: bar\n")
	assert.Contains(t, string(data), "x-bar: 1\n")

	e = Extensions{}
	w = yaml.NewWriter(nil)
	writeExtensions(e, w)
	data, err = w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, ``, string(data))

	e = nil
	w = yaml.NewWriter(nil)
	writeExtensions(e, w)
	data, err = w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, ``, string(data))
}
