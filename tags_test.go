package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTag_WriteYaml(t *testing.T) {
	w := yaml.NewWriter(nil)
	tag := Tag{
		Name:        "foo",
		Description: "test",
		Additional:  &testAdditional{},
		Extensions:  Extensions{"foo": "bar"},
		Comment:     "test comment",
	}
	tag.writeYaml(w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `- name: foo
  #test comment
  description: test
  x-foo: bar
  foo: bar
`
	assert.Equal(t, expect, string(data))
}

func TestTag_WriteYaml_WithExternalDocs(t *testing.T) {
	w := yaml.NewWriter(nil)
	tag := Tag{
		Name:        "foo",
		Description: "test",
		ExternalDocs: &ExternalDocs{
			Url:         "https://example.com",
			Description: "ext desc",
		},
	}
	tag.writeYaml(w)
	data, err := w.Bytes()
	require.NoError(t, err)
	const expect = `- name: foo
  description: test
  externalDocs:
    description: "ext desc"
    url: "https://example.com"
`
	assert.Equal(t, expect, string(data))
}

func TestDefaultTag(t *testing.T) {
	tag := defaultTag("", "foo")
	assert.Equal(t, "foo", tag)
	tag = defaultTag("foo", "bar")
	assert.Equal(t, "bar", tag)
	tag = defaultTag("foo", "")
	assert.Equal(t, "foo", tag)
}
