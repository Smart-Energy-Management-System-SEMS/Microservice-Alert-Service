package controllers

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "microservice-alert-service/alert/application/commandservices"
    "microservice-alert-service/alert/application/queryservices"
    "microservice-alert-service/alert/interfaces/rest/resources"
    "microservice-alert-service/alert/interfaces/rest/transform"
)

type ThresholdController struct {
    commandService *commandservices.ThresholdCommandService
    queryService   *queryservices.ThresholdQueryService
}

func NewThresholdController(commandService *commandservices.ThresholdCommandService, queryService *queryservices.ThresholdQueryService) *ThresholdController {
    return &ThresholdController{commandService: commandService, queryService: queryService}
}

func (c *ThresholdController) CreateThreshold(ctx *gin.Context) {
    var request resources.CreateThresholdRequest
    if err := ctx.ShouldBindJSON(&request); err != nil {
        respondBadRequest(ctx, err)
        return
    }

    cmd, err := transform.ToCreateThresholdCommand(request)
    if err != nil {
        respondBadRequest(ctx, err)
        return
    }

    threshold, err := c.commandService.CreateThreshold(ctx.Request.Context(), cmd)
    if err != nil {
        respondError(ctx, err)
        return
    }

    ctx.JSON(http.StatusCreated, transform.ToThresholdResponse(*threshold))
}

func (c *ThresholdController) GetThresholdsByUser(ctx *gin.Context) {
    userID := ctx.Param("userId")
    parsedID, err := transform.ParseUUID(userID)
    if err != nil {
        respondBadRequest(ctx, err)
        return
    }

    thresholds, err := c.queryService.ListByUser(ctx.Request.Context(), parsedID)
    if err != nil {
        respondError(ctx, err)
        return
    }

    response := make([]resources.ThresholdResponse, 0, len(thresholds))
    for _, threshold := range thresholds {
        response = append(response, transform.ToThresholdResponse(threshold))
    }

    ctx.JSON(http.StatusOK, response)
}
