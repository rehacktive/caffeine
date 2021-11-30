package service

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (s *Server) generateOpenAPIMap(namespaces []string) (map[string]interface{}, error) {
	pathsMap := map[string]interface{}{}
	schemasMap := map[string]interface{}{}

	rootMap := map[string]interface{}{
		"openapi": "3.0.1",
		"info": map[string]interface{}{
			"title":       "Caffeine Application Server",
			"description": "Sample Caffeine application",
			"version":     "0.1.0",
		},
		"externalDocs": map[string]interface{}{
			"description": "Github Project Page",
			"url":         "https://github.com/rehacktive/caffeine",
		},
		"servers": []interface{}{
			map[string]interface{}{
				"url": "http://localhost:8000",
			},
		},
		"paths": pathsMap,
	}

	for _, namespace := range namespaces {
		if strings.HasSuffix(namespace, SchemaId) {
			continue
		}

		path := fmt.Sprintf("/ns/%s/{id}", namespace)
		namespacePath := fmt.Sprintf("/ns/%s", namespace)
		searchPath := fmt.Sprintf("/search/%s", namespace)

		var hasSchema = false
		var schemaNode = map[string]interface{}{}
		var schemaRef = ""

		// if namespace has a schema, add it to the schemas map
		schemaJson, dbErr := s.db.Get(namespace+SchemaId, SchemaId)

		if dbErr != nil {
			//Ignore
		} else {
			parsedSchema := map[string]interface{}{}
			err := json.Unmarshal(schemaJson, &parsedSchema)

			if err != nil {
				return nil, err
			}

			schemasMap[namespace] = parsedSchema
			schemaRef = fmt.Sprintf("#/components/schemas/%v", namespace)

			schemaNode = map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": schemaRef,
				},
			}

			hasSchema = true
		}

		getOperationMap := map[string]interface{}{
			"description": fmt.Sprintf("Get %v by id.", namespace),
			"tags": []interface{}{
				namespace,
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":     "id",
					"in":       "path",
					"required": true,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"responses": map[string]interface{}{
				"default": map[string]interface{}{
					"description": "default response",
					"content":     map[string]interface{}{},
				},
				"200": map[string]interface{}{
					"description": "200 OK",
					"content": map[string]interface{}{
						"application/json": schemaNode,
					},
				},
				"404": map[string]interface{}{
					"description": "404 Not Found",
					"content":     map[string]interface{}{},
				},
			},
		}

		postOperationMap := map[string]interface{}{
			"description": fmt.Sprintf("Insert or update %v with the given id.", namespace),
			"tags": []interface{}{
				namespace,
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":     "id",
					"in":       "path",
					"required": true,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"requestBody": map[string]interface{}{
				"content": map[string]interface{}{
					"application/json": schemaNode,
				},
			},
			"responses": map[string]interface{}{
				"default": map[string]interface{}{
					"description": "default response",
					"content":     map[string]interface{}{},
				},
				"201": map[string]interface{}{
					"description": "201 Created",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{},
					},
				},
			},
		}

		deleteOperationMap := map[string]interface{}{
			"description": fmt.Sprintf("Delete %v with the given id.", namespace),
			"tags": []interface{}{
				namespace,
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"name":     "id",
					"in":       "path",
					"required": true,
					"schema": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"responses": map[string]interface{}{
				"default": map[string]interface{}{
					"description": "default response",
					"content":     map[string]interface{}{},
				},
				"202": map[string]interface{}{
					"description": "200 Accepted",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{},
					},
				},
				"404": map[string]interface{}{
					"description": "404 Not Found",
					"content":     map[string]interface{}{},
				},
			},
		}

		pathsMap[path] = map[string]interface{}{
			"get":    getOperationMap,
			"post":   postOperationMap,
			"delete": deleteOperationMap,
		}

		var getNamespaceSchemaNode = map[string]interface{}{}

		if hasSchema {
			getNamespaceSchema := map[string]interface{}{
				"type": "array",
				"title": fmt.Sprintf("Get All %v", namespace),
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"key": map[string]interface{}{
							"type": "string",
						},
						"value": map[string]interface{}{
							"$ref": schemaRef,
						},
					},
				},
			}


			componentKey := fmt.Sprintf("get-all-%v", namespace)
			schemasMap[componentKey] = getNamespaceSchema
			getNamespaceSchemaRef := fmt.Sprintf("#/components/schemas/%v", componentKey)

			getNamespaceSchemaNode = map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": getNamespaceSchemaRef,
				},
			}
		}

		getNamespaceOperationMap := map[string]interface{}{
			"description": fmt.Sprintf("Get all values for the namespace '%v'.", namespace),
			"tags": []interface{}{
				namespace,
			},
			"parameters": []interface{}{},
			"responses": map[string]interface{}{
				"default": map[string]interface{}{
					"description": "default response",
					"content":     map[string]interface{}{},
				},
				"200": map[string]interface{}{
					"description": "200 OK",
					"content": map[string]interface{}{
						"application/json": getNamespaceSchemaNode,
					},
				},
			},
		}

		deleteNamespaceOperationMap := map[string]interface{}{
			"description": fmt.Sprintf("Delete the namespace '%v'.", namespace),
			"tags": []interface{}{
				namespace,
			},
			"parameters": []interface{}{},
			"responses": map[string]interface{}{
				"default": map[string]interface{}{
					"description": "default response",
					"content":     map[string]interface{}{},
				},
				"200": map[string]interface{}{
					"description": "200 OK",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{},
					},
				},
			},
		}

		pathsMap[namespacePath] = map[string]interface{}{
			"get":    getNamespaceOperationMap,
			"delete": deleteNamespaceOperationMap,
		}

		var searchSchemaNode = map[string]interface{}{}

		if hasSchema {
			searchSchema := map[string]interface{} {
				"title": fmt.Sprintf("Search %v", namespace),
				"properties": map[string]interface{}{
					"results": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"key": map[string]interface{}{
									"type": "string",
								},
								"value": map[string]interface{}{
									"$ref": schemaRef,
								},
							},
						},
					},
				},
			}

			componentKey := fmt.Sprintf("search-%v", namespace)
			schemasMap[componentKey] = searchSchema
			searchSchemaRef := fmt.Sprintf("#/components/schemas/%v", componentKey)

			searchSchemaNode = map[string]interface{}{
				"schema": map[string]interface{}{
					"$ref": searchSchemaRef,
				},
			}
		}

		searchOperationMap := map[string]interface{}{
			"description": fmt.Sprintf("Search namespace '%v' by property (jq syntax).", namespace),
			"tags": []interface{}{
				namespace,
			},
			"parameters": []interface{}{
				map[string]interface{}{
					"in":   "query",
					"name": "filter",
					"schema": map[string]interface{}{
						"type":    "string",
						"example": `select(.firstName=="Jack")`,
					},
				},
			},
			"responses": map[string]interface{}{
				"default": map[string]interface{}{
					"description": "default response",
					"content":     map[string]interface{}{},
				},
				"200": map[string]interface{}{
					"description": "200 OK",
					"content": map[string]interface{}{
						"application/json": searchSchemaNode,
					},
				},
			},
		}

		pathsMap[searchPath] = map[string]interface{}{
			"get": searchOperationMap,
		}
	}

	getAllNamespacesOperationMap := map[string]interface{}{
		"description": fmt.Sprintf("List all namespaces"),
		"tags": []interface{}{
			"namespaces",
		},
		"parameters": []interface{}{},
		"responses": map[string]interface{}{
			"default": map[string]interface{}{
				"description": "default response",
				"content":     map[string]interface{}{},
			},
			"200": map[string]interface{}{
				"description": "200 OK",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{},
				},
			},
		},
	}

	pathsMap["/ns"] = map[string]interface{}{
		"get": getAllNamespacesOperationMap,
	}

	if len(schemasMap) != 0 {
		rootMap["components"] = map[string]interface{}{
			"schemas": schemasMap,
		}
	}

	return rootMap, nil
}
