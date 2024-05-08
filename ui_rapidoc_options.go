package chioas

import (
	"fmt"
	"html/template"
	"strings"
)

const defaultRapidocTemplate = `<!doctype html>
<html lang="en">
    <head>
        <title>{{.title}}</title>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width,minimum-scale=1,initial-scale=1,user-scalable=yes">
        <link rel="stylesheet" href="default.min.css">
        <script src="highlight.min.js"></script>
        <script defer="defer" src="rapidoc-min.js"></script>
        {{.favIcons}}
        <style>{{.stylesOverride}}</style>
        {{.headScript}}
    </head>
    <body>
        <rapi-doc id="thedoc" spec-url="{{.specName}}" {{.add_atts}}>
          {{.logo}}
          {{.innerHtml}}
        </rapi-doc>
        {{.bodyScript}}
    </body>
</html>`

// RapidocOptions describes the rapidoc-ui options
//
// for documentation see https://github.com/rapi-doc/RapiDoc
type RapidocOptions struct {
	ShowHeader             bool   // show-header
	HeadingText            string // heading-text
	Theme                  string // theme="light"
	RenderStyle            string // render-style="view"
	SchemaStyle            string // schema-style="table"
	ShowMethodInNavBar     string // show-method-in-nav-bar="true"
	UsePathInNavBar        bool   // use-path-in-nav-bar="true"
	ShowComponents         bool   // show-components="true"
	HideInfo               bool   // !show-info="true"
	AllowSearch            bool   // allow-search="false"
	AllowAdvancedSearch    bool   // allow-advanced-search="false"
	AllowSpecUrlLoad       bool   // allow-spec-url-load="false"
	AllowSpecFileLoad      bool   // allow-spec-file-load="false"
	DisallowTry            bool   // !allow-try="true"
	DisallowSpecDownload   bool   // !allow-spec-file-download="true"
	AllowServerSelection   bool   // allow-server-selection="false"
	DisallowAuthentication bool   // !allow-authentication="true"
	NoPersistAuth          bool   // !persist-auth="true"
	UpdateRoute            bool   // update-route="true"
	MatchType              string // match-type="regex"
	MonoFont               string // mono-font
	RegularFont            string // regular-font
	HeaderColor            string // header-color
	PrimaryColor           string // primary-color
	SchemaExpandLevel      uint   // schema-expand-level
	// LogoSrc is the src for Rapidoc logo
	LogoSrc string
	// InnerHtml is any inner HTML for the <rapi-doc> element
	InnerHtml  template.HTML
	HeadScript template.JS
	BodyScript template.JS
	FavIcons   FavIcons
	// AdditionalAttributes is used to add any rapidoc attributes
	// not covered by other options
	//
	// Full reference at https://rapidocweb.com/api.html#att
	AdditionalAttributes map[string]string
}

func (o RapidocOptions) GetFavIcons() FavIcons {
	if o.FavIcons != nil {
		return o.FavIcons
	}
	return FavIcons{
		16: "favicon-16x16.png",
		32: "favicon-32x32.png",
	}
}

