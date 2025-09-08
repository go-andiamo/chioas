package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/internal/refs"
	"github.com/go-andiamo/chioas/internal/tags"
	"strings"
)

type RefError struct {
	Msg      string
	Ref      string
	Path     string
	Method   string
	Item     any
	ItemName string
}

func (e *RefError) Error() string {
	return e.Msg
}

func (d *Definition) CheckRefs() (result []error) {
	for k, v := range d.Methods {
		result = append(result, v.checkRefs(root, k, d)...)
	}
	d.Paths.refsWalk("", func(path string, pathDef Path) {
		result = append(result, pathDef.checkRefs(path, d)...)
		for k, v := range pathDef.Methods {
			result = append(result, v.checkRefs(path, k, d)...)
		}
	})
	if d.Components != nil {
		result = append(result, d.Components.checkRefs(d)...)
	}
	return result
}

func (ps Paths) refsWalk(currPath string, fn func(path string, pathDef Path)) {
	for k, v := range ps {
		fn(currPath+k, v)
		v.Paths.refsWalk(currPath+k, fn)
	}
}

func (m Method) checkRefs(path string, method string, def *Definition) (result []error) {
	result = append(result, m.QueryParams.checkRefs(path, method, def)...)
	if m.Request != nil {
		result = append(result, m.Request.checkRefs(path, method, def)...)
	}
	for _, r := range m.Responses {
		result = append(result, r.checkRefs(path, method, def)...)
	}
	return result
}

func (c *Components) checkRefs(def *Definition) (result []error) {
	for i, s := range c.Schemas {
		seen := map[string]bool{
			refs.Canonical(tags.Schemas, s.Name): true,
		}
		result = append(result, s.checkRefs(fmt.Sprintf(refs.ComponentsPrefix+tags.Schemas+"[%d]", i), "", def, seen)...)
	}
	for name, r := range c.Requests {
		result = append(result, r.checkRefs(fmt.Sprintf(refs.ComponentsPrefix+tags.RequestBodies+"[%s]", name), "", def)...)
	}
	for name, r := range c.Responses {
		result = append(result, r.checkRefs(fmt.Sprintf(refs.ComponentsPrefix+tags.Responses+"[%s]", name), "", def)...)
	}
	for name, p := range c.Parameters {
		result = append(result, p.checkRefs(fmt.Sprintf(refs.ComponentsPrefix+tags.Parameters+"[%s]", name), name, def)...)
	}
	for i, eg := range c.Examples {
		result = append(result, eg.checkRefs(fmt.Sprintf(refs.ComponentsPrefix+tags.Examples+"[%d]", i), "", def)...)
	}
	return result
}

func (p CommonParameter) checkRefs(path string, name string, def *Definition) (result []error) {
	ref, area, ok, err := isInternalRef(p.SchemaRef, tags.Schemas)
	if ok {
		err = def.RefCheck(area, ref)
	}
	if err != nil {
		result = append(result, &RefError{
			Msg:      err.Error(),
			Ref:      p.SchemaRef,
			Path:     path,
			ItemName: name,
			Item:     p,
		})
	}
	if p.Schema != nil {
		result = append(result, checkHasSchemaRefWithSchema(p.SchemaRef, path, "", name, p)...)
		result = append(result, p.Schema.checkRefs(path, "", def, nil)...)
	}
	return result
}

func (p Path) checkRefs(path string, def *Definition) (result []error) {
	for name, pp := range p.PathParams {
		ref, area, ok, err := isInternalRef(pp.Ref, tags.Parameters)
		if ok {
			err = def.RefCheck(area, ref)
		}
		if err != nil {
			result = append(result, &RefError{
				Msg:      err.Error(),
				Ref:      pp.Ref,
				Path:     path,
				ItemName: name,
				Item:     pp,
			})
		}
		ref, area, ok, err = isInternalRef(pp.SchemaRef, tags.Schemas)
		if ok {
			err = def.RefCheck(area, ref)
		}
		if err != nil {
			result = append(result, &RefError{
				Msg:      err.Error(),
				Ref:      pp.SchemaRef,
				Path:     path,
				ItemName: name,
				Item:     pp,
			})
		}
		if pp.Schema != nil {
			result = append(result, checkHasSchemaRefWithSchema(pp.SchemaRef, path, "", name, pp)...)
			result = append(result, pp.Schema.checkRefs(path, "", def, nil)...)
		}
	}
	return result
}

func checkHasSchemaRefWithSchema(ref string, path string, method string, name string, item any) (result []error) {
	if ref != "" {
		result = []error{&RefError{
			Msg:      "has both schema and schemaRef",
			Path:     path,
			Method:   method,
			ItemName: name,
			Item:     item,
		}}
	}
	return result
}

