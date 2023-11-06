package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDiscriminator_WriteYaml(t *testing.T) {
	d := &Discriminator{}
	w := yaml.NewWriter(nil)
	d.writeYaml(w)
	_, err := w.Bytes()
	assert.Error(t, err)

	d = &Discriminator{
		PropertyName: "foo",
		Additional:   &testAdditional{},
		Extensions:   map[string]any{"foo": "bar"},
		Comment:      "this is a comment",
	}
	w = yaml.NewWriter(nil)
	d.writeYaml(w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `discriminator:
  #this is a comment
  propertyName: foo
  x-foo: bar
  foo: bar
`
	assert.Equal(t, expect, string(data))

	d = &Discriminator{
		PropertyName: "foo",
		Mapping: map[string]string{
			"a": "my-schema1",
			"b": "my-schema2",
		},
	}
	w = yaml.NewWriter(nil)
	d.writeYaml(w)
	data, err = w.Bytes()
	assert.NoError(t, err)
	assert.Contains(t, string(data), `    a: "#/components/schemas/my-schema1"`)
	assert.Contains(t, string(data), `    b: "#/components/schemas/my-schema2"`)
}
