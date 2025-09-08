package refs

import (
	"github.com/go-andiamo/chioas/internal/tags"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNormalize(t *testing.T) {
	t.Run("unprefixed", func(t *testing.T) {
		s := Normalize(tags.Schemas, "foo")
		require.Equal(t, "foo", s)
	})
	t.Run("prefixed", func(t *testing.T) {
		s := Normalize(tags.Schemas, ComponentsPrefix+tags.Schemas+"/foo")
		require.Equal(t, "foo", s)
	})
	t.Run("prefixed and escaped", func(t *testing.T) {
		s := Normalize(tags.Schemas, ComponentsPrefix+tags.Schemas+"/foo~1b")
		require.Equal(t, "foo/b", s)
		s = Normalize(tags.Schemas, ComponentsPrefix+tags.Schemas+"/foo~0b")
		require.Equal(t, "foo~b", s)
	})
}

func TestCanonical(t *testing.T) {
	s := Canonical(tags.Schemas, "foo")
	require.Equal(t, ComponentsPrefix+tags.Schemas+"/foo", s)
}
