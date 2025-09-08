package chioas

import (
	"encoding/json"
	"fmt"
	"github.com/go-andiamo/chioas/internal/refs"
	"github.com/go-andiamo/chioas/internal/tags"
	"github.com/go-andiamo/chioas/internal/values"
	"github.com/go-andiamo/splitter"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

const (
	unMsgMustBeObject   = `property %q must be an object`
	unMsgMustBeArray    = `property %q must be an array`
	unMsgMustBeString   = `property %q must be a string`
	unMsgInvalidElement = `property %q contains invalid element`
	unMsgInvalidValue   = `property %q contains invalid value`
)

func (d *Definition) UnmarshalJSON(data []byte) error {
	m := make(map[string]any)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	return d.unmarshalObj(m)
}

func (d *Definition) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]any)
	if err := value.Decode(&m); err != nil {
		return err
	}
	return d.unmarshalObj(m)
}

func (d *Definition) unmarshalObj(m map[string]any) (err error) {
	d.Extensions = extensionsFrom(m)
	if v, err := objFromProperty[Info](m, tags.Info); err != nil {
		return err
	} else if v != nil {
		d.Info = *v
	}
	if v, err := objFromProperty[ExternalDocs](m, tags.ExternalDocs); err != nil {
		return err
	} else if v != nil {
		d.Info.ExternalDocs = v
	}
	if d.Tags, err = sliceFromProperty[Tag](m, tags.Tags); err == nil {
		if d.Servers, err = serversFrom(m); err == nil {
			if d.Security, err = securityFrom(m); err == nil {
				if d.Components, err = componentsFrom(m); err == nil {
					err = d.unmarshalPaths(m)
				}
			}
		}
	}
	if err == nil {
		err = d.correctPathParams()
	}
	return err
}

func (d *Definition) correctPathParams() (err error) {
	// look at all paths - and extract any method query params that are 'In' path and move them up to the path
	err = d.WalkPaths(func(path string, pathDef *Path) (cont bool, err error) {
		pathDef.PathParams = pathDef.extractPathParams(d)
		return true, nil
	})
	return err
}

func (p *Path) extractPathParams(def *Definition) (result PathParams) {
	keys := maps.Keys(p.Methods)
	for _, k := range keys {
		m := p.Methods[k]
		method := &m
		if pps, removed := method.extractPathParams(def); removed {
			p.Methods[k] = *method
			if result == nil {
				result = pps
			} else {
				for pk, pp := range pps {
					result[pk] = pp
				}
			}
		}
	}
	return result
}

func (m *Method) extractPathParams(def *Definition) (PathParams, bool) {
	if len(m.QueryParams) > 0 {
		result := make(PathParams, len(m.QueryParams))
		removed := false
		rl := 0
		for i := 0; i < len(m.QueryParams); i++ {
			if name, pp, ok := isPathParam(m.QueryParams[i], def); ok {
				removed = true
				result[name] = *pp
			} else {
				m.QueryParams[rl] = m.QueryParams[i]
				rl++
			}
		}
		if removed {
			m.QueryParams = m.QueryParams[:rl]
			return result, true
		}
	}
	return nil, false
}

func isPathParam(qp QueryParam, def *Definition) (string, *PathParam, bool) {
	if qp.Ref != "" {
		if def.Components != nil {
			if cp, ok := def.Components.Parameters[refs.Normalize(tags.Parameters, qp.Ref)]; ok && cp.In == values.Path {
				return cp.Name, &PathParam{
					Ref: qp.Ref,
				}, true
			}
		}
	} else if qp.In == values.Path {
		name := qp.Name
		return name, &PathParam{
			Description: qp.Description,
			Example:     qp.Example,
			Extensions:  qp.Extensions,
			Additional:  qp.Additional,
			Comment:     qp.Comment,
			Schema:      qp.Schema,
			SchemaRef:   qp.SchemaRef,
			Ref:         qp.Ref,
		}, true
	}
	return "", nil, false
}

func (d *Definition) unmarshalPaths(m map[string]any) (err error) {
	if v, ok := m[tags.Paths]; ok {
		if paths, ok := v.(map[string]any); ok {
			holders := make([]*pathHolder, 0, len(paths))
			var rootPath *pathHolder
			for path, v := range paths {
				if vm, ok := v.(map[string]any); ok {
					if path == root {
						rootPath = &pathHolder{
							origPath: path,
							path:     path,
							obj:      vm,
						}
					} else {
						holders = append(holders, &pathHolder{
							origPath: path,
							path:     path,
							obj:      vm,
							subs:     make(map[string]*pathHolder, 0),
						})
					}
				} else {
					err = fmt.Errorf(unMsgMustBeObject, path)
					break
				}
			}
			if err == nil {
				if rootPath != nil {
					if d.Methods, err = rootPath.getMethods(); err != nil {
						return err
					}
				}
				var treeHolders map[string]*pathHolder
				d.Paths = make(Paths, len(treeHolders))
				if treeHolders, err = unflattenHolders(holders); err == nil {
					for _, holder := range treeHolders {
						if d.Paths[holder.path], err = holder.getPath(); err != nil {
							return err
						}
					}
				}
			}
		} else {
			err = fmt.Errorf(unMsgMustBeObject, tags.Paths)
		}
	}
	return err
}

