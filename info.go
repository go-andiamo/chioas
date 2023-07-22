package chioas

import "github.com/go-andiamo/chioas/yaml"

type Info struct {
	Title          string
	Description    string
	Version        string
	TermsOfService string
	Contact        *Contact
	License        *License
	Additional     Additional
	ExternalDocs   *ExternalDocs
}

func (i Info) writeYaml(w yaml.Writer) {
	w.WriteTagStart(tagNameInfo).
		WriteTagValue(tagNameTitle, defValue(i.Title, defaultTitle)).
		WriteTagValue(tagNameDescription, i.Description).
		WriteTagValue(tagNameVersion, defValue(i.Version, "1.0.0")).
		WriteTagValue(tagNameTermsOfService, i.TermsOfService)
	if i.Contact != nil {
		i.Contact.writeYaml(w)
	}
	if i.License != nil {
		i.License.writeYaml(w)
	}
	if i.Additional != nil {
		i.Additional.Write(i, w)
	}
	w.WriteTagEnd()
	if i.ExternalDocs != nil {
		i.ExternalDocs.writeYaml(w)
	}
}

type Contact struct {
	Name  string
	Url   string
	Email string
}

func (c *Contact) writeYaml(w yaml.Writer) {
	if c.Name != "" || c.Url != "" || c.Email != "" {
		w.WriteTagStart(tagNameContact).
			WriteTagValue(tagNameName, c.Name).
			WriteTagValue(tagNameUrl, c.Url).
			WriteTagValue(tagNameEmail, c.Email).
			WriteTagEnd()
	}
}

type License struct {
	Name string
	Url  string
}

func (l *License) writeYaml(w yaml.Writer) {
	if l.Name != "" {
		w.WriteTagStart(tagNameLicense).
			WriteTagValue(tagNameName, l.Name).
			WriteTagValue(tagNameUrl, l.Url).
			WriteTagEnd()
	}
}

type ExternalDocs struct {
	Description string
	Url         string
}

func (x *ExternalDocs) writeYaml(w yaml.Writer) {
	if x.Url != "" {
		w.WriteTagStart(tagNameExternalDocs).
			WriteTagValue(tagNameDescription, x.Description).
			WriteTagValue(tagNameUrl, x.Url).
			WriteTagEnd()
	}
}
