package routes

import (
	"iot-lock/controllers"

	"github.com/gin-gonic/gin"
)

func SetupAccessLogRoutes(router *gin.RouterGroup, controller *controllers.AccessLogController) {
	accessLogs := router.Group("/access-logs")
	accessLogs.POST("", controller.Create)
	accessLogs.GET(":id", controller.GetByID)
	accessLogs.GET("/recent", controller.ListRecent)
	accessLogs.GET("/card/:cardID", controller.ListByCardID)
}
