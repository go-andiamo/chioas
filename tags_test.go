package chioas

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTag_WriteYaml(t *testing.T) {
	w := newYamlWriter(nil)
	tag := Tag{
		Name:        "foo",
		Description: "test",
	}
	tag.writeYaml(w)
	data, err := w.bytes()
	require.NoError(t, err)
	const expect = `- name: "foo"
  description: "test"
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
