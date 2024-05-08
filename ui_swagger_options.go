package chioas

import (
	"html/template"
	"strings"
)

const defaultSwaggerTemplate = `<!DOCTYPE html>
<html lang="en">
  <head>
    <title>{{.title}}</title>
    <meta charset="utf-8" />
    <link rel="stylesheet" type="text/css" href="./swagger-ui.css" />
    <link rel="stylesheet" type="text/css" href="./index.css" />
    {{.favIcons}}
    <style>{{.stylesOverride}}</style>
    {{.headScript}}
  </head>
  <body>
    {{.headerHtml}}
    <div id="swagger-ui"></div>
    <script src="./swagger-ui-bundle.js" charset="UTF-8"> </script>
    <script src="./swagger-ui-standalone-preset.js" charset="UTF-8"> </script>
    <script>
      window.onload = function() {
		let cfg = {{.swaggeropts}}
		{{.swaggerpresets}}
		{{.swaggerplugins}}
        const ui = SwaggerUIBundle(cfg)
        window.ui = ui
      }
    </script>
    {{.bodyScript}}
  </body>
</html>`

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
	FavIcons                 FavIcons        `json:"-"`
	HeaderHtml               template.HTML   `json:"-"`
	HeadScript               template.JS     `json:"-"`
	BodyScript               template.JS     `json:"-"`
}

var defaultSwaggerFavIcons = FavIcons{
	16: "favicon-16x16.png",
	32: "favicon-32x32.png",
}

func (o SwaggerOptions) GetFavIcons() FavIcons {
	if o.FavIcons != nil {
		return o.FavIcons
	}
	return defaultSwaggerFavIcons
}

func (o SwaggerOptions) ToMap() map[string]any {
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
	if o.HeaderHtml != "" {
		m[htmlTagHeaderHtml] = o.HeaderHtml
	}
	if o.HeadScript != "" {
		m[htmlTagHeadScript] = template.HTML(`<script>` + o.HeadScript + `</script>`)
	}
	if o.BodyScript != "" {
		m[htmlTagBodyScript] = template.HTML(`<script>` + o.BodyScript + `</script>`)
	}
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

// SwaggerPlugin is used by SwaggerOptions.Plugins
type SwaggerPlugin string

// SwaggerPreset is used by SwaggerOptions.Presets
type SwaggerPreset string
