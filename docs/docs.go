// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// 2019-11-19 22:04:19.389931 +0800 CST m=+0.030611284

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/sd/cpu": {
            "get": {
                "description": "CPUCheck",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "sd"
                ],
                "summary": "CPUCheck checks the cpu usage.",
                "responses": {
                    "200": {
                        "description": "OK - Load average: xx, xx, xx | Cores: x",
                        "schema": {
                            "$ref": "#/definitions/user.CreateResponse"
                        }
                    }
                }
            }
        },
        "/sd/disk": {
            "get": {
                "description": "DiskCheck",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "sd"
                ],
                "summary": "DiskCheck checks the disk usage.",
                "responses": {
                    "200": {
                        "description": "OK - Free space: xxxMB (xxGB) / xxxMB (xxGB) | Used: xx%",
                        "schema": {
                            "$ref": "#/definitions/user.CreateResponse"
                        }
                    }
                }
            }
        },
        "/sd/health": {
            "get": {
                "description": "HealthCheck",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "sd"
                ],
                "summary": "HealthCheck shows ` + "`" + `OK` + "`" + ` as the ping-pong result.",
                "responses": {
                    "200": {
                        "description": "OK ",
                        "schema": {
                            "$ref": "#/definitions/user.CreateResponse"
                        }
                    }
                }
            }
        },
        "/sd/ram": {
            "get": {
                "description": "RAMCheck",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "sd"
                ],
                "summary": "RAMCheck checks the disk usage.",
                "responses": {
                    "200": {
                        "description": "OK - Free space: xxMB (xxGB) / xxMB (xxGB) | Used: xx%",
                        "schema": {
                            "$ref": "#/definitions/user.CreateResponse"
                        }
                    }
                }
            }
        },
        "/user": {
            "get": {
                "description": "List users",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "List users in the database",
                "parameters": [
                    {
                        "description": "List users",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.ListRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "{\"code\":0,\"message\":\"OK\",\"data\":{\"username\":\"admin\"}}",
                        "schema": {
                            "$ref": "#/definitions/user.ListResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Add a new user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Add new user to the database",
                "parameters": [
                    {
                        "description": "Create a new user",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.CreateRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "{\"code\":0,\"message\":\"OK\",\"data\":{\"username\":\"admin\"}}",
                        "schema": {
                            "$ref": "#/definitions/user.CreateResponse"
                        }
                    }
                }
            }
        },
        "/user/:id": {
            "put": {
                "description": "Update user info",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Update update a exist user account info.",
                "parameters": [
                    {
                        "description": "Update",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.UserModel"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "{\"code\":0,\"message\":\"OK\",\"data\":{\"username\":\"admin\"}}",
                        "schema": {
                            "$ref": "#/definitions/model.UserModel"
                        }
                    }
                }
            },
            "delete": {
                "description": "Del a user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Delete a user from the database",
                "parameters": [
                    {
                        "description": "Delete a user",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.CreateRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "{\"code\":0,\"message\":\"OK\",\"data\":{\"username\":\"admin\"}}",
                        "schema": {
                            "$ref": "#/definitions/user.CreateResponse"
                        }
                    }
                }
            }
        },
        "/user/:username": {
            "get": {
                "description": "Get a user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get a user from the database",
                "parameters": [
                    {
                        "description": "Delete a user",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/user.CreateRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "{\"code\":0,\"message\":\"OK\",\"data\":{\"username\":\"admin\"}}",
                        "schema": {
                            "$ref": "#/definitions/user.CreateResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.UserInfo": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "password": {
                    "type": "string"
                },
                "sayHello": {
                    "type": "string"
                },
                "updatedAt": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "model.UserModel": {
            "type": "object",
            "required": [
                "password",
                "username"
            ],
            "properties": {
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "user.CreateRequest": {
            "type": "object",
            "properties": {
                "password": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "user.CreateResponse": {
            "type": "object",
            "properties": {
                "username": {
                    "type": "string"
                }
            }
        },
        "user.ListRequest": {
            "type": "object",
            "properties": {
                "limit": {
                    "type": "integer"
                },
                "offset": {
                    "type": "integer"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "user.ListResponse": {
            "type": "object",
            "properties": {
                "totalCount": {
                    "type": "integer"
                },
                "userList": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.UserInfo"
                    }
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "",
	Host:        "",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "",
	Description: "",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
