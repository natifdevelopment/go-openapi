package builder

import (
	"reflect"
	"strings"
	"time"
)

// StructToSchemaMap converts a Go struct type to an OpenAPI 3.0 schema
// represented as map[string]interface{}.
func StructToSchemaMap(t reflect.Type) map[string]interface{} {
	if t == nil {
		return nil
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return map[string]interface{}{"type": "string", "format": "date-time"}
		}
		return structFieldsToSchema(t)
	case reflect.Slice, reflect.Array:
		return map[string]interface{}{
			"type":  "array",
			"items": typeToSchemaMap(t.Elem()),
		}
	case reflect.Map:
		return map[string]interface{}{
			"type":                 "object",
			"additionalProperties": typeToSchemaMap(t.Elem()),
		}
	default:
		return typeToSchemaMap(t)
	}
}

// typeToSchemaMap converts a Go type to an OpenAPI schema map.
func typeToSchemaMap(t reflect.Type) map[string]interface{} {
	if t == nil {
		return map[string]interface{}{"type": "string"}
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return map[string]interface{}{"type": "string", "format": "date-time"}
		}
		// For nested structs, use $ref by name
		name := t.Name()
		if name != "" {
			return map[string]interface{}{"$ref": "#/components/schemas/" + name}
		}
		return structFieldsToSchema(t)
	case reflect.Slice, reflect.Array:
		return map[string]interface{}{
			"type":  "array",
			"items": typeToSchemaMap(t.Elem()),
		}
	case reflect.Map:
		return map[string]interface{}{
			"type":                 "object",
			"additionalProperties": typeToSchemaMap(t.Elem()),
		}
	case reflect.String:
		return map[string]interface{}{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return map[string]interface{}{"type": "integer", "format": "int32"}
	case reflect.Int64:
		return map[string]interface{}{"type": "integer", "format": "int64"}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return map[string]interface{}{"type": "integer", "format": "int32"}
	case reflect.Uint64:
		return map[string]interface{}{"type": "integer", "format": "int64"}
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{"type": "number", "format": "float"}
	case reflect.Bool:
		return map[string]interface{}{"type": "boolean"}
	case reflect.Interface:
		return map[string]interface{}{"type": "object"}
	default:
		return map[string]interface{}{"type": "string"}
	}
}

// structFieldsToSchema converts a struct's fields to an OpenAPI object schema.
func structFieldsToSchema(t reflect.Type) map[string]interface{} {
	props := make(map[string]interface{})
	var required []interface{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.PkgPath != "" {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		fieldName := field.Name
		omitempty := false
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
			for _, opt := range parts[1:] {
				if opt == "omitempty" {
					omitempty = true
				}
			}
		}

		// Embedded struct: merge properties
		if jsonTag == "" && field.Anonymous {
			embeddedSchema := StructToSchemaMap(field.Type)
			if embeddedProps, ok := embeddedSchema["properties"].(map[string]interface{}); ok {
				for propName, propSchema := range embeddedProps {
					props[propName] = propSchema
				}
			}
			continue
		}

		isRequired := !omitempty && field.Type.Kind() != reflect.Ptr
		if isRequired {
			required = append(required, fieldName)
		}

		fieldSchema := typeToSchemaMap(field.Type)

		if example := field.Tag.Get("example"); example != "" {
			fieldSchema["example"] = example
		}
		if desc := field.Tag.Get("description"); desc != "" {
			fieldSchema["description"] = desc
		}

		props[fieldName] = fieldSchema
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

// StructToSchema is kept for backward compatibility but now delegates to StructToSchemaMap.
// Deprecated: use StructToSchemaMap instead.
func StructToSchema(t reflect.Type) map[string]interface{} {
	return StructToSchemaMap(t)
}

// SchemaNameFromType returns the schema name for a Go type.
func SchemaNameFromType(t reflect.Type) string {
	if t == nil {
		return ""
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Struct {
		return t.Name()
	}
	return ""
}