func (o RapidocOptions) ToMap() map[string]any {
	atts := make([]string, 0, 20+len(o.AdditionalAttributes))
	if !o.hasAdditional("show-header") {
		atts = append(atts, fmt.Sprintf(`show-header="%t"`, o.ShowHeader))
	}
	if o.HeadingText != "" && !o.hasAdditional("heading-text") {
		atts = append(atts, fmt.Sprintf(`heading-text="%s"`, safeAttValue(o.HeadingText)))
	}
	if o.Theme != "" && !o.hasAdditional("theme") {
		atts = append(atts, fmt.Sprintf(`theme="%s"`, safeAttValue(o.Theme)))
	}
	if o.RenderStyle != "" && !o.hasAdditional("render-style") {
		atts = append(atts, fmt.Sprintf(`render-style="%s"`, safeAttValue(o.RenderStyle)))
	}
	if o.SchemaStyle != "" && !o.hasAdditional("schema-style") {
		atts = append(atts, fmt.Sprintf(`schema-style="%s"`, safeAttValue(o.SchemaStyle)))
	}
	if o.ShowMethodInNavBar != "" && !o.hasAdditional("show-method-in-nav-bar") {
		atts = append(atts, fmt.Sprintf(`show-method-in-nav-bar="%s"`, safeAttValue(o.ShowMethodInNavBar)))
	}
	if o.UsePathInNavBar && !o.hasAdditional("use-path-in-nav-bar") {
		atts = append(atts, `use-path-in-nav-bar="true"`)
	}
	if o.ShowComponents && !o.hasAdditional("show-components") {
		atts = append(atts, `show-components="true"`)
	}
	if !o.hasAdditional("show-info") {
		atts = append(atts, fmt.Sprintf(`show-info="%t"`, !o.HideInfo))
	}
	if !o.hasAdditional("allow-search") {
		atts = append(atts, fmt.Sprintf(`allow-search="%t"`, o.AllowSearch))
	}
	if !o.hasAdditional("allow-advanced-search") {
		atts = append(atts, fmt.Sprintf(`allow-advanced-search="%t"`, o.AllowAdvancedSearch))
	}
	if !o.hasAdditional("allow-spec-url-load") {
		atts = append(atts, fmt.Sprintf(`allow-spec-url-load="%t"`, o.AllowSpecUrlLoad))
	}
	if !o.hasAdditional("allow-spec-file-load") {
		atts = append(atts, fmt.Sprintf(`allow-spec-file-load="%t"`, o.AllowSpecFileLoad))
	}
	if !o.hasAdditional("allow-try") {
		atts = append(atts, fmt.Sprintf(`allow-try="%t"`, !o.DisallowTry))
	}
	if !o.hasAdditional("allow-spec-file-download") {
		atts = append(atts, fmt.Sprintf(`allow-spec-file-download="%t"`, !o.DisallowSpecDownload))
	}
	if !o.hasAdditional("allow-server-selection") {
		atts = append(atts, fmt.Sprintf(`allow-server-selection="%t"`, o.AllowServerSelection))
	}
	if !o.hasAdditional("allow-authentication") {
		atts = append(atts, fmt.Sprintf(`allow-authentication="%t"`, !o.DisallowAuthentication))
	}
	if !o.hasAdditional("persist-auth") {
		atts = append(atts, fmt.Sprintf(`persist-auth="%t"`, !o.NoPersistAuth))
	}
	if !o.hasAdditional("update-route") {
		atts = append(atts, fmt.Sprintf(`update-route="%t"`, o.UpdateRoute))
	}
	if o.MatchType != "" && !o.hasAdditional("match-type") {
		atts = append(atts, fmt.Sprintf(`match-type="%s"`, safeAttValue(o.MatchType)))
	}
	if o.MonoFont != "" && !o.hasAdditional("mono-font") {
		atts = append(atts, fmt.Sprintf(`mono-font="%s"`, safeAttValue(o.MonoFont)))
	}
	if o.RegularFont != "" && !o.hasAdditional("regular-font") {
		atts = append(atts, fmt.Sprintf(`regular-font="%s"`, safeAttValue(o.RegularFont)))
	}
	if o.HeaderColor != "" && !o.hasAdditional("header-color") {
		atts = append(atts, fmt.Sprintf(`header-color="%s"`, safeAttValue(o.HeaderColor)))
	}
	if o.PrimaryColor != "" && !o.hasAdditional("primary-color") {
		atts = append(atts, fmt.Sprintf(`primary-color="%s"`, safeAttValue(o.PrimaryColor)))
	}
	if o.SchemaExpandLevel > 0 && !o.hasAdditional("schema-expand-level") {
		atts = append(atts, fmt.Sprintf(`schema-expand-level="%d"`, o.SchemaExpandLevel))
	}
	for k, v := range o.AdditionalAttributes {
		atts = append(atts, k+`="`+safeAttValue(v)+`"`)
	}
	result := map[string]any{
		"add_atts": template.HTMLAttr(strings.Join(atts, " ")),
	}
	if o.LogoSrc != "" {
		result["logo"] = template.HTML(fmt.Sprintf(`<img id="logo" slot="logo" src="%s">`, safeAttValue(o.LogoSrc)))
	}
	if o.InnerHtml != "" {
		result["innerHtml"] = o.InnerHtml
	}
	if o.HeadScript != "" {
		result[htmlTagHeadScript] = template.HTML(`<script>` + o.HeadScript + `</script>`)
	}
	if o.BodyScript != "" {
		result[htmlTagBodyScript] = template.HTML(`<script>` + o.BodyScript + `</script>`)
	}
	return result
}

func safeAttValue(s string) string {
	return strings.ReplaceAll(s, `"`, "&quot;")
}

func (o RapidocOptions) hasAdditional(n string) bool {
	_, has := o.AdditionalAttributes[n]
	return has
}
