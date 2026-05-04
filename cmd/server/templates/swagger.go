package templates

import (
	"fmt"
	"net/http"
)

const swaggerJSON = `{
  "openapi": "3.0.3",
  "info": {
    "title": "DataUtil API",
    "description": "RESTful API for dataset processing, filtering, transformation, validation, and database operations. Supports CSV, JSON, XML, and Excel formats with JWT-based authentication.",
    "version": "1.0.0",
    "contact": {
      "name": "API Support",
      "email": "support@datautil.io"
    }
  },
  "servers": [
    {
      "url": "http://localhost:8080",
      "description": "Local development server"
    }
  ],
  "security": [
    {
      "bearerAuth": []
    }
  ],
  "tags": [
    {"name": "Authentication", "description": "User registration and login endpoints"},
    {"name": "Data Operations", "description": "Filter, transform, validate, and export datasets"},
    {"name": "Database", "description": "Query and manipulate database tables"},
    {"name": "Users", "description": "User management and logs"}
  ],
  "paths": {
    "/api/health": {
      "get": {
        "summary": "Health check endpoint",
        "tags": ["Health"],
        "responses": {
          "200": {
            "description": "Server is healthy",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "status": {"type": "string"},
                    "timestamp": {"type": "string"}
                  }
                }
              }
            }
          }
        }
      }
    },
    "/api/auth/register": {
      "post": {
        "summary": "Register a new user",
        "tags": ["Authentication"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["username", "email", "password"],
                "properties": {
                  "username": {"type": "string", "description": "Unique username"},
                  "email": {"type": "string", "format": "email", "description": "User email address"},
                  "password": {"type": "string", "format": "password", "description": "User password"}
                }
              }
            }
          }
        },
        "responses": {
          "201": {"description": "User created successfully"},
          "400": {"description": "Invalid request or user already exists"}
        }
      }
    },
    "/api/auth/login": {
      "post": {
        "summary": "Login and get JWT token",
        "tags": ["Authentication"],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["email", "password"],
                "properties": {
                  "email": {"type": "string", "format": "email"},
                  "password": {"type": "string", "format": "password"}
                }
              }
            }
          }
        },
        "responses": {
          "200": {"description": "Login successful"},
          "401": {"description": "Invalid credentials"}
        }
      }
    },
    "/api/data/filter": {
      "post": {
        "summary": "Filter rows from dataset",
        "tags": ["Data Operations"],
        "security": [{"bearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["input"],
                "properties": {
                  "input": {"type": "string", "description": "Input file path (required)"},
                  "where": {"type": "string", "description": "Filter condition"},
                  "select": {"type": "string", "description": "Columns to select"},
                  "invert": {"type": "boolean"},
                  "output": {"type": "string"},
                  "format": {"type": "string", "enum": ["csv", "json", "xml"]}
                }
              }
            }
          }
        },
        "responses": {
          "200": {"description": "Filtered dataset"},
          "401": {"description": "Authentication required"}
        }
      }
    },
    "/api/data/transform": {
      "post": {
        "summary": "Transform dataset",
        "tags": ["Data Operations"],
        "security": [{"bearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["input"],
                "properties": {
                  "input": {"type": "string"},
                  "add": {"type": "string"},
                  "remove": {"type": "string"},
                  "rename": {"type": "string"},
                  "output": {"type": "string"},
                  "format": {"type": "string", "enum": ["csv", "json", "xml"]}
                }
              }
            }
          }
        },
        "responses": {
          "200": {"description": "Transformed dataset"}
        }
      }
    },
    "/api/data/validate": {
      "post": {
        "summary": "Validate dataset",
        "tags": ["Data Operations"],
        "security": [{"bearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["input"],
                "properties": {
                  "input": {"type": "string"},
                  "required": {"type": "string"},
                  "types": {"type": "string"}
                }
              }
            }
          }
        },
        "responses": {
          "200": {"description": "Validation result"}
        }
      }
    },
    "/api/data/export": {
      "post": {
        "summary": "Export dataset",
        "tags": ["Data Operations"],
        "security": [{"bearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["input", "to"],
                "properties": {
                  "input": {"type": "string"},
                  "output": {"type": "string"},
                  "to": {"type": "string", "enum": ["csv", "json", "xml"]},
                  "pretty": {"type": "boolean"},
                  "stats": {"type": "boolean"}
                }
              }
            }
          }
        },
        "responses": {
          "200": {"description": "Exported data"}
        }
      }
    },
    "/api/data/import": {
      "post": {
        "summary": "Import file to database",
        "tags": ["Data Operations"],
        "security": [{"bearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["source", "table"],
                "properties": {
                  "source": {"type": "string"},
                  "table": {"type": "string"},
                  "create": {"type": "boolean"},
                  "ifNotExists": {"type": "boolean"},
                  "truncate": {"type": "boolean"}
                }
              }
            }
          }
        },
        "responses": {
          "200": {"description": "Import successful"}
        }
      }
    },
    "/api/db/query": {
      "post": {
        "summary": "Execute SQL query",
        "tags": ["Database"],
        "security": [{"bearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["sql"],
                "properties": {
                  "sql": {"type": "string"},
                  "limit": {"type": "integer", "default": 100}
                }
              }
            }
          }
        },
        "responses": {
          "200": {"description": "Query results"}
        }
      }
    },
    "/api/db/insert": {
      "post": {
        "summary": "Insert row",
        "tags": ["Database"],
        "security": [{"bearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["table", "values"],
                "properties": {
                  "table": {"type": "string"},
                  "values": {"type": "string"}
                }
              }
            }
          }
        },
        "responses": {
          "201": {"description": "Insert successful"}
        }
      }
    },
    "/api/db/update": {
      "post": {
        "summary": "Update rows",
        "tags": ["Database"],
        "security": [{"bearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["table", "set"],
                "properties": {
                  "table": {"type": "string"},
                  "set": {"type": "string"},
                  "where": {"type": "string"}
                }
              }
            }
          }
        },
        "responses": {
          "200": {"description": "Update successful"}
        }
      }
    },
    "/api/db/delete": {
      "post": {
        "summary": "Delete rows",
        "tags": ["Database"],
        "security": [{"bearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["table"],
                "properties": {
                  "table": {"type": "string"},
                  "where": {"type": "string"}
                }
              }
            }
          }
        },
        "responses": {
          "200": {"description": "Delete successful"}
        }
      }
    },
    "/api/users": {
      "get": {
        "summary": "List users",
        "tags": ["Users"],
        "security": [{"bearerAuth": []}],
        "responses": {
          "200": {"description": "User list"}
        }
      }
    },
    "/api/logs": {
      "get": {
        "summary": "Get logs",
        "tags": ["Users"],
        "security": [{"bearerAuth": []}],
        "responses": {
          "200": {"description": "Operation logs"}
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "bearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
      }
    }
  }
}`

