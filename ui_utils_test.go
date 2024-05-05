package chioas

import (
	"github.com/stretchr/testify/assert"
	"html/template"
	"testing"
)

func TestFavIcons(t *testing.T) {
	o := FavIcons{
		64: "test.png",
	}
	html := o.toHtml()
	assert.Equal(t, template.HTML(`<link rel="icon" type="image/png" href="./test.png" sizes="64x64" />`), html)
}
