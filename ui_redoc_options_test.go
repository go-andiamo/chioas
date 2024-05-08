package chioas

import (
	"github.com/stretchr/testify/assert"
	"html/template"
	"testing"
)

func TestRedocOptions_ToMap_Empty(t *testing.T) {
	o := &RedocOptions{}
	m := o.ToMap()
	assert.Empty(t, m)
}

func TestRedocOptions_ToMap_Theme(t *testing.T) {
	o := &RedocOptions{
		Theme: &RedocTheme{
			Spacing:     &RedocSpacing{},
			Breakpoints: &RedocBreakpoints{},
			Colors:      &RedocColors{},
			Typography: &RedocTypography{
				OptimizeSpeed: false,
				Headings:      &RedocHeadings{},
				Code:          &RedocCode{},
				Links:         &RedocLinks{},
			},
			Sidebar: &RedocSidebar{
				GroupItems:  &RedocGroupItems{},
				Level1Items: &RedocLevel1Items{},
				Arrow:       &RedocArrow{},
			},
			Logo: &RedocLogo{},
			RightPanel: &RedocRightPanel{
				Servers: &RedocRightPanelServers{
					Overlay: &RedocRightPanelServersOverlay{},
					Url:     &RedocRightPanelServersUrl{},
				},
			},
			Fab: &RedocFab{},
		},
	}
	m := o.ToMap()
	assert.Equal(t, 1, len(m))
	assert.Contains(t, m, "theme")
	m = m["theme"].(map[string]any)
	assert.Equal(t, 8, len(m))
	assert.Contains(t, m, "spacing")
	assert.Contains(t, m, "breakpoints")
	assert.Contains(t, m, "colors")
	assert.Contains(t, m, "typography")
	assert.Contains(t, m, "sidebar")
	assert.Contains(t, m, "logo")
	assert.Contains(t, m, "rightPanel")
	assert.Contains(t, m, "fab")
}

func TestRedocOptions_ToMap_Theme_Empty(t *testing.T) {
	o := &RedocOptions{
		Theme: &RedocTheme{},
	}
	m := o.ToMap()
	assert.Equal(t, 1, len(m))
	assert.Contains(t, m, "theme")
	m = m["theme"].(map[string]any)
	assert.Empty(t, m)
}

func TestRedocOptions_ToMap_HeaderAndScripts(t *testing.T) {
	o := &RedocOptions{
		HeaderHtml: `<div>HEADER</div>`,
		HeadScript: `head();`,
		BodyScript: `body();`,
	}
	m := o.ToMap()
	assert.Equal(t, template.HTML(`<div>HEADER</div>`), m[htmlTagHeaderHtml])
	assert.Equal(t, template.HTML(`<script>head();</script>`), m[htmlTagHeadScript])
	assert.Equal(t, template.HTML(`<script>body();</script>`), m[htmlTagBodyScript])
}
