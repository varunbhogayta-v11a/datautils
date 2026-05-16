package templates

import (
	"fmt"
	"net/http"
)

// Note: The main Swagger spec is defined in cmd/server.go
// This file is kept for reference and future template usage

const swaggerHTML = `<!DOCTYPE html>
<html>
<head>
  <title>DataUtil API - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.10.5/swagger-ui.css" />
  <style>
    body { margin: 0; padding: 0; }
    .swagger-ui .topbar { display: none; }
    .swagger-ui .info .title { font-size: 2.5em; }
  </style>
</head>
<body>
	<div id="swagger-ui"></div>
	<div class="loading" id="loading">Loading Swagger UI...</div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.10.5/swagger-ui-bundle.js" charset="UTF-8"></script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.10.5/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
	<script>
		window.onload = function() {
			var loading = document.getElementById("loading");
			try {
				if (window.SwaggerUIBundle) {
					var opts = {
						url: "/swagger.yaml",
						dom_id: "#swagger-ui",
						deepLinking: true,
						docExpansion: "list",
						filter: true
					};
					if (SwaggerUIBundle.presets && SwaggerUIBundle.presets.apis && SwaggerUIBundle.standalonePreset) {
						opts.presets = [SwaggerUIBundle.presets.apis, SwaggerUIBundle.standalonePreset];
					}
					if (SwaggerUIBundle.plugins && SwaggerUIBundle.plugins.DownloadUrl) {
						opts.plugins = [SwaggerUIBundle.plugins.DownloadUrl];
					}
					window.ui = SwaggerUIBundle(opts);
				} else if (window.SwaggerUI) {
					var opts2 = {
						url: "/swagger.yaml",
						dom_id: "#swagger-ui",
						deepLinking: true,
						docExpansion: "list",
						filter: true
					};
					if (SwaggerUI.presets && SwaggerUI.presets.apis && SwaggerUI.standalonePreset) {
						opts2.presets = [SwaggerUI.presets.apis, SwaggerUI.standalonePreset];
					}
					window.ui = SwaggerUI(opts2);
				} else {
					throw new Error('Swagger UI bundle not found');
				}

				if (loading) loading.style.display = "none";
			} catch (e) {
				if (loading) loading.innerHTML = "Error loading Swagger UI: " + e.message;
				else console.error('Error loading Swagger UI:', e);
			}
		};
	</script>
</body>
</html>`

// ServeSwaggerUI serves the Swagger UI HTML
func ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, swaggerHTML)
}
