package chioas

import (
	"encoding/json"
	"github.com/go-andiamo/chioas/yaml"
)

// Properties is an ordered collection of Property
type Properties []Property

// Property represents the OAS definition of a property
type Property struct {
	// Name is the OAS name of the property
	Name string
	// Description is the OAS description of the property
	Description string
	// Type is the OAS type of the property
	//
	// Should be one of "string", "object", "array", "boolean", "integer", "number" or "null"
	Type string
	// ItemType is the OAS type of array items
	//
	// only used if Type = "array"
	ItemType string
	// Properties is the ordered collection of sub-properties
	//
	// Only used if Type == "object" (or Type == "array" and ItemType == "object"
	Properties Properties
	// Required indicates the property is required
	//
	// see also Schema.RequiredProperties
	Required bool
	// Format is the OAS format for the property
	Format string
	// Example is the OAS example for the property
	Example any
	// Enum is the OAS enum of the property
	Enum []any
	// Deprecated is the OAS deprecated flag for the property
	Deprecated bool
	// Constraints is the OAS constraints for the property
	Constraints Constraints
	// SchemaRef is the OAS schema reference
	//
	// Only used if value is a non-empty string
	//
	// If the value does not contain a path (i.e. does not contain any "/") then the ref
	// path will be the value prefixed with components schemas path.  For example, specifying "foo"
	// will result in a schema ref:
	//   schema:
	//     $ref: "#/components/schemas/foo"
	SchemaRef string
	// Extensions is extension OAS yaml properties
	Extensions Extensions
	// Additional is any additional OAS spec yaml to be written
	Additional Additional
	// Comment is any comment(s) to appear in the OAS spec yaml
	Comment string
}

func (p Property) writeYaml(w yaml.Writer, top bool) {
	w.WriteTagStart("\"" + p.Name + "\"").
		WriteComments(p.Comment)
	if p.SchemaRef != "" {
		writeSchemaRef(p.SchemaRef, p.Type == tagValueTypeArray, w)
	} else {
		w.WriteTagValue(tagNameDescription, p.Description).
			WriteTagValue(tagNameType, defValue(p.Type, tagValueTypeString))
		if p.Type == tagValueTypeArray {
			w.WriteTagStart(tagNameItems).
				WriteTagValue(tagNameType, defValue(p.ItemType, tagValueTypeString))
		}
		w.WriteTagValue(tagNameExample, p.Example).
			WriteTagValue(tagNameFormat, nilString(p.Format))
		if len(p.Enum) > 0 {
			w.WriteTagStart(tagNameEnum)
			for _, e := range p.Enum {
				w.WriteItem(e)
			}
			w.WriteTagEnd()
		}
		w.WriteTagValue(tagNameRequired, nilBool(p.Required && !top)).
			WriteTagValue(tagNameDeprecated, nilBool(p.Deprecated))
		p.Constraints.writeYaml(w)
		if (p.Type == tagValueTypeObject || (p.Type == tagValueTypeArray && p.ItemType == tagValueTypeObject)) && len(p.Properties) > 0 {
			w.WriteTagStart(tagNameProperties)
			for _, sub := range p.Properties {
				sub.writeYaml(w, false)
			}
			w.WriteTagEnd()
		}
		if p.Type == tagValueTypeArray {
			w.WriteTagEnd()
		}
	}
	writeExtensions(p.Extensions, w)
	writeAdditional(p.Additional, p, w)
	w.WriteTagEnd()
}

// Constraints defines the constraints for an OAS property
type Constraints struct {
	Pattern          string
	Maximum          json.Number
	Minimum          json.Number
	ExclusiveMaximum bool
	ExclusiveMinimum bool
	Nullable         bool
	MultipleOf       uint
	MaxLength        uint
	MinLength        uint
	MaxItems         uint
	MinItems         uint
	UniqueItems      bool
	MaxProperties    uint
	MinProperties    uint
	// Additional is any other OAS constraints for a property (that are not currently defined in Constraints)
	Additional map[string]any
}

func (c Constraints) writeYaml(w yaml.Writer) {
	w.WriteTagValue(tagNamePattern, nilString(c.Pattern)).
		WriteTagValue(tagNameMaximum, nilNumber(c.Maximum)).
		WriteTagValue(tagNameMinimum, nilNumber(c.Minimum)).
		WriteTagValue(tagNameExclusiveMaximum, nilBool(c.ExclusiveMaximum)).
		WriteTagValue(tagNameExclusiveMinimum, nilBool(c.ExclusiveMinimum)).
		WriteTagValue(tagNameNullable, nilBool(c.Nullable)).
		WriteTagValue(tagNameMultipleOf, nilUint(c.MultipleOf)).
		WriteTagValue(tagNameMaxLength, nilUint(c.MaxLength)).
		WriteTagValue(tagNameMinLength, nilUint(c.MinLength)).
		WriteTagValue(tagNameMaxItems, nilUint(c.MaxItems)).
		WriteTagValue(tagNameMinItems, nilUint(c.MinItems)).
		WriteTagValue(tagNameUniqueItems, nilBool(c.UniqueItems)).
		WriteTagValue(tagNameMaxProperties, nilUint(c.MaxProperties)).
		WriteTagValue(tagNameMinProperties, nilUint(c.MinProperties))
	if c.Additional != nil {
		for k, v := range c.Additional {
			w.WriteTagValue(k, v)
		}
	}
}
