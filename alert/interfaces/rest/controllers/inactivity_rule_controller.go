package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"microservice-alert-service/alert/application/commandservices"
	"microservice-alert-service/alert/application/queryservices"
	"microservice-alert-service/alert/interfaces/rest/resources"
	"microservice-alert-service/alert/interfaces/rest/transform"
)

type InactivityRuleController struct {
	commandService *commandservices.InactivityRuleCommandService
	queryService   *queryservices.InactivityRuleQueryService
}

func NewInactivityRuleController(commandService *commandservices.InactivityRuleCommandService, queryService *queryservices.InactivityRuleQueryService) *InactivityRuleController {
	return &InactivityRuleController{commandService: commandService, queryService: queryService}
}

func (c *InactivityRuleController) CreateRule(ctx *gin.Context) {
	var request resources.CreateInactivityRuleRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		respondBadRequest(ctx, err)
		return
	}

	cmd, err := transform.ToCreateInactivityRuleCommand(request)
	if err != nil {
		respondBadRequest(ctx, err)
		return
	}

	rule, err := c.commandService.CreateRule(ctx.Request.Context(), cmd)
	if err != nil {
		respondError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, transform.ToInactivityRuleResponse(*rule))
}

func (c *InactivityRuleController) GetRulesByUser(ctx *gin.Context) {
	userID := ctx.Param("userId")
	parsedID, err := transform.ParseUUID(userID)
	if err != nil {
		respondBadRequest(ctx, err)
		return
	}

	rules, err := c.queryService.ListByUser(ctx.Request.Context(), parsedID)
	if err != nil {
		respondError(ctx, err)
		return
	}

	response := make([]resources.InactivityRuleResponse, 0, len(rules))
	for _, rule := range rules {
		response = append(response, transform.ToInactivityRuleResponse(rule))
	}

	ctx.JSON(http.StatusOK, response)
}
