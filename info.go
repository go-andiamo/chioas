package chioas

import (
	"github.com/go-andiamo/chioas/internal/tags"
	"github.com/go-andiamo/chioas/yaml"
)

// Info represents the OAS info section of the spec
type Info struct {
	// Title is the OAS title
	Title string
	// Description is the OAS description
	Description string
	// Version is the OAS version (of the api)
	Version string
	// TermsOfService is the OAS terms of service
	TermsOfService string
	// Contact is the optional OAS contact info
	Contact *Contact
	// License is the optional OAS license info
	License *License
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
	// ExternalDocs is the optional eternal docs (for the entire spec)
	ExternalDocs *ExternalDocs
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
}

func (i Info) writeYaml(w yaml.Writer) {
	w.WriteTagStart(tags.Info).
		WriteComments(i.Comment).
		WriteTagValue(tags.Title, defValue(i.Title, defaultTitle)).
		WriteTagValue(tags.Description, i.Description).
		WriteTagValue(tags.Version, defValue(i.Version, "1.0.0")).
		WriteTagValue(tags.TermsOfService, i.TermsOfService)
	if i.Contact != nil {
		i.Contact.writeYaml(w)
	}
	if i.License != nil {
		i.License.writeYaml(w)
	}
	writeExtensions(i.Extensions, w)
	writeAdditional(i.Additional, i, w)
	w.WriteTagEnd()
	if i.ExternalDocs != nil {
		i.ExternalDocs.writeYaml(w)
	}
}

// Contact represents the OAS contact section of the info
type Contact struct {
	Name  string
	Url   string
	Email string
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
}

func (c *Contact) writeYaml(w yaml.Writer) {
	if c.Name != "" || c.Url != "" || c.Email != "" {
		w.WriteTagStart(tags.Contact).
			WriteComments(c.Comment).
			WriteTagValue(tags.Name, c.Name).
			WriteTagValue(tags.Url, c.Url).
			WriteTagValue(tags.Email, c.Email)
		writeExtensions(c.Extensions, w)
		w.WriteTagEnd()
	}
}

// License represents the OAS license section of the info
type License struct {
	Name string
	Url  string
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
}

func (l *License) writeYaml(w yaml.Writer) {
	if l.Name != "" {
		w.WriteTagStart(tags.License).
			WriteComments(l.Comment).
			WriteTagValue(tags.Name, l.Name).
			WriteTagValue(tags.Url, l.Url)
		writeExtensions(l.Extensions, w)
		w.WriteTagEnd()
	}
}

// ExternalDocs represents the OAS external docs for a spec
type ExternalDocs struct {
	Description string
	Url         string
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
}

func (x *ExternalDocs) writeYaml(w yaml.Writer) {
	if x.Url != "" {
		w.WriteTagStart(tags.ExternalDocs).
			WriteComments(x.Comment).
			WriteTagValue(tags.Description, x.Description).
			WriteTagValue(tags.Url, x.Url)
		writeExtensions(x.Extensions, w)
		w.WriteTagEnd()
	}
}
