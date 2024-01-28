package chioas

import (
	"github.com/go-andiamo/chioas/yaml"
	"github.com/go-andiamo/urit"
	"golang.org/x/exp/slices"
	"net/http"
	"sort"
	"strings"
)

var defaultResponses = Responses{
	http.StatusOK: {},
}

// Methods is a map of Method (where the key is a http.Method)
type Methods map[string]Method

// Method represents the definition of a method (as used by Path and, for root methods, Definition)
type Method struct {
	// Description is the OAS description of the method
	Description string
	// Summary is the OAS summary of the method
	Summary string
	// Handler is the http.HandlerFunc to be used by Chi
	//
	// Can also be specified as a string - which must be a public method on the interface passed to Definition.SetupRoutes
	//
	// Can also be specified as a method expression - e.g
	//   Handler: (*myApi).GetSomething
	// (the method specified must be a public method on the interface passed to Definition.SetupRoutes
	//
	// Can also be specified as a GetHandler func - which is called to acquire the http.HandlerFunc
	Handler any
	// OperationId is the OAS operation id of the method
	//
	// This can be overridden by providing a DocOptions.OperationIdentifier
	OperationId string
	// Tag is the OAS tag of the method
	//
	// If this is an empty string and any ancestor Path.Tag is set then that ancestor tag is used
	Tag string
	// QueryParams is the OAS query params for the method
	//
	// Can also be used to specify header params (see QueryParam.In)
	QueryParams QueryParams
	// Request is the optional OAS request body for the method
	Request *Request
	// Responses is the OAS responses for the method
	//
	// If no responses are specified, the DocOptions.DefaultResponses is used
	//
	// If there are no DocOptions.DefaultResponses specified, then a http.StatusOK response is used
	Responses Responses
	// Deprecated is the OAS deprecated flag for the method
	Deprecated bool
	// Security is the OAS security schemes used by the method
	Security SecuritySchemes
	// OptionalSecurity if set to true, adds an entry to the OAS method security e.g.
	//  security:
	//   - {}
	OptionalSecurity bool
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
	// HideDocs if set to true, hides this method from the OAS docs
	HideDocs bool
}

// MethodsOrder defines the order in which methods appear in docs
var MethodsOrder = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodOptions,
	http.MethodConnect,
	http.MethodTrace,
}

func compareMethods(ma, mb string) bool {
	a := slices.Index(MethodsOrder, ma)
	b := slices.Index(MethodsOrder, mb)
	if a == -1 && b == -1 {
		return ma < mb
	} else if a == -1 {
		return false
	} else if b == -1 {
		return true
	}
	return a < b
}

func (m Methods) sorted(add ...string) (result []string) {
	result = make([]string, 0, len(m)+len(add))
	for k := range m {
		result = append(result, k)
	}
	for _, k := range add {
		if _, has := m[k]; !has {
			result = append(result, k)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return compareMethods(result[i], result[j])
	})
	return
}

func (m Methods) hasVisibleMethods(opts *DocOptions) bool {
	result := false
	for n, mDef := range m {
		if result = !mDef.HideDocs && (n != http.MethodHead || !opts.HideHeadMethods); result {
			break
		}
	}
	return result
}

func (m Methods) getWithoutHead() (getM Method, has bool) {
	if getM, has = m[http.MethodGet]; has {
		if _, hasHd := m[http.MethodHead]; hasHd {
			has = false
		}
	}
	return
}

func (m Methods) hasOptions() bool {
	_, has := m[http.MethodOptions]
	return has
}

