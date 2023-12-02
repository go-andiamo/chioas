package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteAdditional(t *testing.T) {
	ap := AdditionalOasProperties{"foo": "bar"}
	w := yaml.NewWriter(nil)
	writeAdditional(ap, nil, w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `foo: bar
`
	assert.Equal(t, expect, string(data))
}

func TestAdditionalOasProperties(t *testing.T) {
	ap := AdditionalOasProperties{"foo": "bar"}
	w := yaml.NewWriter(nil)
	ap.Write(nil, w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `foo: bar
`
	assert.Equal(t, expect, string(data))
}
