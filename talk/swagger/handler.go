package swagger

import (
	"html/template"
	"net/http"
	"strings"

	"go.zoe.im/x/talk"
)

const swaggerUITemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "{{.SpecURL}}",
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

type Handler struct {
	config    Config
	generator *Generator
	endpoints []*talk.Endpoint
	specCache []byte
	uiTmpl    *template.Template
}

func NewHandler(cfg Config) *Handler {
	tmpl := template.Must(template.New("swagger-ui").Parse(swaggerUITemplate))
	return &Handler{
		config:    cfg,
		generator: NewGenerator(cfg),
		uiTmpl:    tmpl,
	}
}

func (h *Handler) SetEndpoints(endpoints []*talk.Endpoint) {
	h.endpoints = endpoints
	h.specCache = nil
}

func (h *Handler) getSpec() ([]byte, error) {
	if h.specCache != nil {
		return h.specCache, nil
	}

	spec, err := h.generator.GenerateJSON(h.endpoints)
	if err != nil {
		return nil, err
	}
	h.specCache = spec
	return spec, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, h.config.Path)
	if path == "" || path == "/" {
		h.serveUI(w, r)
		return
	}
	if path == "/openapi.json" || path == "/swagger.json" {
		h.serveSpec(w, r)
		return
	}
	http.NotFound(w, r)
}

func (h *Handler) serveUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := struct {
		Title   string
		SpecURL string
	}{
		Title:   h.config.Title,
		SpecURL: h.config.Path + "/openapi.json",
	}

	if err := h.uiTmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render Swagger UI", http.StatusInternalServerError)
	}
}

func (h *Handler) serveSpec(w http.ResponseWriter, r *http.Request) {
	spec, err := h.getSpec()
	if err != nil {
		http.Error(w, "Failed to generate OpenAPI spec", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(spec)
}

func (h *Handler) BasePath() string {
	return h.config.Path
}

func (h *Handler) UIPath() string {
	return h.config.Path + "/"
}

func (h *Handler) SpecPath() string {
	return h.config.Path + "/openapi.json"
}
