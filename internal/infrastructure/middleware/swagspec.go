package middleware

import (
	"github.com/go-openapi/runtime/middleware"
	"log"
	"net/http"
	"os"
	"strings"
)

// SwaggerUIHandler возвращает handlers для Swagger UI
func SwaggerUIHandler() http.Handler {
	return middleware.SwaggerUI(middleware.SwaggerUIOpts{
		Path:    "swagger",
		SpecURL: "/docs/swagger.json",
		Title:   "Геосервис API UI",
	}, nil)
}

// RedocHandler возвращает handlers для Redoc
func RedocHandler() http.Handler {
	return middleware.Redoc(middleware.RedocOpts{
		Path:    "docs",
		SpecURL: "/docs/swagger.yaml",
		Title:   "Геосервис API Documentation",
	}, nil)
}

// SwaggerSpecHandler для обработки файлов спецификации (JSON и YAML)
func SwaggerSpecHandler(w http.ResponseWriter, r *http.Request) {
	// Определяем какой формат запрашивается по URL
	requestPath := r.URL.Path
	var filePath string
	var contentType string

	if strings.HasSuffix(requestPath, ".json") {
		filePath = "docs/swagger.json"
		contentType = "application/json"
	} else if strings.HasSuffix(requestPath, ".yaml") || strings.HasSuffix(requestPath, ".yml") {
		filePath = "docs/swagger.yaml"
		contentType = "application/x-yaml"
	} else {
		// По умолчанию пробуем JSON, затем YAML
		jsonPath := "docs/swagger.json"
		yamlPath := "docs/swagger.yaml"

		if _, err := os.Stat(jsonPath); err == nil {
			filePath = jsonPath
			contentType = "application/json"
		} else if _, err := os.Stat(yamlPath); err == nil {
			filePath = yamlPath
			contentType = "application/x-yaml"
		} else {
			log.Printf("Файлы спецификации не найдены: %s, %s", jsonPath, yamlPath)
			http.NotFound(w, r)
			return
		}
	}

	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Файл спецификации не найден: %s", filePath)
		http.NotFound(w, r)
		return
	}

	// Устанавливаем заголовки
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Обработка preflight запросов
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("Найден файл спецификации: %s", filePath)
	http.ServeFile(w, r, filePath)
}
