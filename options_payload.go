package chioas

// OptionsMethodPayloadBuilder is an interface that can be provided to Definition.OptionsMethodPayloadBuilder and
// allows automatically created OPTIONS methods to return a body payload
type OptionsMethodPayloadBuilder interface {
	BuildPayload(path string, pathDef *Path, def *Definition) (data []byte, headers map[string]string)
}

// NewRootOptionsMethodPayloadBuilder provides an OptionsMethodPayloadBuilder that can be used for Definition.OptionsMethodPayloadBuilder
// and provides the OPTIONS payload body on API root as the OAS spec
func NewRootOptionsMethodPayloadBuilder() OptionsMethodPayloadBuilder {
	return &rootOptionsPayload{}
}

type rootOptionsPayload struct{}

func (r *rootOptionsPayload) BuildPayload(path string, pathDef *Path, def *Definition) (data []byte, headers map[string]string) {
	data = make([]byte, 0)
	headers = map[string]string{}
	if path == root {
		if def.DocOptions.AsJson {
			headers[hdrContentType] = contentTypeJson
		} else {
			headers[hdrContentType] = contentTypeYaml
		}
		if def.DocOptions.specData != nil {
			data = def.DocOptions.specData
		} else if def.DocOptions.AsJson {
			data, _ = def.AsJson()
		} else {
			data, _ = def.AsYaml()
		}
	}
	return
}
