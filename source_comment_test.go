package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSourceComment(t *testing.T) {
	w := yaml.NewWriter(nil)
	pp := PathParam{
		Comment: SourceComment("foo", "bar"),
	}
	pp.writeYaml("foo", w)
	data, err := w.Bytes()
	assert.NoError(t, err)
	assert.Contains(t, string(data), "#source: source_comment_test.go:")
}