var pathSplitter = splitter.MustCreateSplitter('/', splitter.CurlyBrackets).AddDefaultOptions(splitter.IgnoreEmptyFirst, splitter.IgnoreEmptyLast)

func unflattenHolders(holders []*pathHolder) (result map[string]*pathHolder, err error) {
	// sort by paths so that descendants appear after parent...
	slices.SortFunc(holders, func(a, b *pathHolder) int {
		return strings.Compare(a.origPath, b.origPath)
	})
	result = make(map[string]*pathHolder, len(holders))
	add := func(h *pathHolder) error {
		if parts, err := pathSplitter.Split(h.path); err == nil {
			if len(parts) == 1 {
				h.path = "/" + parts[0]
				result[h.path] = h
			} else {
				parent := result["/"+parts[0]]
				if parent == nil {
					parent = &pathHolder{
						path: "/" + parts[0],
						subs: make(map[string]*pathHolder),
					}
					result["/"+parts[0]] = parent
				}
				for _, part := range parts[1 : len(parts)-1] {
					if newParent, ok := parent.subs["/"+part]; ok {
						parent = newParent
					} else {
						newParent = &pathHolder{
							path: "/" + part,
							subs: make(map[string]*pathHolder),
						}
						parent.subs["/"+part] = newParent
						parent = newParent
					}
				}
				parent.subs["/"+parts[len(parts)-1]] = h
			}
			return nil
		} else {
			return fmt.Errorf(`failed to split path %q: %v`, h.path, err)
		}
	}
	for _, holder := range holders {
		if err = add(holder); err != nil {
			break
		}
	}
	return result, err
}

type pathHolder struct {
	origPath string
	path     string
	obj      map[string]any
	subs     map[string]*pathHolder
	methods  Methods
}

func (ph *pathHolder) getPath() (result Path, err error) {
	result.Extensions = extensionsFrom(ph.obj)
	if result.Methods, err = ph.getMethods(); err == nil {
		if len(ph.subs) > 0 {
			result.Paths = make(Paths, len(ph.subs))
			for path, holder := range ph.subs {
				if result.Paths[path], err = holder.getPath(); err != nil {
					break
				}
			}
		}
	}
	return result, err
}

// UnmarshalMethods defines which methods are unmarshalled from OAS
var UnmarshalMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodPatch:   true,
	http.MethodDelete:  true,
	http.MethodOptions: true,
	http.MethodConnect: true,
	http.MethodTrace:   true,
}

func (ph *pathHolder) getMethods() (Methods, error) {
	if ph.methods == nil && ph.obj != nil {
		// build methods...
		ph.methods = make(Methods, len(ph.obj))
		for k, v := range ph.obj {
			method := strings.ToUpper(k)
			if UnmarshalMethods[method] {
				if vm, ok := v.(map[string]any); ok {
					if m, err := fromObj[Method](vm); err == nil {
						ph.methods[method] = *m
					} else {
						return nil, err
					}
				} else {
					return nil, fmt.Errorf(`method %q must be an object (on path %q)`, method, ph.origPath)
				}
			}
		}
	}
	return ph.methods, nil
}

