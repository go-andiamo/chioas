package chioas

import (
	"html/template"
	"reflect"
	"strings"
)

// MappableOptions is an interface that can be used by either DocOptions.RedocOptions or DocOptions.SwaggerOptions
// and converts the options to a map (as used by the HTML template)
type MappableOptions interface {
	ToMap() map[string]any
}

// SwaggerOptions describes the swagger-ui options (as used by DocOptions.SwaggerOptions)
//
// for documentation see https://github.com/swagger-api/swagger-ui/blob/master/docs/usage/configuration.md
type SwaggerOptions struct {
	DomId                    string          `json:"dom_id,omitempty"`
	Layout                   string          `json:"layout,omitempty"`
	DeepLinking              bool            `json:"deepLinking"`
	DisplayOperationId       bool            `json:"displayOperationId"`
	DefaultModelsExpandDepth uint            `json:"defaultModelsExpandDepth,omitempty"`
	DefaultModelExpandDepth  uint            `json:"defaultModelExpandDepth,omitempty"`
	DefaultModelRendering    string          `json:"defaultModelRendering,omitempty"`
	DisplayRequestDuration   bool            `json:"displayRequestDuration"`
	DocExpansion             string          `json:"docExpansion,omitempty"`
	MaxDisplayedTags         uint            `json:"maxDisplayedTags,omitempty"`
	ShowExtensions           bool            `json:"showExtensions"`
	ShowCommonExtensions     bool            `json:"showCommonExtensions"`
	UseUnsafeMarkdown        bool            `json:"useUnsafeMarkdown"`
	SyntaxHighlightTheme     string          `json:"syntaxHighlight.theme,omitempty"`
	TryItOutEnabled          bool            `json:"tryItOutEnabled"`
	RequestSnippetsEnabled   bool            `json:"requestSnippetsEnabled"`
	SupportedSubmitMethods   []string        `json:"supportedSubmitMethods,omitempty"`
	ValidatorUrl             string          `json:"validatorUrl,omitempty"`
	WithCredentials          bool            `json:"withCredentials"`
	PersistAuthorization     bool            `json:"persistAuthorization"`
	Plugins                  []SwaggerPlugin `json:"-"`
	Presets                  []SwaggerPreset `json:"-"`
}

func (o *SwaggerOptions) ToMap() map[string]any {
	m := map[string]any{}
	addMap(m, o.DomId, "dom_id")
	addMap(m, o.Layout, "layout")
	addMap(m, o.DeepLinking, "deepLinking")
	addMap(m, o.DisplayOperationId, "displayOperationId")
	addMap(m, o.DefaultModelsExpandDepth, "defaultModelsExpandDepth")
	addMap(m, o.DefaultModelExpandDepth, "defaultModelExpandDepth")
	addMap(m, o.DefaultModelRendering, "defaultModelRendering")
	addMap(m, o.DisplayRequestDuration, "displayRequestDuration")
	addMap(m, o.DocExpansion, "docExpansion")
	addMap(m, o.MaxDisplayedTags, "maxDisplayedTags")
	addMap(m, o.ShowExtensions, "showExtensions")
	addMap(m, o.ShowCommonExtensions, "showCommonExtensions")
	addMap(m, o.UseUnsafeMarkdown, "useUnsafeMarkdown")
	addMap(m, o.SyntaxHighlightTheme, "syntaxHighlight.theme")
	addMap(m, o.TryItOutEnabled, "tryItOutEnabled")
	addMap(m, o.RequestSnippetsEnabled, "requestSnippetsEnabled")
	addMapNonNil(m, o.SupportedSubmitMethods, "supportedSubmitMethods")
	addMap(m, o.ValidatorUrl, "validatorUrl")
	addMap(m, o.WithCredentials, "withCredentials")
	addMap(m, o.PersistAuthorization, "persistAuthorization")
	return m
}

func (o *SwaggerOptions) jsPlugins() (result template.JS) {
	if l := len(o.Plugins); l > 0 {
		var pb strings.Builder
		pb.Grow(40 * l)
		pb.WriteString("cfg.plugins = [")
		for i, p := range o.Plugins {
			if i > 0 {
				pb.WriteByte(',')
			}
			pb.WriteString(string(p))
		}
		pb.WriteByte(']')
		result = template.JS(pb.String())
	}
	return
}

func (o *SwaggerOptions) jsPresets() (result template.JS) {
	if l := len(o.Presets); l > 0 {
		var pb strings.Builder
		pb.Grow(40 * l)
		pb.WriteString("cfg.presets = [")
		for i, p := range o.Presets {
			if i > 0 {
				pb.WriteByte(',')
			}
			pb.WriteString(string(p))
		}
		pb.WriteByte(']')
		result = template.JS(pb.String())
	} else {
		result = "cfg.presets = [SwaggerUIBundle.presets.apis,SwaggerUIStandalonePreset]"
	}
	return
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

// SwaggerPlugin is used by SwaggerOptions.Plugins
type SwaggerPlugin string

// SwaggerPreset is used by SwaggerOptions.Presets
type SwaggerPreset string
