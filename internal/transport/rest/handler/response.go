package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type responseError struct {
	Message string `json:"message"`
}

type statusResponse struct {
	Status string `json:"status"`
}

func newErrorResponse(ctx *gin.Context, statusCode int, error string, message string) {
	logrus.Error(error)
	ctx.AbortWithStatusJSON(statusCode, responseError{message})
}