func isInternalRef(r string, defArea string) (ref string, area string, ok bool, err error) {
	if r == "" {
		return "", "", false, nil
	}
	if strings.HasPrefix(r, refs.ComponentsPrefix) {
		if parts := strings.Split(r[len(refs.ComponentsPrefix):], "/"); len(parts) == 2 {
			if parts[0] != defArea {
				return "", "", false, fmt.Errorf("incorrect ref path %q - expected %q", r, refs.ComponentsPrefix+defArea+"/"+parts[1])
			}
			return parts[1], parts[0], true, nil
		} else {
			return "", "", false, fmt.Errorf("invalid ref path %q", r)
		}
	} else if a := strings.IndexRune(r, '/'); a == -1 {
		return r, defArea, true, nil
	}
	return "", "", false, nil
}

func (r *Request) checkRefs(path string, method string, def *Definition) (result []error) {
	ref, area, ok, err := isInternalRef(r.Ref, tags.RequestBodies)
	if ok {
		err = def.RefCheck(area, ref)
	}
	if err != nil {
		result = append(result, &RefError{
			Msg:    err.Error(),
			Ref:    r.Ref,
			Path:   path,
			Method: method,
			Item:   r,
		})
	}
	ref, area, ok, err = isInternalRef(r.SchemaRef, tags.Schemas)
	if ok {
		err = def.RefCheck(area, ref)
	}
	if err != nil {
		result = append(result, &RefError{
			Msg:    err.Error(),
			Ref:    r.SchemaRef,
			Path:   path,
			Method: method,
			Item:   r,
		})
	}
	if r.Schema != nil {
		result = append(result, checkHasSchemaRefWithSchema(r.SchemaRef, path, method, "", r)...)
		switch schema := r.Schema.(type) {
		case *Schema:
			result = append(result, schema.checkRefs(path, method, def, nil)...)
		case Schema:
			result = append(result, schema.checkRefs(path, method, def, nil)...)
		}
	}
	result = append(result, r.Examples.checkRefs(path, method, def)...)
	result = append(result, r.AlternativeContentTypes.checkRefs(path, method, def)...)
	return result
}

func (r Response) checkRefs(path string, method string, def *Definition) (result []error) {
	ref, area, ok, err := isInternalRef(r.Ref, tags.Responses)
	if ok {
		err = def.RefCheck(area, ref)
	}
	if err != nil {
		result = append(result, &RefError{
			Msg:    err.Error(),
			Ref:    r.Ref,
			Path:   path,
			Method: method,
			Item:   r,
		})
	}
	ref, area, ok, err = isInternalRef(r.SchemaRef, tags.Schemas)
	if ok {
		err = def.RefCheck(area, ref)
	}
	if err != nil {
		result = append(result, &RefError{
			Msg:    err.Error(),
			Ref:    r.SchemaRef,
			Path:   path,
			Method: method,
			Item:   r,
		})
	}
	if r.Schema != nil {
		result = append(result, checkHasSchemaRefWithSchema(r.SchemaRef, path, method, "", r)...)
		switch schema := r.Schema.(type) {
		case *Schema:
			result = append(result, schema.checkRefs(path, method, def, nil)...)
		case Schema:
			result = append(result, schema.checkRefs(path, method, def, nil)...)
		}
	}
	result = append(result, r.Examples.checkRefs(path, method, def)...)
	result = append(result, r.AlternativeContentTypes.checkRefs(path, method, def)...)
	return result
}

func (c ContentTypes) checkRefs(path string, method string, def *Definition) (result []error) {
	for _, ct := range c {
		ref, area, ok, err := isInternalRef(ct.SchemaRef, tags.Schemas)
		if ok {
			err = def.RefCheck(area, ref)
		}
		if err != nil {
			result = append(result, &RefError{
				Msg:    err.Error(),
				Ref:    ct.SchemaRef,
				Path:   path,
				Method: method,
				Item:   ct,
			})
		}
		if ct.Schema != nil {
			result = append(result, checkHasSchemaRefWithSchema(ct.SchemaRef, path, method, "", ct)...)
			switch schema := ct.Schema.(type) {
			case *Schema:
				result = append(result, schema.checkRefs(path, method, def, nil)...)
			case Schema:
				result = append(result, schema.checkRefs(path, method, def, nil)...)
			}
		}
		result = append(result, ct.Examples.checkRefs(path, method, def)...)
	}
	return result
}

func (egs Examples) checkRefs(path string, method string, def *Definition) (result []error) {
	for _, eg := range egs {
		result = append(result, eg.checkRefs(path, method, def)...)
	}
	return result
}

