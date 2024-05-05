package chioas

import (
	"fmt"
	"html/template"
	"mime"
	"path/filepath"
	"reflect"
	"strings"
)

// MappableOptions is an interface that can be used by either DocOptions.RedocOptions or DocOptions.SwaggerOptions
// and converts the options to a map (as used by the HTML template)
type MappableOptions interface {
	ToMap() map[string]any
}

type OptionsWithFavIcons interface {
	GetFavIcons() FavIcons
}

// FavIcons is a map of icon size and fav icon filename
//
// Used by SwaggerOptions.FavIcons, RedocOptions.FavIcons and RapidocOptions.FavIcons
type FavIcons map[int]string

func (f FavIcons) toHtml() template.HTML {
	result := make([]string, 0, len(f))
	for k, v := range f {
		typeAtt := ""
		if mt := mime.TypeByExtension(filepath.Ext(v)); mt != "" {
			typeAtt = `type="` + mt + `"`
		}
		result = append(result, fmt.Sprintf(`<link rel="icon" %s href="./%s" sizes="%dx%d" />`, typeAtt, v, k, k))
	}
	return template.HTML(strings.Join(result, "\n"))
}

func addMap[T comparable](m map[string]any, pty T, name string) {
	var empty T
	if pty != empty {
		m[name] = pty
	}
}

func addMapDef[T comparable](m map[string]any, pty T, name string, def T) {
	if pty != def {
		m[name] = pty
	}
}

func addMapNonNil(m map[string]any, pty any, name string) {
	if !isNil(reflect.ValueOf(pty)) {
		m[name] = pty
	}
}

func isNil(v reflect.Value) bool {
	if v.IsValid() {
		switch v.Kind() {
		case reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
			return v.IsNil()
		default:
			return v.IsZero()
		}
	}
	return true
}

func addMapMappable(m map[string]any, pty MappableOptions, name string) {
	vo := reflect.ValueOf(pty)
	if !vo.IsNil() {
		m[name] = pty.ToMap()
	}
}
