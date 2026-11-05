package controllers

import (
	"net/http"
	"strconv"

	"iot-lock/services"

	"github.com/gin-gonic/gin"
)

type CardController struct {
	service *services.CardService
}

func NewCardController(service *services.CardService) *CardController {
	return &CardController{service: service}
}

func (ctrl *CardController) Create(ctx *gin.Context) {
	var payload struct {
		CardUID string `json:"card_uid"`
		UserID  uint   `json:"user_id"`
	}

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}

	card, err := ctrl.service.Create(payload.CardUID, payload.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, card)
}

func (ctrl *CardController) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID do cartão inválido"})
		return
	}

	card, err := ctrl.service.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, card)
}

func (ctrl *CardController) GetByUID(ctx *gin.Context) {
	uid := ctx.Param("uid")
	card, err := ctrl.service.GetByUID(uid)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, card)
}

func (ctrl *CardController) List(ctx *gin.Context) {
	cards, err := ctrl.service.List()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, cards)
}

func (ctrl *CardController) UpdateStatus(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID do cartão inválido"})
		return
	}

	var payload struct {
		Active bool `json:"active"`
	}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "dados inválidos"})
		return
	}

	card, err := ctrl.service.UpdateStatus(uint(id), payload.Active)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, card)
}

func (ctrl *CardController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID do cartão inválido"})
		return
	}

	if err := ctrl.service.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "cartão removido com sucesso"})
}