func (eg Example) checkRefs(path string, method string, def *Definition) (result []error) {
	ref, area, ok, err := isInternalRef(eg.ExampleRef, tags.Examples)
	if ok {
		err = def.RefCheck(area, ref)
	}
	if err != nil {
		result = append(result, &RefError{
			Msg:    err.Error(),
			Ref:    eg.ExampleRef,
			Path:   path,
			Method: method,
			Item:   eg,
		})
	}
	return result
}

func (qp QueryParams) checkRefs(path string, method string, def *Definition) (result []error) {
	for _, p := range qp {
		ref, area, ok, err := isInternalRef(p.Ref, tags.Parameters)
		if ok {
			err = def.RefCheck(area, ref)
		}
		if err != nil {
			result = append(result, &RefError{
				Msg:  err.Error(),
				Ref:  p.Ref,
				Path: path,
				Item: p,
			})
		}
		ref, area, ok, err = isInternalRef(p.SchemaRef, tags.Schemas)
		if ok {
			err = def.RefCheck(area, ref)
		}
		if err != nil {
			result = append(result, &RefError{
				Msg:  err.Error(),
				Ref:  p.SchemaRef,
				Path: path,
				Item: p,
			})
		}
		if p.Schema != nil {
			result = append(result, checkHasSchemaRefWithSchema(p.SchemaRef, path, method, "", p)...)
			result = append(result, p.Schema.checkRefs(path, method, def, nil)...)
		}
	}
	return result
}

func (s *Schema) checkRefs(path string, method string, def *Definition, seen map[string]bool) (result []error) {
	if seen == nil {
		seen = make(map[string]bool)
	}
	ref, area, ok, err := isInternalRef(s.SchemaRef, tags.Schemas)
	if ok {
		err = def.RefCheck(area, ref)
		cRef := refs.Canonical(area, ref)
		if seen[cRef] {
			result = append(result, &RefError{
				Msg:    "cyclic ref",
				Ref:    s.SchemaRef,
				Path:   path,
				Method: method,
				Item:   s,
			})
		} else {
			seen[cRef] = true
			defer delete(seen, cRef)
		}
	}
	if err != nil {
		result = append(result, &RefError{
			Msg:    err.Error(),
			Ref:    s.SchemaRef,
			Path:   path,
			Method: method,
			Item:   s,
		})
	}
	for _, p := range s.Properties {
		result = append(result, p.checkRefs(path, method, def, seen)...)
	}
	if s.Discriminator != nil {
		result = append(result, s.Discriminator.checkRefs(path, method, def)...)
	}
	if s.Ofs != nil {
		result = append(result, s.Ofs.checkRefs(path, method, def)...)
	}
	return result
}

func (p Property) checkRefs(path string, method string, def *Definition, seen map[string]bool) (result []error) {
	ref, area, ok, err := isInternalRef(p.SchemaRef, tags.Schemas)
	if ok {
		err = def.RefCheck(area, ref)
		cRef := refs.Canonical(area, ref)
		if seen[cRef] {
			result = append(result, &RefError{
				Msg:    "cyclic ref",
				Ref:    p.SchemaRef,
				Path:   path,
				Method: method,
				Item:   p,
			})
		} else {
			seen[cRef] = true
			defer delete(seen, cRef)
		}
	}
	if err != nil {
		result = append(result, &RefError{
			Msg:    err.Error(),
			Ref:    p.SchemaRef,
			Path:   path,
			Method: method,
			Item:   p,
		})
	}
	for _, ps := range p.Properties {
		result = append(result, ps.checkRefs(path, method, def, seen)...)
	}
	return result
}

func (d *Discriminator) checkRefs(path string, method string, def *Definition) (result []error) {
	for _, r := range d.Mapping {
		ref, area, ok, err := isInternalRef(r, tags.Schemas)
		if ok {
			err = def.RefCheck(area, ref)
		}
		if err != nil {
			result = append(result, &RefError{
				Msg:    err.Error(),
				Ref:    r,
				Path:   path,
				Method: method,
				Item:   d,
			})
		}
	}
	return result
}

func (ofs *Ofs) checkRefs(path string, method string, def *Definition) (result []error) {
	for _, of := range ofs.Of {
		if of.IsRef() {
			ref, area, ok, err := isInternalRef(of.Ref(), tags.Schemas)
			if ok {
				err = def.RefCheck(area, ref)
			}
			if err != nil {
				result = append(result, &RefError{
					Msg:    err.Error(),
					Ref:    of.Ref(),
					Path:   path,
					Method: method,
					Item:   of,
				})
			}
		} else if s := of.Schema(); s != nil {
			result = append(result, s.checkRefs(path, method, def, nil)...)
		}
	}
	return result
}
