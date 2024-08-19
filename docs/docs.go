// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Gabriele Genovese",
            "email": "gabriele.genovese2@studio.unibo.it"
        },
        "license": {
            "name": "AGPL-3.0",
            "url": "https://www.gnu.org/licenses/agpl-3.0.en.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/answer/{id}/vote": {
            "post": {
                "description": "Insert a new vote on a answer",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vote"
                ],
                "summary": "Insert a vote",
                "parameters": [
                    {
                        "type": "string",
                        "description": "code query parameter",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.Vote"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/util.ApiError"
                        }
                    }
                }
            }
        },
        "/answers": {
            "put": {
                "description": "Insert a new answer under a question",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "answer"
                ],
                "summary": "Insert a new answer",
                "parameters": [
                    {
                        "description": "Answer data to insert",
                        "name": "answerReq",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.PutAnswerRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.Answer"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/util.ApiError"
                        }
                    }
                }
            }
        },
        "/answers/{id}": {
            "delete": {
                "description": "Given an andwer ID, delete the answer",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "answer"
                ],
                "summary": "Delete an answer",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Answer id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.Answer"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/util.ApiError"
                        }
                    }
                }
            }
        },
        "/documents": {
            "put": {
                "description": "Insert a new document with all the questions initialised",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "document"
                ],
                "summary": "Insert a new document",
                "parameters": [
                    {
                        "description": "Doc request body",
                        "name": "docRequest",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.PutDocumentRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.Document"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/util.ApiError"
                        }
                    }
                }
            }
        },
        "/documents/{id}": {
            "get": {
                "description": "Given a document's ID, return all the questions",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "document"
                ],
                "summary": "Get a document's divisions",
                "parameters": [
                    {
                        "type": "string",
                        "description": "document id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.Document"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/util.ApiError"
                        }
                    }
                }
            }
        },
        "/questions/{id}": {
            "get": {
                "description": "Given a question ID, return the question and all its answers",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "question"
                ],
                "summary": "Get all answers given a question",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Answer id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/api.Answer"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/util.ApiError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "api.Answer": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "downvotes": {
                    "type": "integer"
                },
                "id": {
                    "description": "taken from from gorm.Model, so we can json strigify properly",
                    "type": "integer"
                },
                "parent": {
                    "type": "integer"
                },
                "question": {
                    "type": "integer"
                },
                "replies": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/api.Answer"
                    }
                },
                "updated_at": {
                    "type": "string"
                },
                "upvotes": {
                    "type": "integer"
                },
                "user": {
                    "type": "string"
                }
            }
        },
        "api.Coord": {
            "type": "object",
            "properties": {
                "end": {
                    "type": "integer"
                },
                "start": {
                    "type": "integer"
                }
            }
        },
        "api.Document": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "questions": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/api.Question"
                    }
                }
            }
        },
        "api.PutAnswerRequest": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string"
                },
                "parent": {
                    "type": "integer"
                },
                "question": {
                    "type": "integer"
                }
            }
        },
        "api.PutDocumentRequest": {
            "type": "object",
            "properties": {
                "coords": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/api.Coord"
                    }
                },
                "id": {
                    "type": "string"
                }
            }
        },
        "api.Question": {
            "type": "object",
            "properties": {
                "answers": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/api.Answer"
                    }
                },
                "created_at": {
                    "type": "string"
                },
                "document": {
                    "type": "string"
                },
                "end": {
                    "type": "integer"
                },
                "id": {
                    "description": "taken from from gorm.Model, so we can json strigify properly",
                    "type": "integer"
                },
                "start": {
                    "type": "integer"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "api.Vote": {
            "type": "object",
            "properties": {
                "answer": {
                    "type": "integer"
                },
                "createdAt": {
                    "description": "taken from from gorm.Model",
                    "type": "string"
                },
                "deletedAt": {
                    "$ref": "#/definitions/gorm.DeletedAt"
                },
                "updatedAt": {
                    "type": "string"
                },
                "user": {
                    "type": "string"
                },
                "vote": {
                    "type": "integer"
                }
            }
        },
        "gorm.DeletedAt": {
            "type": "object",
            "properties": {
                "time": {
                    "type": "string"
                },
                "valid": {
                    "description": "Valid is true if Time is not NULL",
                    "type": "boolean"
                }
            }
        },
        "util.ApiError": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Polleg API",
	Description:      "This is the backend API for Polleg that allows unibo students to answer exam exercises directly on the csunibo website",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}