package chioas

import (
	"fmt"
	"golang.org/x/exp/slices"
	"net/http"
	"reflect"
	"sort"
	"strings"
)

type Methods map[string]Method

type Method struct {
	Description string
	Summary     string
	Handler     http.HandlerFunc
	MethodName  string
	OperationId string
	Tag         string
}

func (m Method) getHandler(this any) http.HandlerFunc {
	if m.Handler != nil {
		return m.Handler
	} else if m.MethodName != "" {
		if this != nil {
			mfn := reflect.ValueOf(this).MethodByName(m.MethodName)
			if !mfn.IsValid() {
				panic(fmt.Sprintf("method '%s' does not exist", m.MethodName))
			}
			return func(writer http.ResponseWriter, request *http.Request) {
				mfn.Call([]reflect.Value{reflect.ValueOf(writer), reflect.ValueOf(request)})
			}
		} else {
			panic(fmt.Sprintf("property .This not set (trying to obtain method '%s')", m.MethodName))
		}
	} else {
		panic("no Handler or Func set on method")
	}
}

var MethodsOrder = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodOptions,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

func (m Methods) writeYaml(parentTag string, w *yamlWriter) {
	type sortMethod struct {
		name   string
		method Method
		order  int
	}
	sortedMethods := make([]sortMethod, 0, len(m))
	for n, me := range m {
		order := slices.Index(MethodsOrder, n)
		if order == -1 {
			order = len(MethodsOrder)
		}
		sortedMethods = append(sortedMethods, sortMethod{
			name:   n,
			method: me,
			order:  order,
		})
	}
	sort.Slice(sortedMethods, func(i, j int) bool {
		return sortedMethods[i].order < sortedMethods[j].order
	})
	for _, mn := range sortedMethods {
		mn.method.writeYaml(parentTag, mn.name, w)
	}
}

func (m Method) writeYaml(parentTag string, method string, w *yamlWriter) {
	w.writeTagStart(strings.ToLower(method))
	w.writeTagValue("summary", m.Summary)
	w.writeTagValue("description", m.Description)
	w.writeTagValue("operationId", m.OperationId)
	if tag := defaultTag(parentTag, m.Tag); tag != "" {
		w.writeTagStart("tags")
		w.writeItem(tag)
		w.writeTagEnd()
	}
	// TODO more?
	w.writeTagEnd()
}
