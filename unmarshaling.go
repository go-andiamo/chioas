package chioas

import (
	"encoding/json"
	"fmt"
	"github.com/go-andiamo/splitter"
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
	if v, err := objFromProperty[Info](m, tagNameInfo); err != nil {
		return err
	} else if v != nil {
		d.Info = *v
	}
	if v, err := objFromProperty[ExternalDocs](m, tagNameExternalDocs); err != nil {
		return err
	} else if v != nil {
		d.Info.ExternalDocs = v
	}
	if d.Tags, err = sliceFromProperty[Tag](m, tagNameTags); err == nil {
		if d.Servers, err = serversFrom(m); err == nil {
			if d.Security, err = securityFrom(m); err == nil {
				if d.Components, err = componentsFrom(m); err == nil {
					err = d.unmarshalPaths(m)
				}
			}
		}
	}
	return err
}

func (d *Definition) unmarshalPaths(m map[string]any) (err error) {
	if v, ok := m[tagNamePaths]; ok {
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
			err = fmt.Errorf(unMsgMustBeObject, tagNamePaths)
		}
	}
	return err
}

var pathSplitter = splitter.MustCreateSplitter('/', splitter.CurlyBrackets).AddDefaultOptions(splitter.IgnoreEmptyFirst, splitter.IgnoreEmptyLast)

