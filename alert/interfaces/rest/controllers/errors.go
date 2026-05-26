package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"microservice-alert-service/alert/interfaces/rest/resources"
	"microservice-alert-service/alert/shared/domain"
)

func respondError(ctx *gin.Context, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		ctx.JSON(http.StatusNotFound, resources.ErrorResponse{Error: "not found"})
		return
	}

	ctx.JSON(http.StatusInternalServerError, resources.ErrorResponse{Error: err.Error()})
}

func respondBadRequest(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, resources.ErrorResponse{Error: err.Error()})
}
