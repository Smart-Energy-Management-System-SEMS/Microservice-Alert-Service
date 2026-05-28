package rest

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"microservice-alert-service/alert/application/commandservices"
	"microservice-alert-service/alert/application/queryservices"
	"microservice-alert-service/alert/interfaces/rest/controllers"
)

func NewRouter(
	alertCommand *commandservices.AlertCommandService,
	alertQuery *queryservices.AlertQueryService,
	thresholdCommand *commandservices.ThresholdCommandService,
	thresholdQuery *queryservices.ThresholdQueryService,
	inactivityCommand *commandservices.InactivityRuleCommandService,
	inactivityQuery *queryservices.InactivityRuleQueryService,
	preferenceCommand *commandservices.NotificationPreferenceCommandService,
	preferenceQuery *queryservices.NotificationPreferenceQueryService,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(cors.Default())

	alertController := controllers.NewAlertController(alertCommand, alertQuery)
	thresholdController := controllers.NewThresholdController(thresholdCommand, thresholdQuery)
	inactivityController := controllers.NewInactivityRuleController(inactivityCommand, inactivityQuery)
	preferenceController := controllers.NewNotificationPreferenceController(preferenceCommand, preferenceQuery)

	api := router.Group("/api/v1")
	{
		api.GET("/health", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{"status": "ok", "service": "alert-service"})
		})

		api.GET("/alerts", alertController.GetAlerts)
		api.GET("/alerts/:id", alertController.GetAlertByID)
		api.GET("/users/:userId/alerts", alertController.GetAlertsByUser)
		api.POST("/alerts", alertController.CreateAlert)
		api.PATCH("/alerts/:id/status", alertController.UpdateAlertStatus)

		api.POST("/thresholds", thresholdController.CreateThreshold)
		api.GET("/users/:userId/thresholds", thresholdController.GetThresholdsByUser)

		api.POST("/inactivity-rules", inactivityController.CreateRule)
		api.GET("/users/:userId/inactivity-rules", inactivityController.GetRulesByUser)

		api.POST("/notification-preferences", preferenceController.CreatePreference)
		api.GET("/users/:userId/notification-preferences", preferenceController.GetPreferencesByUser)
	}

	return router
}
