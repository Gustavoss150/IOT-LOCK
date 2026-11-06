package routes

import (
	"iot-lock/controllers"

	"github.com/gin-gonic/gin"
)

func SetupCardRoutes(router *gin.RouterGroup, controller *controllers.CardController) {
	cards := router.Group("/cards")
	cards.POST("", controller.Create)
	cards.GET("", controller.List)
	cards.POST("/enrollment/start", controller.StartEnrollment)
	cards.GET("/enrollment/status", controller.GetEnrollmentStatus)
	cards.POST("/enrollment/cancel", controller.CancelEnrollment)
	cards.GET(":id", controller.GetByID)
	cards.GET("/uid/:uid", controller.GetByUID)
	cards.PATCH(":id/status", controller.UpdateStatus)
	cards.DELETE(":id", controller.Delete)
}
