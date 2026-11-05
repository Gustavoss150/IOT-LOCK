package routes

import (
	"iot-lock/controllers"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(router *gin.RouterGroup, controller *controllers.UserController) {
	users := router.Group("/users")
	users.POST("", controller.Create)
	users.GET("", controller.List)
	users.GET(":id", controller.GetByID)
	users.PUT(":id", controller.Update)
	users.DELETE(":id", controller.Delete)
}
