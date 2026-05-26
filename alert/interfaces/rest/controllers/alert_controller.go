package controllers

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "microservice-alert-service/alert/application/commandservices"
    "microservice-alert-service/alert/application/queryservices"
    "microservice-alert-service/alert/interfaces/rest/resources"
    "microservice-alert-service/alert/interfaces/rest/transform"
)

type AlertController struct {
    commandService *commandservices.AlertCommandService
    queryService   *queryservices.AlertQueryService
}

func NewAlertController(commandService *commandservices.AlertCommandService, queryService *queryservices.AlertQueryService) *AlertController {
    return &AlertController{commandService: commandService, queryService: queryService}
}

func (c *AlertController) GetAlerts(ctx *gin.Context) {
    alerts, err := c.queryService.ListAll(ctx.Request.Context())
    if err != nil {
        respondError(ctx, err)
        return
    }

    response := make([]resources.AlertResponse, 0, len(alerts))
    for _, alert := range alerts {
        response = append(response, transform.ToAlertResponse(alert))
    }

    ctx.JSON(http.StatusOK, response)
}

func (c *AlertController) GetAlertByID(ctx *gin.Context) {
    alertID := ctx.Param("id")
    alert, err := c.queryService.GetByID(ctx.Request.Context(), parsePathID(alertID, ctx))
    if err != nil {
        respondError(ctx, err)
        return
    }

    ctx.JSON(http.StatusOK, transform.ToAlertResponse(*alert))
}

func (c *AlertController) GetAlertsByUser(ctx *gin.Context) {
    userID := ctx.Param("userId")
    parsedID, err := transform.ParseUUID(userID)
    if err != nil {
        respondBadRequest(ctx, err)
        return
    }

    alerts, err := c.queryService.ListByUser(ctx.Request.Context(), parsedID)
    if err != nil {
        respondError(ctx, err)
        return
    }

    response := make([]resources.AlertResponse, 0, len(alerts))
    for _, alert := range alerts {
        response = append(response, transform.ToAlertResponse(alert))
    }

    ctx.JSON(http.StatusOK, response)
}

func (c *AlertController) CreateAlert(ctx *gin.Context) {
    var request resources.CreateAlertRequest
    if err := ctx.ShouldBindJSON(&request); err != nil {
        respondBadRequest(ctx, err)
        return
    }

    cmd, err := transform.ToCreateAlertCommand(request)
    if err != nil {
        respondBadRequest(ctx, err)
        return
    }

    alert, err := c.commandService.CreateAlert(ctx.Request.Context(), cmd)
    if err != nil {
        respondError(ctx, err)
        return
    }

    ctx.JSON(http.StatusCreated, transform.ToAlertResponse(*alert))
}

func (c *AlertController) UpdateAlertStatus(ctx *gin.Context) {
    var request resources.UpdateAlertStatusRequest
    if err := ctx.ShouldBindJSON(&request); err != nil {
        respondBadRequest(ctx, err)
        return
    }

    cmd, err := transform.ToUpdateAlertStatusCommand(ctx.Param("id"), request)
    if err != nil {
        respondBadRequest(ctx, err)
        return
    }

    if err := c.commandService.UpdateAlertStatus(ctx.Request.Context(), cmd); err != nil {
        respondError(ctx, err)
        return
    }

    ctx.Status(http.StatusNoContent)
}
