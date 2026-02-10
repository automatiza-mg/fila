package llm

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
)

// GenerateSchema gera um schema json para o tipo genérico T.
func GenerateSchema[T any]() *jsonschema.Schema {
	reflector := jsonschema.Reflector{
		DoNotReference:            true,
		AllowAdditionalProperties: false,
	}

	var v T
	return reflector.Reflect(v)
}

// GenerateMapSchema gera um schema json usando [GenerateSchema] e faz o marhsal
// json para um mapa.
func GenerateMapSchema[T any]() (map[string]any, error) {
	schema := GenerateSchema[T]()

	data, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	var schemaMap map[string]any
	err = json.Unmarshal(data, &schemaMap)
	if err != nil {
		return nil, err
	}
	return schemaMap, nil
}
