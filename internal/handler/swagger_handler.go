package handler

import (
	"os"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type SwaggerHandler struct {
	openAPIPath string
}

func NewSwaggerHandler(openAPIPath string) *SwaggerHandler {
	return &SwaggerHandler{openAPIPath: openAPIPath}
}

func (h *SwaggerHandler) ServeSwaggerUI(c *gin.Context) {
	swaggerHTML := `<!DOCTYPE html>
<html>
<head>
    <title>API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin:0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger/openapi.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML))
}

func (h *SwaggerHandler) ServeOpenAPI(c *gin.Context) {
	data, err := os.ReadFile(h.openAPIPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Failed to read OpenAPI spec: " + err.Error(),
			},
		})
		return
	}

	if strings.HasSuffix(h.openAPIPath, ".yml") || strings.HasSuffix(h.openAPIPath, ".yaml") {
		var yamlData interface{}
		if err := yaml.Unmarshal(data, &yamlData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "PARSE_ERROR",
					"message": "Failed to parse YAML: " + err.Error(),
				},
			})
			return
		}
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, yamlData)
	} else {
		c.Header("Content-Type", "application/json")
		c.Data(http.StatusOK, "application/json", data)
	}
}