const swaggerHTML = `<!DOCTYPE html>
<html>
<head>
  <title>DataUtil API - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui.css" />
  <style>
    body { margin: 0; padding: 0; }
    .swagger-ui .topbar { display: none; }
    .swagger-ui .info .title { font-size: 2.5em; }
    .loading { padding: 40px; text-align: center; font-family: sans-serif; color: #666; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <div class="loading" id="loading">Loading Swagger UI...</div>
  <script src="https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui-bundle.js" charset="UTF-8"></script>
  <script>
    window.onload = function() {
      setTimeout(function() {
        var loading = document.getElementById("loading");
        try {
          var Swagger = window.SwaggerUIBundle;
          if (typeof Swagger === "undefined") {
            loading.innerHTML = "Error: SwaggerUI library not loaded. <a href='https://unpkg.com/swagger-ui-dist@5.10.5/swagger-ui-bundle.js' target='_blank'>Check CDN</a>";
            return;
          }
          loading.style.display = "none";
          var ui = Swagger({
            url: "/swagger.json",
            dom_id: "#swagger-ui",
            deepLinking: false,
            presets: [Swagger.presets.apis],
            layout: "BaseLayout"
          });
          window.ui = ui;
        } catch(e) {
          loading.innerHTML = "Error: " + e.message;
        }
      }, 500);
    };
  </script>
</body>
</html>`

func ServeSwaggerSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "inline")
	fmt.Fprint(w, swaggerJSON)
}

func ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, swaggerHTML)
}
