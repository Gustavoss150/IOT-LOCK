package main

import (
	"log"

	"iot-lock/config"
	"iot-lock/entities"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
)

func main() {
	// Carrega .env localmente; em produção, configure DATABASE_URL no ambiente.
	_ = godotenv.Load()

	db, err := config.ConnectDatabase()
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&entities.User{}, &entities.Card{}, &entities.AccessLog{}); err != nil {
		log.Fatalf("migrar banco de dados: %v", err)
	}

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Olá, Gin!",
		})
	})

	if err := r.Run(); err != nil {
		log.Fatalf("iniciar servidor: %v", err)
	}
}
