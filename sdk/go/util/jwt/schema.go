package jwt

import (
	"github.com/xeipuuv/gojsonschema"

	"github.com/virtengine/virtengine/sdk/specs"
)

var (
	schemaLoader *gojsonschema.Schema
)

func init() {
	content, err := specs.GetSpecFile("jwt-schema.json")
	if err != nil {
		panic(err)
	}

	sl := gojsonschema.NewSchemaLoader()

	loader := gojsonschema.NewStringLoader(string(content))
	schemaLoader, err = sl.Compile(loader)
	if err != nil {
		panic(err)
	}
}
