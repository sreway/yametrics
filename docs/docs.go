// Package docs GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/": {
            "get": {
                "description": "show saved metrics",
                "produces": [
                    "text/html"
                ],
                "summary": "Show metrics",
                "operationId": "serverIndex",
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "501": {
                        "description": "Not Implemented"
                    }
                }
            }
        },
        "/ping": {
            "get": {
                "description": "health check metric storage",
                "summary": "ping",
                "operationId": "serverPing",
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        },
        "/update/": {
            "post": {
                "description": "set update metric",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "update metric json",
                "operationId": "serverSetUpdateMetricJSON",
                "parameters": [
                    {
                        "description": "metric data",
                        "name": "m",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/metrics.Metric"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/metrics.Metric"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "501": {
                        "description": "Not Implemented"
                    }
                }
            }
        },
        "/update/{metricType}/{metricName}/{metricValue}": {
            "post": {
                "description": "add or update metric",
                "produces": [
                    "application/json"
                ],
                "summary": "add/update metric",
                "operationId": "serverSetUpdateMetric",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Metric type",
                        "name": "metricType",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Metric name",
                        "name": "metricName",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Metric value",
                        "name": "metricValue",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "501": {
                        "description": "Not Implemented"
                    }
                }
            }
        },
        "/updates/": {
            "post": {
                "description": "set update metrics",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "update metrics json",
                "operationId": "serverBatchMetrics",
                "parameters": [
                    {
                        "description": "metrics data",
                        "name": "m",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/metrics.Metric"
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/metrics.Metric"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "501": {
                        "description": "Not Implemented"
                    }
                }
            }
        },
        "/value/": {
            "post": {
                "description": "update metric value",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "metric value json",
                "operationId": "serverMetricValueJSON",
                "parameters": [
                    {
                        "description": "metric data",
                        "name": "m",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/metrics.Metric"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/metrics.Metric"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "501": {
                        "description": "Not Implemented"
                    }
                }
            }
        },
        "/value/{metricType}/{metricName}": {
            "get": {
                "description": "get metric value",
                "produces": [
                    "text/html"
                ],
                "summary": "get metric value",
                "operationId": "serverGetMetricValue",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Metric name",
                        "name": "metricName",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Metric type",
                        "name": "metricType",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "501": {
                        "description": "Not Implemented"
                    }
                }
            }
        }
    },
    "definitions": {
        "metrics.Metric": {
            "type": "object",
            "properties": {
                "delta": {
                    "type": "integer"
                },
                "hash": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "type": {
                    "type": "string"
                },
                "value": {
                    "type": "number"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
