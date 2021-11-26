package service

import "fmt"

func generateOpenAPIMap(namespaces []string) map[string]interface{} {
	pathsMap := map[string]interface{}{}

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
		path := fmt.Sprintf("/ns/%s/{id}", namespace)
		namespacePath := fmt.Sprintf("/ns/%s", namespace)
		searchPath := fmt.Sprintf("/search/%s", namespace)

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
						"application/json": map[string]interface{}{},
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
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{},
					},
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
						"application/json": map[string]interface{}{},
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
						"example": `select(.name=="jack")`,
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
						"application/json": map[string]interface{}{},
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
	return rootMap
}
