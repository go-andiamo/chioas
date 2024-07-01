package chioas

import (
	"bytes"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromJson(t *testing.T) {
	r := bytes.NewReader([]byte(testPetStoreJson))
	api := &testApi{}
	handlers := Handlers{
		"getRoot": testGetRoot,
	}
	d, err := FromJson(r, &FromOptions{
		Api:      api,
		Handlers: handlers,
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(d.Paths))
	p, ok := d.Paths["/pets"]
	assert.True(t, ok)
	assert.Equal(t, 1, len(p.Paths))
	assert.Equal(t, 2, len(p.Methods))
	m, ok := p.Methods[http.MethodGet]
	assert.True(t, ok)
	assert.NotNil(t, m.Handler)

	assert.Equal(t, 1, len(d.Methods))
	m, ok = d.Methods[http.MethodGet]
	assert.True(t, ok)
	assert.NotNil(t, m.Handler)
	assert.Nil(t, d.DocOptions.specData)

	_, err = FromJson(bytes.NewReader([]byte(testPetStoreJson)), nil)
	assert.Error(t, err)
}

func TestFromJson_ServeDocs(t *testing.T) {
	r := bytes.NewReader([]byte(testPetStoreJson))
	api := &testApi{}
	handlers := Handlers{
		"getRoot": testGetRoot,
	}
	d, err := FromJson(r, &FromOptions{
		DocOptions: &DocOptions{ServeDocs: true},
		Api:        api,
		Handlers:   handlers,
	})
	assert.NoError(t, err)
	assert.NotNil(t, d.DocOptions.specData)
	assert.Equal(t, []byte(testPetStoreJson), d.DocOptions.specData)
	assert.True(t, d.DocOptions.AsJson)

	router := chi.NewRouter()
	err = d.SetupRoutes(router, nil)
	require.NoError(t, err)
	req, _ := http.NewRequest(http.MethodGet, defaultDocsPath+"/"+defaultSpecNameJson, nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, []byte(testPetStoreJson), body)

	req, _ = http.NewRequest(http.MethodGet, "/pets", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
}

func TestFromJson_Errors(t *testing.T) {
	testCases := []struct {
		json         string
		api          any
		handlers     Handlers
		strict       bool
		expectErrMsg string
	}{
		{
			json: `{
  "paths": {
    "/": {
      "get": {
        "x-handler": "getRoot"
      }
    }
  }
}`,
			expectErrMsg: "path '/', method 'GET' - 'x-handler' no method or func found for 'getRoot'",
		},
		{
			json: `{
  "paths": {
    "/": {
      "get": {
        "x-handler": ".GetRoot"
      }
    }
  }
}`,
			expectErrMsg: "path '/', method 'GET' - 'x-handler' no method or func found for '.GetRoot'",
		},
		{
			json: `{
  "paths": {
    "/": {
      "get": {
        "x-handler": ".GetRoot"
      }
    }
  }
}`,
			api:          &testApi{},
			expectErrMsg: "path '/', method 'GET' - 'x-handler' method 'GetRoot' does not exist",
		},
		{
			json: `{
  "paths": {
    "/": {
      "get": {
        "x-handler": ".GetRoot"
      }
    }
  }
}`,
			api:          &testBadApi{},
			expectErrMsg: "path '/', method 'GET' - 'x-handler' method 'GetRoot' is not http.HandlerFunc",
		},
		{
			json: `{
  "paths": {
    "/root": {
      "get": {
        "x-handler": ".GetRoot"
      }
    }
  }
}`,
			api:          &testBadApi{},
			expectErrMsg: "path '/root', method 'GET' - 'x-handler' method 'GetRoot' is not http.HandlerFunc",
		},
		{
			json: `{
  "paths": {
    "/": {
      "get": {
        "x-handler": false
      }
    }
  }
}`,
			expectErrMsg: "path '/', method 'GET' - 'x-handler' tag not a string",
		},
		{
			json: `{
  "paths": {
    "/": {
      "get": {
      }
    }
  }
}`,
			strict:       true,
			expectErrMsg: "path '/', method 'GET' - missing 'x-handler' tag",
		},
		{
			json: `{
  "paths": {
    "/": {
      "get": false
    }
  }
}`,
			expectErrMsg: "method 'GET' on path '/' not a map",
		},
		{
			json: `{
  "paths": {
    "/": false
  }
}`,
			expectErrMsg: "path '/' not a map",
		},
		{
			json: `{
  "paths": {
    "//": {
    }
  }
}`,
			expectErrMsg: "invalid path '//'",
		},
		{
			json: `{
  "paths": false
}`,
			expectErrMsg: "no paths defined",
		},
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("[%d]", i+1), func(t *testing.T) {
			r := bytes.NewReader([]byte(tc.json))
			_, err := FromJson(r, &FromOptions{
				Api:      tc.api,
				Handlers: tc.handlers,
				Strict:   tc.strict,
			})
			assert.Error(t, err)
			assert.Equal(t, tc.expectErrMsg, err.Error())
		})
	}
}

func TestFromYaml(t *testing.T) {
	r := bytes.NewReader([]byte(testPetstoreYaml))
	api := &testApi{}
	handlers := Handlers{
		"getRoot": testGetRoot,
	}
	mwPaths := make([]string, 0)
	d, err := FromYaml(r, &FromOptions{
		Api:      api,
		Handlers: handlers,
		PathMiddlewares: func(path string) chi.Middlewares {
			mwPaths = append(mwPaths, path)
			return nil
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(d.Paths))
	assert.Equal(t, []string{"/", "/pets", "/pets/{petId}"}, mwPaths)
	p, ok := d.Paths["/pets"]
	assert.True(t, ok)
	assert.Equal(t, 1, len(p.Paths))
	assert.Equal(t, 2, len(p.Methods))
	m, ok := p.Methods[http.MethodGet]
	assert.True(t, ok)
	assert.NotNil(t, m.Handler)

	assert.Equal(t, 1, len(d.Methods))
	m, ok = d.Methods[http.MethodGet]
	assert.True(t, ok)
	assert.NotNil(t, m.Handler)
	assert.Nil(t, d.DocOptions.specData)
}

func TestFromYaml_ServeDocs(t *testing.T) {
	r := bytes.NewReader([]byte(testPetstoreYaml))
	api := &testApi{}
	handlers := Handlers{
		"getRoot": testGetRoot,
	}
	d, err := FromYaml(r, &FromOptions{
		DocOptions: &DocOptions{ServeDocs: true},
		Api:        api,
		Handlers:   handlers,
	})
	assert.NoError(t, err)
	assert.NotNil(t, d.DocOptions.specData)
	assert.Equal(t, []byte(testPetstoreYaml), d.DocOptions.specData)
	assert.False(t, d.DocOptions.AsJson)

	router := chi.NewRouter()
	err = d.SetupRoutes(router, nil)
	require.NoError(t, err)
	req, _ := http.NewRequest(http.MethodGet, defaultDocsPath+"/"+defaultSpecName, nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, []byte(testPetstoreYaml), body)

	req, _ = http.NewRequest(http.MethodGet, "/pets", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	req, _ = http.NewRequest(http.MethodGet, "/", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Result().StatusCode)
}

func TestFromYaml_BadYaml(t *testing.T) {
	data := `null`
	r := bytes.NewReader([]byte(data))
	api := &testApi{}
	handlers := Handlers{
		"getRoot": testGetRoot,
	}
	_, err := FromYaml(r, &FromOptions{
		Api:      api,
		Handlers: handlers,
	})
	assert.Error(t, err)
	assert.Equal(t, "bad yaml", err.Error())
}

func testGetRoot(w http.ResponseWriter, r *http.Request) {
}

type testApi struct {
}

func (a *testApi) GetPets(w http.ResponseWriter, r *http.Request) {
}

func (a *testApi) PostPets(w http.ResponseWriter, r *http.Request) {
}

func (a *testApi) GetPet(w http.ResponseWriter, r *http.Request) {
}

func (a *testApi) PutPet(w http.ResponseWriter, r *http.Request) {
}

type testBadApi struct {
}

func (a *testBadApi) GetRoot() {
}

const testPetStoreJson = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Swagger Petstore - OpenAPI 3.0",
    "version": "1.0.17"
  },
  "tags": [
    {
      "name": "pet",
      "description": "Everything about your Pets",
      "externalDocs": {
        "description": "Find out more",
        "url": "http://swagger.io"
      }
    }
  ],
  "paths": {
    "/": {
      "get": {
        "x-handler": "getRoot",
        "responses": {
          "200": {
            "description": "Root discovery",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object"
                }
              }
            }
          }
        }
      }
    },
    "/pets": {
      "get": {
        "summary": "List/query pets",
        "description": "List/query pets",
        "operationId": "getPets",
        "x-handler": ".GetPets",
        "tags": [
          "pet"
        ],
        "parameters": [
          {
            "name": "status",
            "description": "Status values that need to be considered for filter",
            "in": "query",
            "required": false,
            "schema": {
              "$ref": "#/components/schemas/Status"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Resultant pets",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/Pet"
                  }
                }
              }
            }
          }
        }
      },
      "post": {
        "summary": "Add pet",
        "description": "Add pet",
        "operationId": "addPet",
        "x-handler": ".PostPets",
        "tags": [
          "pet"
        ],
        "requestBody": {
          "description": "Pet to be added to the store",
          "required": false,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/Pet"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Created pet",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/Pet"
                  }
                }
              }
            }
          }
        }
      }
    },
    "/pets/{petId}": {
      "get": {
        "summary": "Get an existing pet",
        "description": "Get an existing pet by Id",
        "operationId": "getPet",
        "x-handler": ".GetPet",
        "tags": [
          "pet"
        ],
        "parameters": [
          {
            "name": "petId",
            "description": "id of the Pet",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Successful operation",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Pet"
                }
              }
            }
          }
        }
      },
      "put": {
        "summary": "Update an existing pet",
        "description": "Update an existing pet by Id",
        "operationId": "updatePet",
        "x-handler": ".PutPet",
        "tags": [
          "pet"
        ],
        "parameters": [
          {
            "name": "petId",
            "description": "id of the Pet",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string"
            }
          }
        ],
        "requestBody": {
          "description": "Update an existent pet in the store",
          "required": false,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/Pet"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Successful operation",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Pet"
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Pet": {
        "type": "object",
        "required": [
          "name",
          "photoUrls"
        ],
        "properties": {
          "id": {
            "type": "integer",
            "example": 10
          },
          "name": {
            "type": "string",
            "example": "doggie"
          },
          "category": {
            "$ref": "#/components/schemas/Category"
          },
          "photoUrls": {
            "type": "array",
            "items": {
              "type": "string",
              "example": "https://example.com/dog-picture.jpg"
            }
          }
        }
      },
      "Category": {
        "type": "object",
        "properties": {
          "id": {
            "type": "integer",
            "example": 1
          },
          "name": {
            "type": "string",
            "example": "Dogs"
          }
        }
      },
      "Status": {
        "type": "string",
        "default": "available",
        "enum": [
          "available",
          "pending",
          "sold"
        ]
      }
    }
  }
}`

const testPetstoreYaml = `openapi: "3.0.3"
info:
  title: "Swagger Petstore - OpenAPI 3.0"
  version: "1.0.17"
