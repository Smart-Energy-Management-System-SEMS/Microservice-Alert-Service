package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"microservice-alert-service/alert/infrastructure/configuration"
)

type SwaggerController struct {
	cfg configuration.Config
}

func NewSwaggerController(cfg configuration.Config) *SwaggerController {
	return &SwaggerController{cfg: cfg}
}

func (c *SwaggerController) Index(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusOK, fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Alert Service Swagger</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      SwaggerUIBundle({
        url: '/swagger/openapi.json',
        dom_id: '#swagger-ui',
        presets: [SwaggerUIBundle.presets.apis],
        layout: 'BaseLayout'
      });
    };
  </script>
</body>
</html>`))
}

func (c *SwaggerController) OpenAPI(ctx *gin.Context) {
	baseURL := fmt.Sprintf("%s://%s", requestScheme(ctx), ctx.Request.Host)

	ctx.JSON(http.StatusOK, gin.H{
		"openapi": "3.0.3",
		"info": gin.H{
			"title":       "SEMS Alert Service API",
			"version":     "1.0.0",
			"description": "Swagger basico para probar health, diagnostico Azure/Kafka y publicacion de eventos.",
		},
		"servers": []gin.H{
			{"url": baseURL},
		},
		"components": gin.H{
			"schemas": gin.H{
				"ErrorResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"error": gin.H{"type": "string", "example": "not found"},
					},
				},
				"CreateAlertRequest": gin.H{
					"type":     "object",
					"required": []string{"user_id", "device_id", "alert_type", "title", "message", "severity", "status"},
					"properties": gin.H{
						"user_id":            uuidSchema("5f8f8c44-0c72-4ca8-a80f-0a4df6b5d321"),
						"device_id":          uuidSchema("1ea6be58-5ba4-4e2f-b318-c6cbc689f4d2"),
						"threshold_id":       nullableUUIDSchema("7eb5fe5a-428f-4bb4-ae58-c8f36fd0d93d"),
						"inactivity_rule_id": nullableUUIDSchema("b08b3481-5e16-493f-b327-92d731b47ec0"),
						"alert_type":         stringSchema("threshold_exceeded"),
						"title":              stringSchema("High energy consumption detected"),
						"message":            stringSchema("Device exceeded configured threshold."),
						"severity":           stringSchema("high"),
						"status":             stringSchema("open"),
						"triggered_at":       dateTimeSchema("2026-06-18T19:15:00Z"),
					},
				},
				"UpdateAlertStatusRequest": gin.H{
					"type":     "object",
					"required": []string{"status"},
					"properties": gin.H{
						"status":      stringSchema("resolved"),
						"resolved_at": dateTimeSchema("2026-06-18T19:20:00Z"),
					},
				},
				"AlertResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"alert_id":           uuidSchema("7b2b5f03-58c7-4740-a645-034d4a16f9d3"),
						"user_id":            uuidSchema("5f8f8c44-0c72-4ca8-a80f-0a4df6b5d321"),
						"device_id":          uuidSchema("1ea6be58-5ba4-4e2f-b318-c6cbc689f4d2"),
						"threshold_id":       nullableUUIDSchema("7eb5fe5a-428f-4bb4-ae58-c8f36fd0d93d"),
						"inactivity_rule_id": nullableUUIDSchema("b08b3481-5e16-493f-b327-92d731b47ec0"),
						"alert_type":         stringSchema("threshold_exceeded"),
						"title":              stringSchema("High energy consumption detected"),
						"message":            stringSchema("Device exceeded configured threshold."),
						"severity":           stringSchema("high"),
						"status":             stringSchema("open"),
						"triggered_at":       dateTimeSchema("2026-06-18T19:15:00Z"),
						"resolved_at":        dateTimeNullableSchema("2026-06-18T19:20:00Z"),
					},
				},
				"CreateThresholdRequest": gin.H{
					"type":     "object",
					"required": []string{"user_id", "device_id", "threshold_name", "metric", "operator", "threshold_value"},
					"properties": gin.H{
						"user_id":         uuidSchema("5f8f8c44-0c72-4ca8-a80f-0a4df6b5d321"),
						"device_id":       uuidSchema("1ea6be58-5ba4-4e2f-b318-c6cbc689f4d2"),
						"threshold_name":  stringSchema("High energy threshold"),
						"metric":          stringSchema("consumption_kwh"),
						"operator":        stringSchema(">"),
						"threshold_value": gin.H{"type": "number", "format": "double", "example": 85.5},
						"active":          gin.H{"type": "boolean", "example": true},
					},
				},
				"ThresholdResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"threshold_id":    uuidSchema("7eb5fe5a-428f-4bb4-ae58-c8f36fd0d93d"),
						"user_id":         uuidSchema("5f8f8c44-0c72-4ca8-a80f-0a4df6b5d321"),
						"device_id":       uuidSchema("1ea6be58-5ba4-4e2f-b318-c6cbc689f4d2"),
						"threshold_name":  stringSchema("High energy threshold"),
						"metric":          stringSchema("consumption_kwh"),
						"operator":        stringSchema(">"),
						"threshold_value": gin.H{"type": "number", "format": "double", "example": 85.5},
						"active":          gin.H{"type": "boolean", "example": true},
					},
				},
				"CreateInactivityRuleRequest": gin.H{
					"type":     "object",
					"required": []string{"user_id", "device_id", "rule_name", "max_inactive_minutes"},
					"properties": gin.H{
						"user_id":              uuidSchema("5f8f8c44-0c72-4ca8-a80f-0a4df6b5d321"),
						"device_id":            uuidSchema("1ea6be58-5ba4-4e2f-b318-c6cbc689f4d2"),
						"rule_name":            stringSchema("Offline device rule"),
						"max_inactive_minutes": gin.H{"type": "integer", "example": 30},
						"active":               gin.H{"type": "boolean", "example": true},
					},
				},
				"InactivityRuleResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"inactivity_rule_id":   uuidSchema("b08b3481-5e16-493f-b327-92d731b47ec0"),
						"user_id":              uuidSchema("5f8f8c44-0c72-4ca8-a80f-0a4df6b5d321"),
						"device_id":            uuidSchema("1ea6be58-5ba4-4e2f-b318-c6cbc689f4d2"),
						"rule_name":            stringSchema("Offline device rule"),
						"max_inactive_minutes": gin.H{"type": "integer", "example": 30},
						"active":               gin.H{"type": "boolean", "example": true},
					},
				},
				"CreateNotificationPreferenceRequest": gin.H{
					"type":     "object",
					"required": []string{"user_id", "channel"},
					"properties": gin.H{
						"user_id":           uuidSchema("5f8f8c44-0c72-4ca8-a80f-0a4df6b5d321"),
						"channel":           stringSchema("email"),
						"enabled":           gin.H{"type": "boolean", "example": true},
						"min_severity":      stringSchema("medium"),
						"quiet_hours_start": stringSchema("22:00"),
						"quiet_hours_end":   stringSchema("06:00"),
					},
				},
				"NotificationPreferenceResponse": gin.H{
					"type": "object",
					"properties": gin.H{
						"preference_id":     uuidSchema("3d2cb0f4-12f2-4f8f-b086-0b3b7d79dc55"),
						"user_id":           uuidSchema("5f8f8c44-0c72-4ca8-a80f-0a4df6b5d321"),
						"channel":           stringSchema("email"),
						"enabled":           gin.H{"type": "boolean", "example": true},
						"min_severity":      stringSchema("medium"),
						"quiet_hours_start": dateTimeNullableSchema("0000-01-01T22:00:00Z"),
						"quiet_hours_end":   dateTimeNullableSchema("0000-01-01T06:00:00Z"),
					},
				},
				"KafkaPublishTestRequest": gin.H{
					"type": "object",
					"properties": gin.H{
						"eventType":  stringSchema("alert.created"),
						"eventId":    stringSchema("test-event-id"),
						"occurredAt": dateTimeSchema("2026-06-18T12:00:00Z"),
						"data": gin.H{
							"type":                 "object",
							"additionalProperties": true,
							"example": gin.H{
								"message": "test event from swagger",
							},
						},
					},
				},
			},
		},
		"paths": gin.H{
			"/api/v1/health": gin.H{
				"get": gin.H{
					"summary": "Health check",
					"responses": gin.H{
						"200": gin.H{"description": "Servicio operativo"},
					},
				},
			},
			"/api/v1/alerts": gin.H{
				"get": gin.H{
					"summary":   "Lista todas las alertas",
					"responses": jsonResponses("Lista de alertas", "#/components/schemas/AlertResponse", true),
				},
				"post": gin.H{
					"summary":     "Crea una alerta",
					"requestBody": jsonBody("#/components/schemas/CreateAlertRequest", true),
					"responses": gin.H{
						"201": jsonResponse("Alerta creada", "#/components/schemas/AlertResponse", false),
						"400": jsonResponse("Request invalido", "#/components/schemas/ErrorResponse", false),
						"500": jsonResponse("Error interno", "#/components/schemas/ErrorResponse", false),
					},
				},
			},
			"/api/v1/alerts/{id}": gin.H{
				"get": gin.H{
					"summary":    "Obtiene una alerta por ID",
					"parameters": []gin.H{pathParam("id", "ID de la alerta")},
					"responses": gin.H{
						"200": jsonResponse("Alerta encontrada", "#/components/schemas/AlertResponse", false),
						"400": jsonResponse("ID invalido", "#/components/schemas/ErrorResponse", false),
						"404": jsonResponse("No encontrada", "#/components/schemas/ErrorResponse", false),
					},
				},
			},
			"/api/v1/users/{userId}/alerts": gin.H{
				"get": gin.H{
					"summary":    "Lista alertas por usuario",
					"parameters": []gin.H{pathParam("userId", "ID del usuario")},
					"responses":  jsonResponses("Lista de alertas del usuario", "#/components/schemas/AlertResponse", true),
				},
			},
			"/api/v1/alerts/{id}/status": gin.H{
				"patch": gin.H{
					"summary":     "Actualiza el estado de una alerta",
					"parameters":  []gin.H{pathParam("id", "ID de la alerta")},
					"requestBody": jsonBody("#/components/schemas/UpdateAlertStatusRequest", true),
					"responses": gin.H{
						"204": gin.H{"description": "Estado actualizado"},
						"400": jsonResponse("Request invalido", "#/components/schemas/ErrorResponse", false),
						"404": jsonResponse("No encontrada", "#/components/schemas/ErrorResponse", false),
					},
				},
			},
			"/api/v1/thresholds": gin.H{
				"post": gin.H{
					"summary":     "Crea un threshold",
					"requestBody": jsonBody("#/components/schemas/CreateThresholdRequest", true),
					"responses": gin.H{
						"201": jsonResponse("Threshold creado", "#/components/schemas/ThresholdResponse", false),
						"400": jsonResponse("Request invalido", "#/components/schemas/ErrorResponse", false),
					},
				},
			},
			"/api/v1/users/{userId}/thresholds": gin.H{
				"get": gin.H{
					"summary":    "Lista thresholds por usuario",
					"parameters": []gin.H{pathParam("userId", "ID del usuario")},
					"responses":  jsonResponses("Lista de thresholds", "#/components/schemas/ThresholdResponse", true),
				},
			},
			"/api/v1/inactivity-rules": gin.H{
				"post": gin.H{
					"summary":     "Crea una regla de inactividad",
					"requestBody": jsonBody("#/components/schemas/CreateInactivityRuleRequest", true),
					"responses": gin.H{
						"201": jsonResponse("Regla creada", "#/components/schemas/InactivityRuleResponse", false),
						"400": jsonResponse("Request invalido", "#/components/schemas/ErrorResponse", false),
					},
				},
			},
			"/api/v1/users/{userId}/inactivity-rules": gin.H{
				"get": gin.H{
					"summary":    "Lista reglas de inactividad por usuario",
					"parameters": []gin.H{pathParam("userId", "ID del usuario")},
					"responses":  jsonResponses("Lista de reglas", "#/components/schemas/InactivityRuleResponse", true),
				},
			},
			"/api/v1/notification-preferences": gin.H{
				"post": gin.H{
					"summary":     "Crea una preferencia de notificacion",
					"requestBody": jsonBody("#/components/schemas/CreateNotificationPreferenceRequest", true),
					"responses": gin.H{
						"201": jsonResponse("Preferencia creada", "#/components/schemas/NotificationPreferenceResponse", false),
						"400": jsonResponse("Request invalido", "#/components/schemas/ErrorResponse", false),
					},
				},
			},
			"/api/v1/users/{userId}/notification-preferences": gin.H{
				"get": gin.H{
					"summary":    "Lista preferencias por usuario",
					"parameters": []gin.H{pathParam("userId", "ID del usuario")},
					"responses":  jsonResponses("Lista de preferencias", "#/components/schemas/NotificationPreferenceResponse", true),
				},
			},
			"/api/v1/diagnostics/validate": gin.H{
				"post": gin.H{
					"summary":     "Valida configuracion principal, DB y topics esperados para Azure/Event Hubs",
					"description": "Confirma DATABASE_URL, topicos energy.events y analytics.events, y que alerts.events sea el topic de salida.",
					"responses": gin.H{
						"200": gin.H{"description": "Validacion correcta"},
						"503": gin.H{"description": "Faltan variables o dependencias"},
					},
				},
			},
			"/api/v1/kafka/publish-test": gin.H{
				"post": gin.H{
					"summary":     "Publica un evento de prueba en alerts.events",
					"description": "Si no envias body, el servicio genera uno por defecto.",
					"requestBody": jsonBody("#/components/schemas/KafkaPublishTestRequest", false),
					"responses": gin.H{
						"202": gin.H{"description": "Evento publicado"},
						"503": gin.H{"description": "Kafka producer deshabilitado"},
					},
				},
			},
		},
	})
}

func requestScheme(ctx *gin.Context) string {
	if ctx.Request.TLS != nil {
		return "https"
	}
	if forwardedProto := ctx.GetHeader("X-Forwarded-Proto"); forwardedProto != "" {
		return forwardedProto
	}
	return "http"
}

func stringSchema(example string) gin.H {
	return gin.H{"type": "string", "example": example}
}

func uuidSchema(example string) gin.H {
	return gin.H{"type": "string", "format": "uuid", "example": example}
}

func nullableUUIDSchema(example string) gin.H {
	return gin.H{"type": "string", "format": "uuid", "nullable": true, "example": example}
}

func dateTimeSchema(example string) gin.H {
	return gin.H{"type": "string", "format": "date-time", "example": example}
}

func dateTimeNullableSchema(example string) gin.H {
	return gin.H{"type": "string", "format": "date-time", "nullable": true, "example": example}
}

func jsonBody(schemaRef string, required bool) gin.H {
	return gin.H{
		"required": required,
		"content": gin.H{
			"application/json": gin.H{
				"schema": gin.H{
					"$ref": schemaRef,
				},
			},
		},
	}
}

func jsonResponse(description string, schemaRef string, isArray bool) gin.H {
	schema := gin.H{"$ref": schemaRef}
	if isArray {
		schema = gin.H{
			"type":  "array",
			"items": gin.H{"$ref": schemaRef},
		}
	}
	return gin.H{
		"description": description,
		"content": gin.H{
			"application/json": gin.H{
				"schema": schema,
			},
		},
	}
}

func jsonResponses(description string, schemaRef string, isArray bool) gin.H {
	return gin.H{
		"200": jsonResponse(description, schemaRef, isArray),
		"400": jsonResponse("Request invalido", "#/components/schemas/ErrorResponse", false),
		"500": jsonResponse("Error interno", "#/components/schemas/ErrorResponse", false),
	}
}

func pathParam(name string, description string) gin.H {
	return gin.H{
		"name":        name,
		"in":          "path",
		"required":    true,
		"description": description,
		"schema": gin.H{
			"type":   "string",
			"format": "uuid",
		},
	}
}
