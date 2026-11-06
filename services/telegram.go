package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type TelegramService struct {
	token  string
	chatID string
	client *http.Client
}

func cleanEnvValue(val string) string {
	val = strings.TrimSpace(val)
	val = strings.Trim(val, `"'`)
	return strings.TrimSpace(val)
}

func NewTelegramService() *TelegramService {
	LoadEnv()

	token := cleanEnvValue(os.Getenv("TELEGRAM_TOKEN"))
	if token == "" {
		token = cleanEnvValue(os.Getenv("TELEGRAM_BOT_TOKEN"))
	}
	chatID := cleanEnvValue(os.Getenv("TELEGRAM_CHAT_ID"))

	if token == "" || chatID == "" {
		log.Println("⚠️ TELEGRAM_TOKEN ou TELEGRAM_CHAT_ID não definidos no .env (notificações Telegram desativadas)")
	} else {
		log.Printf("📲 Notificações do Telegram ativadas (Chat ID: %s)", chatID)
	}

	return &TelegramService{
		token:  token,
		chatID: chatID,
		client: &http.Client{Timeout: 8 * time.Second},
	}
}

func (t *TelegramService) SendAlertAsync(message string) {
	if t == nil || t.token == "" || t.chatID == "" {
		return
	}

	go func(msg string) {
		url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
		payload := map[string]string{
			"chat_id":    t.chatID,
			"text":       msg,
			"parse_mode": "HTML",
		}

		body, err := json.Marshal(payload)
		if err != nil {
			log.Printf("⚠️ [Telegram] Erro ao serializar mensagem: %v", err)
			return
		}

		resp, err := t.client.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			log.Printf("⚠️ [Telegram] Erro ao enviar alerta: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			log.Printf("⚠️ [Telegram] API respondeu com status %d", resp.StatusCode)
			return
		}

		log.Println("📲 [Telegram] Alerta enviado com sucesso!")
	}(message)
}
