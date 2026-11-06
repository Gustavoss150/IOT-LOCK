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
	// Garante que card_id aceite NULL caso a tabela já tenha sido criada anteriormente com NOT NULL
	_ = db.Exec("ALTER TABLE access_logs ALTER COLUMN card_id DROP NOT NULL").Error
	// Sincroniza sequences de ID caso registros tenham sido inseridos manualmente
	_ = db.Exec("SELECT setval(pg_get_serial_sequence('users', 'id'), COALESCE((SELECT MAX(id) FROM users), 1), (SELECT COUNT(*) > 0 FROM users))").Error
	_ = db.Exec("SELECT setval(pg_get_serial_sequence('cards', 'id'), COALESCE((SELECT MAX(id) FROM cards), 1), (SELECT COUNT(*) > 0 FROM cards))").Error

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
