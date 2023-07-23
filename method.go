package chioas

import (
	"fmt"
	"github.com/go-andiamo/chioas/yaml"
	"github.com/go-andiamo/urit"
	"golang.org/x/exp/slices"
	"net/http"
	"reflect"
	"sort"
	"strings"
)

var defaultResponses = Responses{
	http.StatusOK: {},
}

type Methods map[string]Method

type Method struct {
	Description string
	Summary     string
	Handler     any // can be a http.HandlerFunc or a string method name
	OperationId string
	Tag         string
	QueryParams QueryParams
	Request     *Request
	Responses   Responses
	Additional  Additional
	HideDocs    bool // hides this method from docs
}

func (m Method) getHandler(path string, method string, thisApi any) (http.HandlerFunc, error) {
	if m.Handler == nil {
		return nil, fmt.Errorf("handler not set (path: %s, method: %s)", path, method)
	} else if hf, ok := m.Handler.(func(http.ResponseWriter, *http.Request)); ok {
		return hf, nil
	} else if mn, ok := m.Handler.(string); ok {
		if thisApi == nil {
			return nil, fmt.Errorf("method by name '%s' can only be used when 'thisApi' arg is passed to Definition.SetupRoutes (path: %s, method: %s)", mn, path, method)
		}
		mfn := reflect.ValueOf(thisApi).MethodByName(mn)
		if !mfn.IsValid() {
			return nil, fmt.Errorf("method name '%s' does not exist (path: %s, method: %s)", mn, path, method)
		}
		if hf, ok = mfn.Interface().(func(http.ResponseWriter, *http.Request)); ok {
			return hf, nil
		}
		return nil, fmt.Errorf("method name '%s' is not http.HandlerFunc (path: %s, method: %s)", mn, path, method)
	}
	return nil, fmt.Errorf("invalid handler type (path: %s, method: %s)", path, method)
}

// MethodsOrder defines the order in which methods appear in docs
var MethodsOrder = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodOptions,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

func methodOrder(m string) int {
	order := slices.Index(MethodsOrder, m)
	if order == -1 {
		order = len(MethodsOrder)
	}
	return order
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

func (m Methods) writeYaml(opts *DocOptions, autoHeads bool, template urit.Template, knownParams PathParams, parentTag string, w yaml.Writer) {
	type sortMethod struct {
		name   string
		method Method
		order  int
	}
	sortedMethods := make([]sortMethod, 0, len(m))
	for n, mDef := range m {
		if !mDef.HideDocs && (n != http.MethodHead || !opts.HideHeadMethods) {
			sortedMethods = append(sortedMethods, sortMethod{
				name:   n,
				method: mDef,
				order:  methodOrder(n),
			})
		}
	}
	if !opts.HideHeadMethods && autoHeads {
		if getM, has := m.getWithoutHead(); has {
			sortedMethods = append(sortedMethods, sortMethod{
				name:   http.MethodHead,
				method: getM,
				order:  methodOrder(http.MethodHead),
			})
		}
	}
	sort.Slice(sortedMethods, func(i, j int) bool {
		return sortedMethods[i].order < sortedMethods[j].order
	})
	var pathVars []urit.PathVar
	if template != nil {
		pathVars = template.Vars()
	}
	for _, mn := range sortedMethods {
		mn.method.writeYaml(opts, pathVars, knownParams, parentTag, mn.name, w)
	}
}

func (m Method) writeYaml(opts *DocOptions, pathVars []urit.PathVar, knownParams PathParams, parentTag string, method string, w yaml.Writer) {
	w.WriteTagStart(strings.ToLower(method)).
		WriteTagValue(tagNameSummary, m.Summary).
		WriteTagValue(tagNameDescription, m.Description).
		WriteTagValue(tagNameOperationId, m.OperationId)
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
		m.Responses.writeYaml(w)
	} else if opts != nil && opts.DefaultResponses != nil && len(opts.DefaultResponses) > 0 {
		opts.DefaultResponses.writeYaml(w)
	} else {
		// no responses - needs something...
		defaultResponses.writeYaml(w)
	}
	if m.Additional != nil {
		m.Additional.Write(m, w)
	}
	w.WriteTagEnd()
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