func (m Methods) writeYaml(opts *DocOptions, autoHeads bool, autoOptions bool, template urit.Template, knownParams PathParams, parentTag string, w yaml.Writer) {
	type sortMethod struct {
		name   string
		method Method
	}
	sortedMethods := make([]sortMethod, 0, len(m))
	for n, mDef := range m {
		if !mDef.HideDocs && (n != http.MethodHead || !opts.HideHeadMethods) {
			sortedMethods = append(sortedMethods, sortMethod{
				name:   n,
				method: mDef,
			})
		}
	}
	if !opts.HideHeadMethods && autoHeads {
		if getM, has := m.getWithoutHead(); has {
			sortedMethods = append(sortedMethods, sortMethod{
				name:   http.MethodHead,
				method: getM,
			})
		}
	}
	if !opts.HideAutoOptionsMethods && autoOptions && !m.hasOptions() {
		sortedMethods = append(sortedMethods, sortMethod{
			name: http.MethodOptions,
			method: Method{
				Responses: Responses{
					http.StatusOK: {
						NoContent: true,
					},
				},
			},
		})
	}
	sort.Slice(sortedMethods, func(i, j int) bool {
		return compareMethods(sortedMethods[i].name, sortedMethods[j].name)
	})
	var pathVars []urit.PathVar
	if template != nil {
		pathVars = template.Vars()
	}
	for _, mn := range sortedMethods {
		mn.method.writeYaml(opts, template, pathVars, knownParams, parentTag, mn.name, w)
	}
}

func (m Method) writeYaml(opts *DocOptions, template urit.Template, pathVars []urit.PathVar, knownParams PathParams, parentTag string, method string, w yaml.Writer) {
	w.WriteTagStart(strings.ToLower(method)).
		WriteComments(m.Comment).
		WriteTagValue(tagNameSummary, m.Summary).
		WriteTagValue(tagNameDescription, m.Description).
		WriteTagValue(tagNameOperationId, m.getOperationId(opts, method, template, parentTag)).
		WriteTagValue(tagNameDeprecated, nilBool(m.Deprecated))
	if m.OptionalSecurity || len(m.Security) > 0 {
		w.WriteTagStart(tagNameSecurity)
		if m.OptionalSecurity {
			w.WriteItem(yaml.LiteralValue{Value: "{}"})
		}
		m.Security.writeYaml(w, true)
		w.WriteTagEnd()
	}
	if tag := defaultTag(parentTag, m.Tag); tag != "" {
		w.WriteTagStart(tagNameTags).
			WriteItem(tag).
			WriteTagEnd()
	}
	m.writeParams(pathVars, knownParams, w)
	if m.Request != nil {
		m.Request.writeYaml(w)
	}
	if m.Responses != nil && len(m.Responses) > 0 {
		m.Responses.writeYaml(method == http.MethodHead, w)
	} else if opts != nil && opts.DefaultResponses != nil && len(opts.DefaultResponses) > 0 {
		opts.DefaultResponses.writeYaml(method == http.MethodHead, w)
	} else {
		// no responses - needs something...
		defaultResponses.writeYaml(method == http.MethodHead, w)
	}
	writeExtensions(m.Extensions, w)
	writeAdditional(m.Additional, m, w)
	w.WriteTagEnd()
}

func (m Method) getOperationId(opts *DocOptions, method string, template urit.Template, parentTag string) string {
	if opts.OperationIdentifier != nil {
		path := "/"
		if template != nil {
			path = template.OriginalTemplate()
		}
		return defValue(opts.OperationIdentifier(m, method, path, parentTag), m.OperationId)
	} else {
		return m.OperationId
	}
}

func (m Method) writeParams(pathVars []urit.PathVar, knownParams PathParams, w yaml.Writer) {
	if has, pathParams := m.hasParams(pathVars, knownParams); has {
		w.WriteTagStart(tagNameParameters)
		m.writePathParams(pathParams, w)
		m.writeQueryParams(w)
		w.WriteTagEnd()
	}
}

func (m Method) hasParams(pathVars []urit.PathVar, knownParams PathParams) (bool, []pathParamHolder) {
	result := len(m.QueryParams) > 0
	pathParams := make([]pathParamHolder, 0)
	for _, pv := range pathVars {
		if pp, ok := knownParams[pv.Name]; ok {
			pathParams = append(pathParams, pathParamHolder{
				name: pv.Name,
				def:  pp,
			})
		} else {
			pathParams = append(pathParams, pathParamHolder{
				name: pv.Name,
			})
		}
	}
	result = result || len(pathParams) > 0
	return result, pathParams
}

type pathParamHolder struct {
	name string
	def  PathParam
}

func (m Method) writePathParams(pathParams []pathParamHolder, w yaml.Writer) {
	for _, pp := range pathParams {
		pp.def.writeYaml(pp.name, w)
	}
}

func (m Method) writeQueryParams(w yaml.Writer) {
	m.QueryParams.writeYaml(w)
}
