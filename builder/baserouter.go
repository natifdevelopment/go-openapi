package builder

import (
	"github.com/natifdevelopment/go-openapi/schemas"
)

// GenerateCRUDPaths generates the standard CRUD endpoints that the
// BaseRouter pattern creates for each module in bbo-api.
//
// Returns a map of path → PathItem (map[string]interface{}).
func GenerateCRUDPaths(prefix, tag, modelSchemaName, requestSchemaName string) map[string]interface{} {
	paths := make(map[string]interface{})
	idPath := prefix + "/{id}"

	bearerSecurity := []interface{}{
		map[string]interface{}{schemas.SecurityBearerAuth: []interface{}{}},
	}

	standardResponses := func() map[string]interface{} {
		r := map[string]interface{}{
			"200": schemas.StandardSuccessResponse(),
		}
		for code, resp := range schemas.StandardErrorResponsesMap() {
			r[code] = resp
		}
		return r
	}

	requestBody := func() map[string]interface{} {
		if requestSchemaName == "" {
			return nil
		}
		return map[string]interface{}{
			"required": true,
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": map[string]interface{}{"$ref": "#/components/schemas/" + requestSchemaName},
				},
			},
		}
	}

	idParam := func() []interface{} {
		return []interface{}{
			map[string]interface{}{
				"name":     "id",
				"in":       "path",
				"required": true,
				"schema":   map[string]interface{}{"type": "string"},
			},
		}
	}

	paginationParams := func() []interface{} {
		return []interface{}{
			map[string]interface{}{"name": "page", "in": "query", "required": false,
				"schema": map[string]interface{}{"type": "integer", "example": 1}},
			map[string]interface{}{"name": "page_size", "in": "query", "required": false,
				"schema": map[string]interface{}{"type": "integer", "example": 10}},
			map[string]interface{}{"name": "sort", "in": "query", "required": false,
				"schema": map[string]interface{}{"type": "string", "example": "-created_at"}},
		}
	}

	// GET {prefix} - List, POST {prefix} - Create
	listCreate := map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "List " + tag,
			"description": "Get paginated list of " + tag + " records.",
			"parameters":  paginationParams(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
		"post": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Create " + tag,
			"description": "Create a new " + tag + " record.",
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}
	if rb := requestBody(); rb != nil {
		listCreate["post"].(map[string]interface{})["requestBody"] = rb
	}
	paths[prefix] = listCreate

	// GET/POST {prefix}/init
	paths[prefix+"/init"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Get page info for " + tag,
			"description": "Get page initialization info (CSRF token, metadata).",
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
		"post": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Init CSRF for " + tag,
			"description": "Initialize CSRF token for create operation.",
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// GET {prefix}/{id} - Retrieve
	paths[idPath] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Get " + tag + " detail",
			"description": "Retrieve a single " + tag + " record by ID.",
			"parameters":  idParam(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// POST {prefix}/{id}/edit/init
	paths[idPath+"/edit/init"] = map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Init CSRF for edit " + tag,
			"description": "Initialize CSRF token for edit operation.",
			"parameters":  idParam(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// POST {prefix}/{id}/edit - Update
	editPath := map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Update " + tag,
			"description": "Update an existing " + tag + " record.",
			"parameters":  idParam(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}
	if rb := requestBody(); rb != nil {
		editPath["post"].(map[string]interface{})["requestBody"] = rb
	}
	paths[idPath+"/edit"] = editPath

	// POST {prefix}/{id}/soft/init
	paths[idPath+"/soft/init"] = map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Init CSRF for soft delete " + tag,
			"description": "Initialize CSRF token for soft delete operation.",
			"parameters":  idParam(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// DELETE {prefix}/{id}/soft
	paths[idPath+"/soft"] = map[string]interface{}{
		"delete": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Soft delete " + tag,
			"description": "Soft delete a " + tag + " record.",
			"parameters":  idParam(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// GET {prefix}/deleted
	paths[prefix+"/deleted"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "List deleted " + tag,
			"description": "Get paginated list of soft-deleted " + tag + " records.",
			"parameters":  paginationParams(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// POST {prefix}/{id}/restore/init
	paths[idPath+"/restore/init"] = map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Init CSRF for restore " + tag,
			"description": "Initialize CSRF token for restore operation.",
			"parameters":  idParam(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// PUT {prefix}/{id}/restore
	paths[idPath+"/restore"] = map[string]interface{}{
		"put": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Restore " + tag,
			"description": "Restore a soft-deleted " + tag + " record.",
			"parameters":  idParam(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// GET {prefix}/{id}/document
	paths[idPath+"/document"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Get document for " + tag,
			"description": "Get document associated with a " + tag + " record.",
			"parameters":  idParam(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// GET {prefix}/{id}/my-approval
	paths[idPath+"/my-approval"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Get my approval for " + tag,
			"description": "Get approval status for the current user.",
			"parameters":  idParam(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// POST {prefix}/{id}/approval
	paths[idPath+"/approval"] = map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Submit approval for " + tag,
			"description": "Submit an approval decision.",
			"parameters":  idParam(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// POST {prefix}/{id}/approve
	paths[idPath+"/approve"] = map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Approve " + tag,
			"description": "Approve a " + tag + " document.",
			"parameters":  idParam(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// POST {prefix}/{id}/reject
	paths[idPath+"/reject"] = map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Reject " + tag,
			"description": "Reject a " + tag + " document.",
			"parameters":  idParam(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// GET {prefix}/draft-count
	paths[prefix+"/draft-count"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Get draft count for " + tag,
			"description": "Get the count of draft " + tag + " records.",
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	// GET {prefix}/{id}/draft
	paths[idPath+"/draft"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []interface{}{tag},
			"summary":     "Get latest draft for " + tag,
			"description": "Get the latest draft of a " + tag + " record.",
			"parameters":  idParam(),
			"responses":   standardResponses(),
			"security":    bearerSecurity,
		},
	}

	return paths
}
