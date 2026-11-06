package main

import (
	"log"

	"iot-lock/config"
	"iot-lock/controllers"
	"iot-lock/entities"
	repoImpl "iot-lock/repositories/impl"
	"iot-lock/routes"
	"iot-lock/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&entities.User{}, &entities.Card{}, &entities.AccessLog{}); err != nil {
		log.Fatalf("migrar banco de dados: %v", err)
	}

	userRepo, err := repoImpl.InitUserRepository(db)
	if err != nil {
		log.Fatalf("inicializar repositório de usuários: %v", err)
	}

	cardRepo, err := repoImpl.InitCardRepository(db)
	if err != nil {
		log.Fatalf("inicializar repositório de cartões: %v", err)
	}

	accessLogRepo, err := repoImpl.InitAccessLogRepository(db)
	if err != nil {
		log.Fatalf("inicializar repositório de logs: %v", err)
	}

	userService := services.NewUserService(userRepo)
	cardService := services.NewCardService(cardRepo)
	accessLogService := services.NewAccessLogService(accessLogRepo)

	mqttBroker := services.NewMQTTBrokerService(cardService, accessLogService)
	mqttBroker.StartMQTT()

	userController := controllers.NewUserController(userService)
	cardController := controllers.NewCardController(cardService)
	accessLogController := controllers.NewAccessLogController(accessLogService)

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Olá, Gin!",
		})
	})

	api := r.Group("/api")
	routes.SetupUserRoutes(api, userController)
	routes.SetupCardRoutes(api, cardController)
	routes.SetupAccessLogRoutes(api, accessLogController)

	if err := r.Run(); err != nil {
		log.Fatalf("iniciar servidor: %v", err)
	}
}
