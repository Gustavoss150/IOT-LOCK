package controllers

import (
	"net/http"
	"strconv"

	"iot-lock/services"

	"github.com/gin-gonic/gin"
)

type AccessLogController struct {
	service *services.AccessLogService
	broker  *services.MQTTBrokerService
}

func NewAccessLogController(service *services.AccessLogService, broker *services.MQTTBrokerService) *AccessLogController {
	return &AccessLogController{service: service, broker: broker}
}

func (ctrl *AccessLogController) RemoteUnlock(ctx *gin.Context) {
	if ctrl.broker == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "broker MQTT não disponível"})
		return
	}

	logEntry, err := ctrl.broker.RemoteUnlock()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, logEntry)
}

func (ctrl *AccessLogController) Create(ctx *gin.Context) {
	var payload struct {
		CardID     *uint  `json:"card_id"`
		CardUID    string `json:"card_uid"`
		Authorized bool   `json:"authorized"`
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}

	logEntry, err := ctrl.service.Create(payload.CardID, payload.CardUID, payload.Authorized)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, logEntry)
}

func (ctrl *AccessLogController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID do registro inválido"})
		return
	}

	logEntry, err := ctrl.service.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, logEntry)
}

func (ctrl *AccessLogController) ListRecent(ctx *gin.Context) {
	limit, offset := parsePaginationParams(ctx)
	logs, err := ctrl.service.ListRecent(limit, offset)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, logs)
}

func (ctrl *AccessLogController) ListByCardID(ctx *gin.Context) {
	cardID, err := strconv.ParseUint(ctx.Param("cardID"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID do cartão inválido"})
		return
	}

	limit, offset := parsePaginationParams(ctx)
	logs, err := ctrl.service.ListByCardID(uint(cardID), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, logs)
}

func parsePaginationParams(ctx *gin.Context) (int, int) {
	limit := 1000
	offset := 0

	if value := ctx.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if value := ctx.Query("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