func (m *Method) unmarshalObj(o map[string]any) (err error) {
	m.Extensions = extensionsFrom(o)
	if m.Description, err = stringFromProperty(o, tags.Description); err == nil {
		if m.Summary, err = stringFromProperty(o, tags.Summary); err == nil {
			if m.OperationId, err = stringFromProperty(o, tags.OperationId); err == nil {
				if m.Deprecated, err = booleanFromProperty(o, tags.Deprecated); err == nil {
					var tgs []string
					if tgs, err = stringsSliceFromProperty(o, tags.Tags); err == nil {
						if len(tgs) > 0 {
							m.Tag = tgs[0]
						}
						if err = m.unmarshalSecurity(o); err == nil {
							if m.Request, err = objFromProperty[Request](o, tags.RequestBody); err == nil {
								if m.QueryParams, err = sliceFromProperty[QueryParam](o, tags.Parameters); err == nil {
									var responses []Response
									var codes []string
									if responses, codes, err = namedSliceFromProperty[Response](o, tags.Responses); err == nil && responses != nil {
										m.Responses = make(Responses, len(responses))
										for i, response := range responses {
											var code int
											if code, err = strconv.Atoi(codes[i]); err != nil {
												err = fmt.Errorf(`invalid response code %q`, codes[i])
												break
											}
											m.Responses[code] = response
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return err
}

func (m *Method) unmarshalSecurity(o map[string]any) (err error) {
	var secs []methodSecurity
	if secs, err = sliceFromProperty[methodSecurity](o, tags.Security); err == nil {
		if len(secs) == 1 && !secs[0].set {
			m.OptionalSecurity = true
		} else {
			for _, sec := range secs {
				if sec.set {
					m.Security = append(m.Security, SecurityScheme{Name: sec.name})
				}
			}
		}
	}
	return err
}

type methodSecurity struct {
	set  bool
	name string
}

func (ms *methodSecurity) unmarshalObj(m map[string]any) error {
	if len(m) != 0 {
		ms.set = true
		for k := range m {
			ms.name = k
			break
		}
	}
	return nil
}

type unmarshaler interface {
	unmarshalObj(m map[string]any) error
}

func objFromProperty[T any](m map[string]any, name string) (*T, error) {
	if v, ok := m[name]; ok {
		if vm, ok := v.(map[string]any); ok {
			return fromObj[T](vm)
		} else {
			return nil, fmt.Errorf(unMsgMustBeObject, name)
		}
	}
	return nil, nil
}

func fromObj[T any](m map[string]any) (*T, error) {
	t := new(T)
	if um, ok := isUnmarshaler(t); ok {
		err := um.unmarshalObj(m)
		return t, err
	} else {
		return nil, fmt.Errorf(`type '%T' does not implement unmarshaler.unmarshalObj()`, t)
	}
}

func isUnmarshaler(t any) (unmarshaler, bool) {
	u, ok := t.(unmarshaler)
	return u, ok
}

func sliceFromProperty[T any](m map[string]any, name string) ([]T, error) {
	if v, ok := m[name]; ok {
		if vs, ok := v.([]any); ok {
			result := make([]T, 0, len(vs))
			for _, v := range vs {
				if vm, ok := v.(map[string]any); ok {
					if item, err := fromObj[T](vm); err == nil {
						result = append(result, *item)
					} else {
						return nil, err
					}
				} else {
					return nil, fmt.Errorf(unMsgInvalidElement, name)
				}
			}
			return result, nil
		} else {
			return nil, fmt.Errorf(unMsgMustBeArray, name)
		}
	}
	return nil, nil
}

func namedSliceFromProperty[T any](m map[string]any, name string) ([]T, []string, error) {
	if v, ok := m[name]; ok {
		if vs, ok := v.(map[string]any); ok {
			result := make([]T, 0, len(vs))
			names := make([]string, 0, len(vs))
			for k, v := range vs {
				if vm, ok := v.(map[string]any); ok {
					if item, err := fromObj[T](vm); err == nil {
						result = append(result, *item)
						names = append(names, k)
					} else {
						return nil, nil, err
					}
				} else {
					return nil, nil, fmt.Errorf(unMsgInvalidElement, name)
				}
			}
			return result, names, nil
		} else if vs, ok := v.(map[any]any); ok {
			// because yaml unmarshals differently
			result := make([]T, 0, len(vs))
			names := make([]string, 0, len(vs))
			for k, v := range vs {
				if vm, ok := v.(map[string]any); ok {
					if item, err := fromObj[T](vm); err == nil {
						result = append(result, *item)
						names = append(names, fmt.Sprintf("%v", k))
					} else {
						return nil, nil, err
					}
				} else {
					return nil, nil, fmt.Errorf(unMsgInvalidElement, name)
				}
			}
			return result, names, nil
		} else {
			return nil, nil, fmt.Errorf(unMsgMustBeObject, name)
		}
	}
	return nil, nil, nil
}

// UnmarshalStrictRef is a global setting for Definition unmarshalling
// and, if set to true, will cause errors when an object (e.g. schema, etc.) in the OAS spec
// has a "$ref" property along with other properties
var UnmarshalStrictRef = false

func hasRef(m map[string]any) (string, bool, error) {
	var err error
	if v, ok := m[tags.Ref]; ok {
		if UnmarshalStrictRef && len(m) > 1 {
			return "", false, fmt.Errorf(`invalid spec - object has both %q and other properties`, tags.Ref)
		}
		if vs, ok := v.(string); ok {
			return vs, true, nil
		} else {
			err = fmt.Errorf(unMsgMustBeString, tags.Ref)
		}
	}
	return "", false, err
}

func stringFromProperty(m map[string]any, name string) (s string, err error) {
	if v, ok := m[name]; ok {
		if vs, ok := v.(string); ok {
			return vs, nil
		}
		err = fmt.Errorf(unMsgMustBeString, name)
	}
	return "", err
}

func jsonNumberFromProperty(m map[string]any, name string) (jn json.Number, err error) {
	if v, ok := m[name]; ok {
		switch vt := v.(type) {
		case json.Number:
			return vt, nil
		case string:
			return json.Number(vt), nil
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return json.Number(fmt.Sprintf("%v", vt)), nil
		case float32:
			if !math.IsNaN(float64(vt)) && !math.IsInf(float64(vt), 0) {
				return json.Number(fmt.Sprintf("%v", vt)), nil
			}
		case float64:
			if !math.IsNaN(vt) && !math.IsInf(vt, 0) {
				return json.Number(fmt.Sprintf("%v", vt)), nil
			}
		}
		return "", fmt.Errorf(unMsgInvalidValue, name)
	}
	return "", err
}

func uintFromProperty(m map[string]any, name string) (uint, error) {
	if v, ok := m[name]; ok {
		switch vt := v.(type) {
		case uint:
			return vt, nil
		case uint8:
			return uint(vt), nil
		case uint16:
			return uint(vt), nil
		case uint32:
			return uint(vt), nil
		case uint64:
			return uint(vt), nil
		case int:
			if vt >= 0 {
				return uint(vt), nil
			}
		case int8:
			if vt >= 0 {
				return uint(vt), nil
			}
		case int16:
			if vt >= 0 {
				return uint(vt), nil
			}
		case int32:
			if vt >= 0 {
				return uint(vt), nil
			}
		case int64:
			if vt >= 0 {
				return uint(vt), nil
			}
		case json.Number:
			if n, err := strconv.ParseUint(vt.String(), 0, 64); err == nil {
				return uint(n), nil
			}
		case string:
			if n, err := strconv.ParseUint(vt, 0, 64); err == nil {
				return uint(n), nil
			}
		case float32:
			if !math.IsNaN(float64(vt)) && !math.IsInf(float64(vt), 0) && vt >= 0 {
				return uint(vt), nil
			}
		case float64:
			if !math.IsNaN(vt) && !math.IsInf(vt, 0) && vt >= 0 {
				return uint(vt), nil
			}
		}
		return 0, fmt.Errorf(unMsgInvalidValue, name)
	}
	return 0, nil
}

func stringsSliceFromProperty(m map[string]any, name string) (s []string, err error) {
	if v, ok := m[name]; ok {
		if vs, ok := v.([]any); ok {
			s = make([]string, len(vs))
			for i, iv := range vs {
				if ivs, ok := iv.(string); ok {
					s[i] = ivs
				} else {
					return nil, fmt.Errorf(unMsgInvalidElement, name)
				}
			}
			return s, nil
		}
		err = fmt.Errorf(unMsgMustBeArray, name)
	}
	return nil, err
}

func anySliceFromProperty(m map[string]any, name string) (s []any, err error) {
	if v, ok := m[name]; ok {
		if vs, ok := v.([]any); ok {
			return vs, nil
		}
		err = fmt.Errorf(unMsgMustBeArray, name)
	}
	return nil, err
}

func booleanFromProperty(m map[string]any, name string) (b bool, err error) {
	if v, ok := m[name]; ok {
		if vb, ok := v.(bool); ok {
			return vb, nil
		} else if vs, ok := v.(string); ok && (vs == "true" || vs == "false") {
			return vs == "true", nil
		}
		err = fmt.Errorf(unMsgInvalidValue, name)
	}
	return false, err
}

func extensionsFrom(m map[string]any) Extensions {
	result := make(Extensions)
	for k, v := range m {
		if strings.HasPrefix(k, "x-") {
			result[k[2:]] = v
		}
	}
	return result
}

func (i *Info) unmarshalObj(m map[string]any) (err error) {
	i.Extensions = extensionsFrom(m)
	if i.Title, err = stringFromProperty(m, tags.Title); err == nil {
		if i.Description, err = stringFromProperty(m, tags.Description); err == nil {
			if i.Version, err = stringFromProperty(m, tags.Version); err == nil {
				if i.TermsOfService, err = stringFromProperty(m, tags.TermsOfService); err == nil {
					if i.Contact, err = objFromProperty[Contact](m, tags.Contact); err == nil {
						i.License, err = objFromProperty[License](m, tags.License)
					}
				}
			}
		}
	}
	return err
}

func (c *Contact) unmarshalObj(m map[string]any) (err error) {
	c.Extensions = extensionsFrom(m)
	if c.Name, err = stringFromProperty(m, tags.Name); err == nil {
		if c.Url, err = stringFromProperty(m, tags.Url); err == nil {
			c.Email, err = stringFromProperty(m, tags.Email)
		}
	}
	return err
}

func (l *License) unmarshalObj(m map[string]any) (err error) {
	l.Extensions = extensionsFrom(m)
	if l.Name, err = stringFromProperty(m, tags.Name); err == nil {
		l.Url, err = stringFromProperty(m, tags.Url)
	}
	return err
}

func (x *ExternalDocs) unmarshalObj(m map[string]any) (err error) {
	x.Extensions = extensionsFrom(m)
	if x.Description, err = stringFromProperty(m, tags.Description); err == nil {
		x.Url, err = stringFromProperty(m, tags.Url)
	}
	return err
}

func (t *Tag) unmarshalObj(m map[string]any) (err error) {
	t.Extensions = extensionsFrom(m)
	if t.Name, err = stringFromProperty(m, tags.Name); err == nil {
		if t.Description, err = stringFromProperty(m, tags.Description); err == nil {
			t.ExternalDocs, err = objFromProperty[ExternalDocs](m, tags.ExternalDocs)
		}
	}
	return err
}

func serversFrom(m map[string]any) (Servers, error) {
	if v, ok := m[tags.Servers]; ok {
		if vs, ok := v.([]any); ok {
			result := make(Servers, len(vs))
			for _, sv := range vs {
				if svm, ok := sv.(map[string]any); ok {
					if url, err := stringFromProperty(svm, tags.Url); err == nil {
						if s, err := fromObj[Server](svm); err == nil {
							result[url] = *s
						} else {
							return nil, err
						}
					} else {
						return nil, err
					}
				} else {
					return nil, fmt.Errorf(unMsgInvalidElement, tags.Servers)
				}
			}
			return result, nil
		} else {
			return nil, fmt.Errorf(unMsgMustBeArray, tags.Servers)
		}
	}
	return nil, nil
}

func (s *Server) unmarshalObj(m map[string]any) (err error) {
	s.Extensions = extensionsFrom(m)
	s.Description, err = stringFromProperty(m, tags.Description)
	return err
}

func securityFrom(m map[string]any) (SecuritySchemes, error) {
	if s, ok := m[tags.Security]; ok {
		if sv, ok := s.([]any); ok {
			result := make(SecuritySchemes, 0)
			for _, i := range sv {
				if im, ok := i.(map[string]any); ok {
					for k, v := range im {
						sec := &SecurityScheme{Name: k, Scopes: make([]string, 0)}
						if scopes, ok := v.([]any); ok {
							for _, str := range scopes {
								if scope, ok := str.(string); ok {
									sec.Scopes = append(sec.Scopes, scope)
								} else {
									return nil, fmt.Errorf(unMsgInvalidElement, k)
								}
							}
						} else {
							return nil, fmt.Errorf(unMsgInvalidElement, k)
						}
						result = append(result, *sec)
					}
				} else {
					return nil, fmt.Errorf(unMsgInvalidElement, tags.Security)
				}
			}
			return result, nil
		} else {
			return nil, fmt.Errorf(unMsgMustBeArray, tags.Security)
		}
	}
	return nil, nil
}

func componentsFrom(m map[string]any) (*Components, error) {
	if v, ok := m[tags.Components]; ok {
		if mv, ok := v.(map[string]any); ok {
			result := &Components{Extensions: extensionsFrom(mv)}
			if items, names, err := namedSliceFromProperty[Schema](mv, tags.Schemas); err == nil {
				for i, _ := range items {
					items[i].Name = names[i]
				}
				result.Schemas = items
			} else {
				return nil, err
			}
			if items, names, err := namedSliceFromProperty[SecurityScheme](mv, tags.SecuritySchemes); err == nil {
				for i, _ := range items {
					items[i].Name = names[i]
				}
				result.SecuritySchemes = items
			} else {
				return nil, err
			}
			if items, names, err := namedSliceFromProperty[Example](mv, tags.Examples); err == nil {
				for i, _ := range items {
					items[i].Name = names[i]
				}
				result.Examples = items
			} else {
				return nil, err
			}
			if items, names, err := namedSliceFromProperty[CommonParameter](mv, tags.Parameters); err == nil {
				result.Parameters = make(CommonParameters, len(items))
				for i, _ := range items {
					result.Parameters[names[i]] = items[i]
				}
			} else {
				return nil, err
			}
			if items, names, err := namedSliceFromProperty[Request](mv, tags.RequestBodies); err == nil {
				result.Requests = make(CommonRequests, len(items))
				for i, _ := range items {
					result.Requests[names[i]] = items[i]
				}
			} else {
				return nil, err
			}
			if items, names, err := namedSliceFromProperty[Response](mv, tags.Responses); err == nil {
				result.Responses = make(CommonResponses, len(items))
				for i, _ := range items {
					result.Responses[names[i]] = items[i]
				}
			} else {
				return nil, err
			}
			return result, nil
		} else {
			return nil, fmt.Errorf(unMsgMustBeObject, tags.Components)
		}
	}
	return nil, nil
}

func (s *SecurityScheme) unmarshalObj(m map[string]any) (err error) {
	s.Extensions = extensionsFrom(m)
	if s.Description, err = stringFromProperty(m, tags.Description); err == nil {
		if s.Type, err = stringFromProperty(m, tags.Type); err == nil {
			if s.Scheme, err = stringFromProperty(m, tags.Scheme); err == nil {
				if s.ParamName, err = stringFromProperty(m, tags.Name); err == nil {
					s.In, err = stringFromProperty(m, tags.In)
				}
			}
		}
	}
	return err
}

func (eg *Example) unmarshalObj(m map[string]any) (err error) {
	var ref string
	var ok bool
	if ref, ok, err = hasRef(m); ok {
		eg.ExampleRef = ref
	} else if err == nil {
		eg.Extensions = extensionsFrom(m)
		if eg.Summary, err = stringFromProperty(m, tags.Summary); err == nil {
			if eg.Description, err = stringFromProperty(m, tags.Description); err == nil {
				eg.Value = m[tags.Value]
			}
		}
	}
	return err
}

func (p *CommonParameter) unmarshalObj(m map[string]any) (err error) {
	p.Extensions = extensionsFrom(m)
	if p.Name, err = stringFromProperty(m, tags.Name); err == nil {
		if p.Description, err = stringFromProperty(m, tags.Description); err == nil {
			if p.Required, err = booleanFromProperty(m, tags.Required); err == nil {
				if p.In, err = stringFromProperty(m, tags.In); err == nil {
					p.Example = m[tags.Example]
					p.SchemaRef, p.Schema, err = schemaFrom(m)
				}
			}
		}
	}
	return err
}

func (p *QueryParam) unmarshalObj(m map[string]any) (err error) {
	var ref string
	var ok bool
	if ref, ok, err = hasRef(m); ok {
		p.Ref = ref
	} else if err == nil {
		p.Extensions = extensionsFrom(m)
		if p.Name, err = stringFromProperty(m, tags.Name); err == nil {
			if p.Description, err = stringFromProperty(m, tags.Description); err == nil {
				if p.Required, err = booleanFromProperty(m, tags.Required); err == nil {
					if p.In, err = stringFromProperty(m, tags.In); err == nil {
						p.Example = m[tags.Example]
						p.SchemaRef, p.Schema, err = schemaFrom(m)
					}
				}
			}
		}
	}
	return err
}

func schemaFrom(m map[string]any) (ref string, schema *Schema, err error) {
	if v, ok := m[tags.Schema]; ok {
		if vm, ok := v.(map[string]any); ok {
			ref, ok, err = hasRef(vm)
			if !ok && err == nil {
				schema, err = fromObj[Schema](vm)
			}
		} else {
			err = fmt.Errorf(unMsgMustBeObject, tags.Schema)
		}
	}
	return ref, schema, err
}

func (s *Schema) unmarshalObj(m map[string]any) (err error) {
	var ref string
	var ok bool
	if ref, ok, err = hasRef(m); ok {
		s.SchemaRef = ref
	} else if err == nil {
		s.Extensions = extensionsFrom(m)
		if s.Name, err = stringFromProperty(m, tags.Name); err == nil {
			if s.Description, err = stringFromProperty(m, tags.Description); err == nil {
				if s.Type, err = stringFromProperty(m, tags.Type); err == nil {
					if s.Format, err = stringFromProperty(m, tags.Format); err == nil {
						if s.RequiredProperties, err = stringsSliceFromProperty(m, tags.Required); err == nil {
							if s.Properties, err = unmarshalProperties(m); err == nil {
								if s.Discriminator, err = objFromProperty[Discriminator](m, tags.Discriminator); err == nil {
									if s.Ofs, err = ofsFrom(m); err == nil {
										s.Default = m[tags.Default]
										s.Example = m[tags.Example]
										s.Enum, err = anySliceFromProperty(m, tags.Enum)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return err
}

func unmarshalProperties(m map[string]any) (ptys Properties, err error) {
	if _, ok := m[tags.Properties]; ok {
		var vs []Property
		var names []string
		if vs, names, err = namedSliceFromProperty[Property](m, tags.Properties); err == nil {
			ptys = make(Properties, len(vs))
			for i, _ := range vs {
				vs[i].Name = names[i]
				ptys[i] = vs[i]
			}
		}
	}
	return ptys, err
}

func (p *Property) unmarshalObj(m map[string]any) (err error) {
	var ref string
	var ok bool
	if ref, ok, err = hasRef(m); ok {
		p.SchemaRef = ref
	} else if err == nil {
		p.Extensions = extensionsFrom(m)
		if p.Name, err = stringFromProperty(m, tags.Name); err == nil {
			if p.Description, err = stringFromProperty(m, tags.Description); err == nil {
				if p.Type, err = stringFromProperty(m, tags.Type); err == nil {
					if p.ItemType, err = stringFromProperty(m, tags.ItemType); err == nil {
						if p.Required, err = booleanFromProperty(m, tags.Required); err == nil {
							if p.Format, err = stringFromProperty(m, tags.Format); err == nil {
								if p.Deprecated, err = booleanFromProperty(m, tags.Deprecated); err == nil {
									if p.Properties, err = unmarshalProperties(m); err == nil {
										p.Example = m[tags.Example]
										if p.Enum, err = anySliceFromProperty(m, tags.Enum); err == nil {
											err = p.unmarshalConstraints(m)
										}
									}
								}
							}
						}
					}

				}
			}
		}
	}
	return err
}

func (p *Property) unmarshalConstraints(m map[string]any) (err error) {
	if p.Constraints.Pattern, err = stringFromProperty(m, tags.Pattern); err == nil {
		if p.Constraints.Maximum, err = jsonNumberFromProperty(m, tags.Maximum); err == nil {
			if p.Constraints.Minimum, err = jsonNumberFromProperty(m, tags.Minimum); err == nil {
				if p.Constraints.ExclusiveMinimum, err = booleanFromProperty(m, tags.ExclusiveMinimum); err == nil {
					if p.Constraints.ExclusiveMaximum, err = booleanFromProperty(m, tags.ExclusiveMaximum); err == nil {
						if p.Constraints.Nullable, err = booleanFromProperty(m, tags.Nullable); err == nil {
							if p.Constraints.UniqueItems, err = booleanFromProperty(m, tags.UniqueItems); err == nil {
								if p.Constraints.MultipleOf, err = uintFromProperty(m, tags.MultipleOf); err == nil {
									if p.Constraints.MaxLength, err = uintFromProperty(m, tags.MaxLength); err == nil {
										if p.Constraints.MinLength, err = uintFromProperty(m, tags.MinLength); err == nil {
											if p.Constraints.MaxItems, err = uintFromProperty(m, tags.MaxItems); err == nil {
												if p.Constraints.MinItems, err = uintFromProperty(m, tags.MinItems); err == nil {
													if p.Constraints.MaxProperties, err = uintFromProperty(m, tags.MaxProperties); err == nil {
														p.Constraints.MinProperties, err = uintFromProperty(m, tags.MinProperties)
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return err
}

func ofsFrom(m map[string]any) (ofs *Ofs, err error) {
	var v any
	var ok bool
	var name string
	for _, tag := range []string{tags.AllOf, tags.OneOf, tags.AnyOf} {
		if v, ok = m[tag]; ok {
			ofs = &Ofs{
				OfType: (map[string]OfType{
					tags.AllOf: AllOf,
					tags.OneOf: OneOf,
					tags.AnyOf: AnyOf,
				})[tag],
			}
			name = tag
			break
		}
	}
	if ofs != nil {
		if vs, ok := v.([]any); ok {
			for _, vv := range vs {
				if vm, ok := vv.(map[string]any); ok {
					var ref string
					if ref, ok, err = hasRef(vm); ok {
						ofs.Of = append(ofs.Of, &Of{SchemaRef: ref})
					} else if err != nil {
						break
					} else {
						var schema *Schema
						if schema, err = fromObj[Schema](vm); err == nil {
							ofs.Of = append(ofs.Of, &Of{SchemaDef: schema})
						} else {
							break
						}
					}
				} else {
					err = fmt.Errorf(unMsgInvalidElement, name)
					break
				}
			}
		} else {
			err = fmt.Errorf(unMsgMustBeArray, name)
		}
	}
	return ofs, err
}

func (d *Discriminator) unmarshalObj(m map[string]any) (err error) {
	d.Extensions = extensionsFrom(m)
	if d.PropertyName, err = stringFromProperty(m, tags.PropertyName); err == nil {
		if v, ok := m[tags.Mapping]; ok {
			if mm, ok := v.(map[string]any); ok {
				d.Mapping = make(map[string]string, len(mm))
				for k, mmv := range mm {
					if mmvs, ok := mmv.(string); ok {
						d.Mapping[k] = mmvs
					} else {
						err = fmt.Errorf(unMsgInvalidValue, tags.Discriminator+"."+tags.Mapping)
					}
				}
			} else {
				err = fmt.Errorf(unMsgMustBeObject, tags.Discriminator+"."+tags.Mapping)
			}
		}
	}
	return err
}

func (r *Request) unmarshalObj(m map[string]any) (err error) {
	var ref string
	var ok bool
	if ref, ok, err = hasRef(m); ok {
		r.Ref = ref
	} else if err == nil {
		r.Extensions = extensionsFrom(m)
		if r.Examples, err = sliceFromProperty[Example](m, tags.Examples); err == nil {
			if r.Description, err = stringFromProperty(m, tags.Description); err == nil {
				if r.Required, err = booleanFromProperty(m, tags.Required); err == nil {
					err = r.unmarshalContent(m)
				}
			}
		}
	}
	return err
}

func (r *Request) unmarshalContent(m map[string]any) (err error) {
	var cts []contentType
	var names []string
	if cts, names, err = namedSliceFromProperty[contentType](m, tags.Content); err == nil {
		if len(cts) == 1 {
			r.ContentType = names[0]
			r.IsArray = cts[0].isArray
			r.Schema = cts[0].schema
			r.SchemaRef = cts[0].ref
			r.Examples = cts[0].examples
			r.Extensions = cts[0].extensions
		} else {
			r.AlternativeContentTypes = ContentTypes{}
			jat := slices.Index(names, contentTypeJson)
			if jat == -1 {
				jat = 0
			}
			for i, ct := range cts {
				if i == jat {
					r.ContentType = names[i]
					r.IsArray = ct.isArray
					r.Schema = ct.schema
					r.SchemaRef = ct.ref
					r.Examples = ct.examples
					r.Extensions = ct.extensions
				} else {
					r.AlternativeContentTypes[names[i]] = ContentType{
						Schema:     ct.schema,
						SchemaRef:  ct.ref,
						IsArray:    ct.isArray,
						Examples:   ct.examples,
						Extensions: ct.extensions,
					}
				}
			}
		}
	}
	return err
}

func (r *Response) unmarshalObj(m map[string]any) (err error) {
	var ref string
	var ok bool
	if ref, ok, err = hasRef(m); ok {
		r.Ref = ref
	} else if err == nil {
		r.Extensions = extensionsFrom(m)
		if r.Examples, err = sliceFromProperty[Example](m, tags.Examples); err == nil {
			if r.Description, err = stringFromProperty(m, tags.Description); err == nil {
				err = r.unmarshalContent(m)
			}
		}
	}
	return err
}

func (r *Response) unmarshalContent(m map[string]any) (err error) {
	var cts []contentType
	var names []string
	if cts, names, err = namedSliceFromProperty[contentType](m, tags.Content); err == nil {
		if len(cts) == 1 {
			r.ContentType = names[0]
			r.IsArray = cts[0].isArray
			r.Schema = cts[0].schema
			r.SchemaRef = cts[0].ref
			r.Examples = cts[0].examples
			r.Extensions = cts[0].extensions
		} else {
			r.AlternativeContentTypes = ContentTypes{}
			jat := slices.Index(names, contentTypeJson)
			if jat == -1 {
				jat = 0
			}
			for i, ct := range cts {
				if i == jat {
					r.ContentType = names[i]
					r.IsArray = ct.isArray
					r.Schema = ct.schema
					r.SchemaRef = ct.ref
					r.Examples = ct.examples
					r.Extensions = ct.extensions
				} else {
					r.AlternativeContentTypes[names[i]] = ContentType{
						Schema:     ct.schema,
						SchemaRef:  ct.ref,
						IsArray:    ct.isArray,
						Examples:   ct.examples,
						Extensions: ct.extensions,
					}
				}
			}
		}
	}
	return err
}

type contentType struct {
	isArray    bool
	xType      string
	ref        string
	schema     *Schema
	extensions Extensions
	examples   Examples
}

func (ct *contentType) unmarshalObj(m map[string]any) (err error) {
	ct.extensions = extensionsFrom(m)
	if items, names, err := namedSliceFromProperty[Example](m, tags.Examples); err == nil {
		for i, _ := range items {
			items[i].Name = names[i]
		}
		ct.examples = items
	} else {
		return err
	}
	if s, ok := m[tags.Schema]; ok {
		if sm, ok := s.(map[string]any); ok {
			if ct.xType, err = stringFromProperty(sm, tags.Type); err == nil {
				v, ok := sm[tags.Items]
				if ok {
					if ct.xType != values.TypeArray {
						err = fmt.Errorf(`property %q contains property %q when type is not %q`, tags.Schema, tags.Items, values.TypeArray)
					} else if im, ok := v.(map[string]any); ok {
						ct.isArray = true
						sm = im
					} else {
						err = fmt.Errorf(unMsgMustBeObject, tags.Items)
					}
				} else if ct.xType == values.TypeArray {
					err = fmt.Errorf(`property %q must contain property %q when type is %q`, tags.Schema, tags.Items, values.TypeArray)
				}
				if err == nil {
					var ref string
					if ref, ok, err = hasRef(sm); ok {
						ct.ref = ref
					} else {
						ct.schema, err = fromObj[Schema](sm)
					}
				}
			}
		} else {
			err = fmt.Errorf(unMsgMustBeObject, tags.Schema)
		}
	} else {
		err = fmt.Errorf(`property %q value is missing "schema" property`, tags.Content)
	}
	return err
}
