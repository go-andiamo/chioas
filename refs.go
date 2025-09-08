package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/internal/refs"
	"github.com/go-andiamo/chioas/internal/tags"
	"github.com/go-andiamo/chioas/internal/values"
	"github.com/go-andiamo/chioas/yaml"
	"golang.org/x/exp/slices"
	"strings"
)

func writeSchemaRef(ref string, isArray bool, w yaml.Writer) {
	if isArray {
		w.WriteTagValue(tags.Type, values.TypeArray).
			WriteTagStart(tags.Items)
	}
	if strings.Contains(ref, "/") {
		w.WriteTagValue(tags.Ref, ref)
	} else {
		writeRef(tags.Schemas, ref, w)
	}
	if isArray {
		w.WriteTagEnd()
	}
}

func refCheck(area, ref string, w yaml.Writer) string {
	area, ref, chk := needsRefCheck(area, ref)
	if !chk {
		return ref
	}
	if err := w.RefChecker(nil).RefCheck(area, ref); err != nil {
		w.SetError(err)
	}
	return refs.ComponentsPrefix + area + "/" + ref
}

func needsRefCheck(area, ref string) (string, string, bool) {
	if strings.Contains(ref, "/") {
		if strings.HasPrefix(ref, refs.ComponentsPrefix) {
			if pts := strings.Split(ref[len(refs.ComponentsPrefix):], "/"); len(pts) == 2 {
				return pts[0], pts[1], true
			}
		}
		return area, ref, false
	}
	return area, ref, true
}

func writeRef(area, ref string, w yaml.Writer) {
	area, ref, chk := needsRefCheck(area, ref)
	if !chk {
		w.WriteTagValue(tags.Ref, ref)
		return
	}
	if err := w.RefChecker(nil).RefCheck(area, ref); err != nil {
		w.SetError(err)
		return
	}
	w.WriteTagValue(tags.Ref, refs.ComponentsPrefix+area+"/"+ref)
}

func writeItemRef(area, ref string, w yaml.Writer) {
	area, ref, chk := needsRefCheck(area, ref)
	if !chk {
		w.WriteItemValue(tags.Ref, ref)
		return
	}
	if err := w.RefChecker(nil).RefCheck(area, ref); err != nil {
		w.SetError(err)
		return
	}
	w.WriteItemValue(tags.Ref, refs.ComponentsPrefix+area+"/"+ref)
}

// RefCheck implements yaml.RefChecker
// checks that refs specified exist in Definition.Components (if DocOptions.CheckRefs is set to true)
func (d *Definition) RefCheck(area, ref string) error {
	if d.Components == nil {
		return fmt.Errorf("$ref '%s%s/%s' invalid (definition has no components)", refs.ComponentsPrefix, area, ref)
	}
	ok := false
	switch area {
	case tags.Schemas:
		ok = slices.ContainsFunc(d.Components.Schemas, func(s Schema) bool {
			return s.Name == ref
		})
	case tags.RequestBodies:
		if d.Components.Requests != nil {
			_, ok = d.Components.Requests[ref]
		}
	case tags.Responses:
		if d.Components.Responses != nil {
			_, ok = d.Components.Responses[ref]
		}
	case tags.Parameters:
		if d.Components.Parameters != nil {
			_, ok = d.Components.Parameters[ref]
		}
	case tags.Examples:
		ok = slices.ContainsFunc(d.Components.Examples, func(eg Example) bool {
			return eg.Name == ref
		})
	}
	if !ok {
		return fmt.Errorf("$ref '%s%s/%s' invalid", refs.ComponentsPrefix, area, ref)
	}
	return nil
}
