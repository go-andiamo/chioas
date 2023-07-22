package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComponents_WriteYaml(t *testing.T) {
	c := &Components{
		Schemas: Schemas{
			{
				Name: "test",
			},
		},
		Additional: &testAdditional{},
	}
	w := yaml.NewWriter(nil)
	c.writeYaml(w)

	data, err := w.Bytes()
	assert.NoError(t, err)
	const expect = `components:
  schemas:
    "test":
      type: "object"
  foo: "bar"
`
	assert.Equal(t, expect, string(data))
}
