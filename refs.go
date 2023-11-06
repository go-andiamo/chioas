package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"golang.org/x/exp/slices"
	"strings"
)

const refComponentsPrefix = "#/" + tagNameComponents + "/"

func writeSchemaRef(ref string, isArray bool, w yaml.Writer) {
	if isArray {
		w.WriteTagValue(tagNameType, tagValueTypeArray).
			WriteTagStart(tagNameItems)
	}
	if strings.Contains(ref, "/") {
		w.WriteTagValue(tagNameRef, ref)
	} else {
		writeRef(tagNameSchemas, ref, w)
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
	return refComponentsPrefix + area + "/" + ref
}

func needsRefCheck(area, ref string) (string, string, bool) {
	if strings.Contains(ref, "/") {
		if strings.HasPrefix(ref, refComponentsPrefix) {
			if pts := strings.Split(ref[len(refComponentsPrefix):], "/"); len(pts) == 2 {
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
		w.WriteTagValue(tagNameRef, ref)
		return
	}
	if err := w.RefChecker(nil).RefCheck(area, ref); err != nil {
		w.SetError(err)
		return
	}
	w.WriteTagValue(tagNameRef, refComponentsPrefix+area+"/"+ref)
}

func writeItemRef(area, ref string, w yaml.Writer) {
	area, ref, chk := needsRefCheck(area, ref)
	if !chk {
		w.WriteItemValue(tagNameRef, ref)
		return
	}
	if err := w.RefChecker(nil).RefCheck(area, ref); err != nil {
		w.SetError(err)
		return
	}
	w.WriteItemValue(tagNameRef, refComponentsPrefix+area+"/"+ref)
}

// RefCheck implements yaml.RefChecker
// checks that refs specified exist in Definition.Components (if DocOptions.CheckRefs is set to true)
func (d *Definition) RefCheck(area, ref string) error {
	if d.Components == nil {
		return fmt.Errorf("$ref '%s%s/%s' invalid (definition has no components)", refComponentsPrefix, area, ref)
	}
	ok := false
	switch area {
	case tagNameSchemas:
		ok = slices.ContainsFunc(d.Components.Schemas, func(s Schema) bool {
			return s.Name == ref
		})
	case tagNameRequestBodies:
		if d.Components.Requests != nil {
			_, ok = d.Components.Requests[ref]
		}
	case tagNameResponses:
		if d.Components.Responses != nil {
			_, ok = d.Components.Responses[ref]
		}
	case tagNameParameters:
		if d.Components.Parameters != nil {
			_, ok = d.Components.Parameters[ref]
		}
	case tagNameExamples:
		ok = slices.ContainsFunc(d.Components.Examples, func(eg Example) bool {
			return eg.Name == ref
		})
	}
	if !ok {
		return fmt.Errorf("$ref '%s%s/%s' invalid", refComponentsPrefix, area, ref)
	}
	return nil
}
