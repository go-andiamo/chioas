package refs

import (
	"github.com/go-andiamo/chioas/internal/tags"
	"strings"
)

const (
	ComponentsPrefix = "#/" + tags.Components + "/"
)

func Normalize(area string, ref string) string {
	prefix := ComponentsPrefix + area + "/"
	if tail, ok := strings.CutPrefix(ref, prefix); ok {
		// strip internal prefix and unescape JSON Pointer: ~1 => /, ~0 => ~
		return strings.ReplaceAll(strings.ReplaceAll(tail, "~1", "/"), "~0", "~")
	}
	return ref
}

// Canonical ref like "#/components/schemas/User"
func Canonical(area, name string) string {
	return ComponentsPrefix + area + "/" + name
}