tags:
  - name: "pet"
    description: "Everything about your Pets"
    externalDocs:
      description: "Find out more"
      url: "http://swagger.io"
paths:
  "/":
    get:
      x-handler: "getRoot"
      responses:
        200:
          description: Root discovery
          content:
            application/json:
              schema:
                type: object
  "/pets":
    get:
      summary: "List/query pets"
      description: "List/query pets"
      operationId: "getPets"
      x-handler: ".GetPets"
      tags:
        - "pet"
      parameters:
        - name: "status"
          description: "Status values that need to be considered for filter"
          in: "query"
          required: false
          schema:
            $ref: "#/components/schemas/Status"
      responses:
        200:
          description: "Resultant pets"
          content:
            application/json:
              schema:
                type: "array"
                items:
                  $ref: "#/components/schemas/Pet"
    post:
      summary: "Add pet"
      description: "Add pet"
      operationId: "addPet"
      x-handler: ".PostPets"
      tags:
        - "pet"
      requestBody:
        description: "Pet to be added to the store"
        required: false
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Pet"
      responses:
        200:
          description: "Created pet"
          content:
            application/json:
              schema:
                type: "array"
                items:
                  $ref: "#/components/schemas/Pet"
  "/pets/{petId}":
    get:
      summary: "Get an existing pet"
      description: "Get an existing pet by Id"
      operationId: "getPet"
      x-handler: ".GetPet"
      tags:
        - "pet"
      parameters:
        - name: "petId"
          description: "id of the Pet"
          in: "path"
          required: true
          schema:
            type: "string"
      responses:
        200:
          description: "Successful operation"
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pet"
    put:
      summary: "Update an existing pet"
      description: "Update an existing pet by Id"
      operationId: "updatePet"
      x-handler: ".PutPet"
      tags:
        - "pet"
      parameters:
        - name: "petId"
          description: "id of the Pet"
          in: "path"
          required: true
          schema:
            type: "string"
      requestBody:
        description: "Update an existent pet in the store"
        required: false
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/Pet"
      responses:
        200:
          description: "Successful operation"
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pet"
components:
  schemas:
    "Pet":
      type: "object"
      required:
        - "name"
        - "photoUrls"
      properties:
        "id":
          type: "integer"
          example: 10
        "name":
          type: "string"
          example: "doggie"
        "category":
          $ref: "#/components/schemas/Category"
        "photoUrls":
          type: "array"
          items:
            type: "string"
            example: "https://example.com/dog-picture.jpg"
    "Category":
      type: "object"
      properties:
        "id":
          type: "integer"
          example: 1
        "name":
          type: "string"
          example: "Dogs"
    "Status":
      type: "string"
      default: "available"
      enum:
        - "available"
        - "pending"
        - "sold"
`