func unflattenHolders(holders []*pathHolder) (result map[string]*pathHolder, err error) {
	result = make(map[string]*pathHolder, len(holders))
	add := func(h *pathHolder) error {
		if parts, err := pathSplitter.Split(h.path); err == nil {
			if len(parts) == 1 {
				if curr, ok := result["/"+parts[0]]; ok {
					curr.path = "/" + parts[0]
				} else {
					result["/"+parts[0]] = h
				}
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
			for _, holder := range ph.subs {
				if result.Paths[holder.path], err = holder.getPath(); err != nil {
					return
				}
			}
		}
	}
	return result, err
}

func (ph *pathHolder) getMethods() (Methods, error) {
	if ph.methods == nil && ph.obj != nil {
		// build methods...
		ph.methods = make(Methods, len(ph.obj))
		for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodConnect, http.MethodTrace} {
			var v any
			var ok bool
			if v, ok = ph.obj[strings.ToLower(method)]; !ok {
				v, ok = ph.obj[method]
			}
			if ok {
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
	if m.Description, err = stringFromProperty(o, tagNameDescription); err == nil {
		if m.Summary, err = stringFromProperty(o, tagNameSummary); err == nil {
			if m.OperationId, err = stringFromProperty(o, tagNameOperationId); err == nil {
				if m.Deprecated, err = booleanFromProperty(o, tagNameDeprecated); err == nil {
					var tags []string
					if tags, err = stringsSliceFromProperty(o, tagNameTags); err == nil {
						if len(tags) > 0 {
							m.Tag = tags[0]
						}
						if err = m.unmarshalSecurity(o); err == nil {
							if m.Request, err = objFromProperty[Request](o, tagNameRequestBody); err == nil {
								if m.QueryParams, err = sliceFromProperty[QueryParam](o, tagNameParameters); err == nil {
									var responses []Response
									var codes []string
									if responses, codes, err = namedSliceFromProperty[Response](o, tagNameResponses); err == nil && responses != nil {
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
	if secs, err = sliceFromProperty[methodSecurity](o, tagNameSecurity); err == nil {
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

func hasRef(m map[string]any) (string, bool, error) {
	var err error
	if v, ok := m[tagNameRef]; ok {
		if vs, ok := v.(string); ok {
			return vs, true, nil
		} else {
			err = fmt.Errorf(unMsgMustBeString, tagNameRef)
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
			return json.Number(fmt.Sprintf("%s", vt)), nil
		case float32, float64:
			return json.Number(fmt.Sprintf("%s", vt)), nil
		default:
			err = fmt.Errorf(unMsgInvalidValue, name)
		}
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
					err = fmt.Errorf(unMsgInvalidElement, name)
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

func isUnmarshaler(t any) (unmarshaler, bool) {
	u, ok := t.(unmarshaler)
	return u, ok
}

func (i *Info) unmarshalObj(m map[string]any) (err error) {
	i.Extensions = extensionsFrom(m)
	if i.Title, err = stringFromProperty(m, tagNameTitle); err == nil {
		if i.Description, err = stringFromProperty(m, tagNameDescription); err == nil {
			if i.Version, err = stringFromProperty(m, tagNameVersion); err == nil {
				if i.TermsOfService, err = stringFromProperty(m, tagNameTermsOfService); err == nil {
					if i.Contact, err = objFromProperty[Contact](m, tagNameContact); err == nil {
						i.License, err = objFromProperty[License](m, tagNameLicense)
					}
				}
			}
		}
	}
	return err
}

func (c *Contact) unmarshalObj(m map[string]any) (err error) {
	c.Extensions = extensionsFrom(m)
	if c.Name, err = stringFromProperty(m, tagNameName); err == nil {
		if c.Url, err = stringFromProperty(m, tagNameUrl); err == nil {
			c.Email, err = stringFromProperty(m, tagNameEmail)
		}
	}
	return err
}

func (l *License) unmarshalObj(m map[string]any) (err error) {
	l.Extensions = extensionsFrom(m)
	if l.Name, err = stringFromProperty(m, tagNameName); err == nil {
		l.Url, err = stringFromProperty(m, tagNameUrl)
	}
	return err
}

func (x *ExternalDocs) unmarshalObj(m map[string]any) (err error) {
	x.Extensions = extensionsFrom(m)
	if x.Description, err = stringFromProperty(m, tagNameDescription); err == nil {
		x.Url, err = stringFromProperty(m, tagNameUrl)
	}
	return err
}

func (t *Tag) unmarshalObj(m map[string]any) (err error) {
	t.Extensions = extensionsFrom(m)
	if t.Name, err = stringFromProperty(m, tagNameName); err == nil {
		if t.Description, err = stringFromProperty(m, tagNameDescription); err == nil {
			t.ExternalDocs, err = objFromProperty[ExternalDocs](m, tagNameExternalDocs)
		}
	}
	return err
}

func serversFrom(m map[string]any) (Servers, error) {
	if v, ok := m[tagNameServers]; ok {
		if vs, ok := v.([]any); ok {
			result := make(Servers, len(vs))
			for _, sv := range vs {
				if svm, ok := sv.(map[string]any); ok {
					if url, err := stringFromProperty(svm, tagNameUrl); err == nil {
						if s, err := fromObj[Server](svm); err == nil {
							result[url] = *s
						} else {
							return nil, err
						}
					} else {
						return nil, err
					}
				} else {
					return nil, fmt.Errorf(unMsgInvalidElement, tagNameServers)
				}
			}
			return result, nil
		} else {
			return nil, fmt.Errorf(unMsgMustBeArray, tagNameServers)
		}
	}
	return nil, nil
}

func (s *Server) unmarshalObj(m map[string]any) (err error) {
	s.Extensions = extensionsFrom(m)
	s.Description, err = stringFromProperty(m, tagNameDescription)
	return err
}

func securityFrom(m map[string]any) (SecuritySchemes, error) {
	if s, ok := m[tagNameSecurity]; ok {
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
					return nil, fmt.Errorf(unMsgInvalidElement, tagNameSecurity)
				}
			}
			return result, nil
		} else {
			return nil, fmt.Errorf(unMsgMustBeArray, tagNameSecurity)
		}
	}
	return nil, nil
}

func componentsFrom(m map[string]any) (*Components, error) {
	if v, ok := m[tagNameComponents]; ok {
		if mv, ok := v.(map[string]any); ok {
			result := &Components{Extensions: extensionsFrom(mv)}
			if items, names, err := namedSliceFromProperty[Schema](mv, tagNameSchemas); err == nil {
				for i, _ := range items {
					items[i].Name = names[i]
				}
				result.Schemas = items
			} else {
				return nil, err
			}
			if items, names, err := namedSliceFromProperty[SecurityScheme](mv, tagNameSecuritySchemes); err == nil {
				for i, _ := range items {
					items[i].Name = names[i]
				}
				result.SecuritySchemes = items
			} else {
				return nil, err
			}
			if items, names, err := namedSliceFromProperty[Example](mv, tagNameExamples); err == nil {
				for i, _ := range items {
					items[i].Name = names[i]
				}
				result.Examples = items
			} else {
				return nil, err
			}
			if items, names, err := namedSliceFromProperty[CommonParameter](mv, tagNameParameters); err == nil {
				result.Parameters = make(CommonParameters, len(items))
				for i, _ := range items {
					result.Parameters[names[i]] = items[i]
				}
			} else {
				return nil, err
			}
			if items, names, err := namedSliceFromProperty[Request](mv, tagNameRequestBodies); err == nil {
				result.Requests = make(CommonRequests, len(items))
				for i, _ := range items {
					result.Requests[names[i]] = items[i]
				}
			} else {
				return nil, err
			}
			if items, names, err := namedSliceFromProperty[Response](mv, tagNameResponses); err == nil {
				result.Responses = make(CommonResponses, len(items))
				for i, _ := range items {
					result.Responses[names[i]] = items[i]
				}
			} else {
				return nil, err
			}
			return result, nil
		} else {
			return nil, fmt.Errorf(unMsgMustBeObject, tagNameComponents)
		}
	}
	return nil, nil
}

func (s *SecurityScheme) unmarshalObj(m map[string]any) (err error) {
	s.Extensions = extensionsFrom(m)
	if s.Description, err = stringFromProperty(m, tagNameDescription); err == nil {
		if s.Type, err = stringFromProperty(m, tagNameType); err == nil {
			if s.Scheme, err = stringFromProperty(m, tagNameScheme); err == nil {
				if s.ParamName, err = stringFromProperty(m, tagNameName); err == nil {
					s.In, err = stringFromProperty(m, tagNameIn)
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
		if eg.Summary, err = stringFromProperty(m, tagNameSummary); err == nil {
			if eg.Description, err = stringFromProperty(m, tagNameDescription); err == nil {
				eg.Value = m[tagNameValue]
			}
		}
	}
	return err
}

func (p *CommonParameter) unmarshalObj(m map[string]any) (err error) {
	p.Extensions = extensionsFrom(m)
	if p.Name, err = stringFromProperty(m, tagNameName); err == nil {
		if p.Description, err = stringFromProperty(m, tagNameDescription); err == nil {
			if p.Required, err = booleanFromProperty(m, tagNameRequired); err == nil {
				if p.In, err = stringFromProperty(m, tagNameIn); err == nil {
					p.Example = m[tagNameExample]
					p.SchemaRef, p.Schema, err = schemaFrom(m)
				}
			}
		}
	}
	return err
}

func (p *QueryParam) unmarshalObj(m map[string]any) (err error) {
	p.Extensions = extensionsFrom(m)
	if p.Name, err = stringFromProperty(m, tagNameName); err == nil {
		if p.Description, err = stringFromProperty(m, tagNameDescription); err == nil {
			if p.Required, err = booleanFromProperty(m, tagNameRequired); err == nil {
				if p.In, err = stringFromProperty(m, tagNameIn); err == nil {
					p.Example = m[tagNameExample]
					p.SchemaRef, p.Schema, err = schemaFrom(m)
				}
			}
		}
	}
	return err
}

func schemaFrom(m map[string]any) (ref string, schema *Schema, err error) {
	if v, ok := m[tagNameSchema]; ok {
		if vm, ok := v.(map[string]any); ok {
			ref, ok, err = hasRef(vm)
			if !ok || err == nil {
				schema, err = fromObj[Schema](vm)
			}
		} else {
			err = fmt.Errorf(unMsgMustBeObject, tagNameSchema)
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
		if s.Name, err = stringFromProperty(m, tagNameName); err == nil {
			if s.Description, err = stringFromProperty(m, tagNameDescription); err == nil {
				if s.Type, err = stringFromProperty(m, tagNameType); err == nil {
					if s.Format, err = stringFromProperty(m, tagNameFormat); err == nil {
						if s.RequiredProperties, err = stringsSliceFromProperty(m, tagNameRequired); err == nil {
							if s.Properties, err = unmarshalProperties(m); err == nil {
								if s.Discriminator, err = objFromProperty[Discriminator](m, tagNameDiscriminator); err == nil {
									if s.Ofs, err = ofsFrom(m); err == nil {
										s.Default = m[tagNameDefault]
										s.Example = m[tagNameExample]
										s.Enum, err = anySliceFromProperty(m, tagNameEnum)
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
	if _, ok := m[tagNameProperties]; ok {
		var vs []Property
		var names []string
		if vs, names, err = namedSliceFromProperty[Property](m, tagNameProperties); err == nil {
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
		if p.Name, err = stringFromProperty(m, tagNameName); err == nil {
			if p.Description, err = stringFromProperty(m, tagNameDescription); err == nil {
				if p.Type, err = stringFromProperty(m, tagNameType); err == nil {
					if p.ItemType, err = stringFromProperty(m, tagNameItemType); err == nil {
						if p.Required, err = booleanFromProperty(m, tagNameRequired); err == nil {
							if p.Format, err = stringFromProperty(m, tagNameFormat); err == nil {
								if p.Deprecated, err = booleanFromProperty(m, tagNameDeprecated); err == nil {
									if p.Properties, err = unmarshalProperties(m); err == nil {
										p.Example = m[tagNameExample]
										if p.Enum, err = anySliceFromProperty(m, tagNameEnum); err == nil {
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
	if p.Constraints.Pattern, err = stringFromProperty(m, tagNamePattern); err == nil {
		if p.Constraints.Maximum, err = jsonNumberFromProperty(m, tagNameMaximum); err == nil {
			if p.Constraints.Minimum, err = jsonNumberFromProperty(m, tagNameMinimum); err == nil {
				if p.Constraints.ExclusiveMinimum, err = booleanFromProperty(m, tagNameExclusiveMinimum); err == nil {
					if p.Constraints.ExclusiveMaximum, err = booleanFromProperty(m, tagNameExclusiveMaximum); err == nil {
						if p.Constraints.Nullable, err = booleanFromProperty(m, tagNameNullable); err == nil {
							if p.Constraints.UniqueItems, err = booleanFromProperty(m, tagNameUniqueItems); err == nil {
								if p.Constraints.MultipleOf, err = uintFromProperty(m, tagNameMultipleOf); err == nil {
									if p.Constraints.MaxLength, err = uintFromProperty(m, tagNameMaxLength); err == nil {
										if p.Constraints.MinLength, err = uintFromProperty(m, tagNameMinLength); err == nil {
											if p.Constraints.MaxItems, err = uintFromProperty(m, tagNameMaxItems); err == nil {
												if p.Constraints.MinItems, err = uintFromProperty(m, tagNameMinItems); err == nil {
													if p.Constraints.MaxProperties, err = uintFromProperty(m, tagNameMaxProperties); err == nil {
														p.Constraints.MinProperties, err = uintFromProperty(m, tagNameMinProperties)
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
	var tagName string
	for _, tag := range []string{tagNameAllOf, tagNameOneOf, tagNameAnyOf} {
		if v, ok = m[tag]; ok {
			ofs = &Ofs{
				OfType: (map[string]OfType{
					tagNameAllOf: AllOf,
					tagNameOneOf: OneOf,
					tagNameAnyOf: AnyOf,
				})[tag],
			}
			tagName = tag
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
					err = fmt.Errorf(unMsgInvalidElement, tagName)
					break
				}
			}
		} else {
			err = fmt.Errorf(unMsgMustBeArray, tagName)
		}
	}
	return ofs, err
}

func (d *Discriminator) unmarshalObj(m map[string]any) (err error) {
	d.Extensions = extensionsFrom(m)
	if d.PropertyName, err = stringFromProperty(m, tagNamePropertyName); err == nil {
		if v, ok := m[tagNameMapping]; ok {
			if mm, ok := v.(map[string]any); ok {
				d.Mapping = make(map[string]string, len(mm))
				for k, mmv := range mm {
					if mmvs, ok := mmv.(string); ok {
						d.Mapping[k] = mmvs
					} else {
						err = fmt.Errorf(unMsgInvalidValue, tagNameDiscriminator+"."+tagNameMapping)
					}
				}
			} else {
				err = fmt.Errorf(unMsgMustBeObject, tagNameDiscriminator+"."+tagNameMapping)
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
		if r.Examples, err = sliceFromProperty[Example](m, tagNameExamples); err == nil {
			if r.Description, err = stringFromProperty(m, tagNameDescription); err == nil {
				if r.Required, err = booleanFromProperty(m, tagNameRequired); err == nil {
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
	if cts, names, err = namedSliceFromProperty[contentType](m, tagNameContent); err == nil {
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
						Schema:    ct.schema,
						SchemaRef: ct.ref,
						IsArray:   ct.isArray,
						Examples:  ct.examples,
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
		if r.Examples, err = sliceFromProperty[Example](m, tagNameExamples); err == nil {
			if r.Description, err = stringFromProperty(m, tagNameDescription); err == nil {
				err = r.unmarshalContent(m)
			}
		}
	}
	return err
}

func (r *Response) unmarshalContent(m map[string]any) (err error) {
	var cts []contentType
	var names []string
	if cts, names, err = namedSliceFromProperty[contentType](m, tagNameContent); err == nil {
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
						Schema:    ct.schema,
						SchemaRef: ct.ref,
						IsArray:   ct.isArray,
						Examples:  ct.examples,
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
	if items, names, err := namedSliceFromProperty[Example](m, tagNameExamples); err == nil {
		for i, _ := range items {
			items[i].Name = names[i]
		}
		ct.examples = items
	} else {
		return err
	}
	if s, ok := m[tagNameSchema]; ok {
		if sm, ok := s.(map[string]any); ok {
			if ct.xType, err = stringFromProperty(sm, tagNameType); err == nil {
				v, ok := sm[tagNameItems]
				if ok {
					if ct.xType != tagValueTypeArray {
						err = fmt.Errorf(`property %q contains property %q when type is not %q`, tagNameSchema, tagNameItems, tagValueTypeArray)
					} else if im, ok := v.(map[string]any); ok {
						ct.isArray = true
						sm = im
					} else {
						err = fmt.Errorf(unMsgMustBeObject, tagNameItems)
					}
				} else if ct.xType == tagValueTypeArray {
					err = fmt.Errorf(`property %q must contain property %q when type is %q`, tagNameSchema, tagNameItems, tagValueTypeArray)
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
			err = fmt.Errorf(unMsgMustBeObject, tagNameSchema)
		}
	} else {
		err = fmt.Errorf(`property %q value is missing "schema" property`, tagNameContent)
	}
	return err
}
