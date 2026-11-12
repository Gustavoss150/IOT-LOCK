package main

import (
	"log"
	"os"

	"iot-lock/config"
	"iot-lock/controllers"
	"iot-lock/migrations"
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

	if err := migrations.Run(db); err != nil {
		log.Fatalf("validar e migrar banco de dados: %v", err)
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
	telegramService := services.NewTelegramService()

	mqttBroker := services.NewMQTTBrokerService(userService, cardService, accessLogService, telegramService)
	mqttBroker.StartMQTT()

	userController := controllers.NewUserController(userService)
	cardController := controllers.NewCardController(cardService, mqttBroker)
	accessLogController := controllers.NewAccessLogController(accessLogService, mqttBroker)

	r := gin.Default()

	// Habilita CORS para permitir chamadas do front-end hospedado na Vercel para o back-end no Render
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Olá, Gin!",
		})
	})

	api := r.Group("/api")
	routes.SetupUserRoutes(api, userController)
	routes.SetupCardRoutes(api, cardController)
	routes.SetupAccessLogRoutes(api, accessLogController)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("iniciar servidor: %v", err)
	}
}
