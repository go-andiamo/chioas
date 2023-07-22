package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServers_WriteYaml(t *testing.T) {
	w := yaml.NewWriter(nil)
	s := Servers{}
	s.writeYaml(w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, ``, string(data))

	w = yaml.NewWriter(nil)
	s = Servers{
		"api/v1": Server{},
	}
	s.writeYaml(w)
	data, err = w.Bytes()
	assert.NoError(t, err)
	const expect = `servers:
  - url: "api/v1"
`
	assert.Equal(t, expect, string(data))
}

func TestServer_WriteYaml(t *testing.T) {
	w := yaml.NewWriter(nil)
	s := Server{}
	s.writeYaml("api/v1", w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, `- url: "api/v1"
`, string(data))

	w = yaml.NewWriter(nil)
	s = Server{
		Description: "foo",
		Additional:  &testAdditional{},
	}
	s.writeYaml("api/v1", w)
	data, err = w.Bytes()
	assert.NoError(t, err)
	assert.Equal(t, `- url: "api/v1"
  description: "foo"
  foo: "bar"
`, string(data))
}
