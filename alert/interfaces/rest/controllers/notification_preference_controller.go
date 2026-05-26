package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"microservice-alert-service/alert/application/commandservices"
	"microservice-alert-service/alert/application/queryservices"
	"microservice-alert-service/alert/interfaces/rest/resources"
	"microservice-alert-service/alert/interfaces/rest/transform"
)

type NotificationPreferenceController struct {
	commandService *commandservices.NotificationPreferenceCommandService
	queryService   *queryservices.NotificationPreferenceQueryService
}

func NewNotificationPreferenceController(commandService *commandservices.NotificationPreferenceCommandService, queryService *queryservices.NotificationPreferenceQueryService) *NotificationPreferenceController {
	return &NotificationPreferenceController{commandService: commandService, queryService: queryService}
}

func (c *NotificationPreferenceController) CreatePreference(ctx *gin.Context) {
	var request resources.CreateNotificationPreferenceRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		respondBadRequest(ctx, err)
		return
	}

	cmd, err := transform.ToCreateNotificationPreferenceCommand(request)
	if err != nil {
		respondBadRequest(ctx, err)
		return
	}

	preference, err := c.commandService.CreatePreference(ctx.Request.Context(), cmd)
	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, transform.ToNotificationPreferenceResponse(*preference))
}

func (c *NotificationPreferenceController) GetPreferencesByUser(ctx *gin.Context) {
	userID := ctx.Param("userId")
	parsedID, err := transform.ParseUUID(userID)
	if err != nil {
		respondBadRequest(ctx, err)
		return
	}

	preferences, err := c.queryService.ListByUser(ctx.Request.Context(), parsedID)
	if err != nil {
		respondError(ctx, err)
		return
	}

	response := make([]resources.NotificationPreferenceResponse, 0, len(preferences))
	for _, preference := range preferences {
		response = append(response, transform.ToNotificationPreferenceResponse(preference))
	}

	ctx.JSON(http.StatusOK, response)
}
