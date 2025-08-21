package chioas

import (
	"github.com/stretchr/testify/assert"
	"html/template"
	"testing"
)

func TestRapidocOptions_ToMap(t *testing.T) {
	o := &RapidocOptions{}
	m := o.ToMap()
	assert.Equal(t, 1, len(m))
	atts := m["add_atts"].(template.HTMLAttr)
	const defaultAtts = `show-header="false" show-info="true" allow-search="false" allow-advanced-search="false" allow-spec-url-load="false" allow-spec-file-load="false" allow-try="true" allow-spec-file-download="true" allow-server-selection="false" allow-authentication="true" persist-auth="true" update-route="false"`
	assert.Equal(t, defaultAtts, string(atts))

	o = &RapidocOptions{
		HeadingText:        "foo1",
		Theme:              "foo2",
		RenderStyle:        "foo3",
		SchemaStyle:        "foo4",
		ShowMethodInNavBar: "true",
		UsePathInNavBar:    true,
		ShowComponents:     true,
		MonoFont:           "foo5",
		RegularFont:        "foo6",
		HeaderColor:        "foo7",
		PrimaryColor:       "foo8",
		SchemaExpandLevel:  1,
		MatchType:          "foo9",
		AdditionalAttributes: map[string]string{
			"text-color": "#000000",
			"bg-color":   "#ffffff",
		},
	}
	m = o.ToMap()
	assert.Equal(t, 1, len(m))
	atts = m["add_atts"].(template.HTMLAttr)
	assert.Contains(t, string(atts), `heading-text="foo1"`)
	assert.Contains(t, string(atts), `theme="foo2"`)
	assert.Contains(t, string(atts), `render-style="foo3"`)
	assert.Contains(t, string(atts), `schema-style="foo4"`)
	assert.Contains(t, string(atts), `show-method-in-nav-bar="true"`)
	assert.Contains(t, string(atts), `use-path-in-nav-bar="true"`)
	assert.Contains(t, string(atts), `show-components="true"`)
	assert.Contains(t, string(atts), `mono-font="foo5"`)
	assert.Contains(t, string(atts), `regular-font="foo6"`)
	assert.Contains(t, string(atts), `header-color="foo7"`)
	assert.Contains(t, string(atts), `primary-color="foo8"`)
	assert.Contains(t, string(atts), `match-type="foo9"`)
	assert.Contains(t, string(atts), `bg-color="#ffffff"`)
	assert.Contains(t, string(atts), `text-color="#000000"`)

	o = &RapidocOptions{
		LogoSrc:   "foo",
		LogoStyle: "width:20px",
	}
	m = o.ToMap()
	assert.Equal(t, 2, len(m))
	logoHtml := m["logo"].(template.HTML)
	assert.Equal(t, `<img id="logo" slot="logo" src="foo" style="width:20px">`, string(logoHtml))

	const testInner = `<div slot="nav-logo" style="display: flex; align-items: center; justify-content: center;"> 
    <img src="dog.png" style="width:40px; margin-right: 20px"> <span style="color:#fff"> <b>nav-logo</b> slot </span>
</div>`
	o = &RapidocOptions{
		InnerHtml:  testInner,
		HeadScript: `head();`,
		BodyScript: `body();`,
	}
	m = o.ToMap()
	assert.Equal(t, 4, len(m))
	inner := m["innerHtml"].(template.HTML)
	assert.Equal(t, testInner, string(inner))
	assert.Equal(t, template.HTML(`<script>head();</script>`), m[htmlTagHeadScript])
	assert.Equal(t, template.HTML(`<script>body();</script>`), m[htmlTagBodyScript])
}

func TestRapidocOptions_OverrideFavIcons(t *testing.T) {
	o := RapidocOptions{
		FavIcons: FavIcons{
			64: "test.png",
		},
	}
	icons := o.GetFavIcons()
	assert.NotNil(t, icons)
	html := icons.toHtml()
	assert.Equal(t, template.HTML(`<link rel="icon" type="image/png" href="./test.png" sizes="64x64" />`), html)
}
