package services

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

const (
	TopicAccessRequest  = "smartlock/acesso/solicitacao" // ESP -> Backend
	TopicAccessResponse = "smartlock/acesso/resposta"    // Backend -> ESP
)

type AccessRequestPayload struct {
	DeviceID string `json:"device_id"`
	UID      string `json:"uid"`
}

type AccessResponsePayload struct {
	Status string `json:"status"`
	UID    string `json:"uid"`
}

type MQTTBrokerService struct {
	client           mqtt.Client
	cardService      *CardService
	accessLogService *AccessLogService
}

func NewMQTTBrokerService(cardService *CardService, accessLogService *AccessLogService) *MQTTBrokerService {
	return &MQTTBrokerService{
		cardService:      cardService,
		accessLogService: accessLogService,
	}
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}
}

func formatBrokerURL(broker string) string {
	broker = strings.TrimSpace(broker)
	if !strings.Contains(broker, "://") {
		broker = "tls://" + broker
	}
	// Adiciona a porta 8883 padrão do HiveMQ Cloud TLS caso não esteja informada
	parts := strings.Split(broker, "://")
	if len(parts) == 2 && !strings.Contains(parts[1], ":") {
		broker = broker + ":8883"
	}
	return broker
}

func (m *MQTTBrokerService) StartMQTT() {
	LoadEnv()

	user := os.Getenv("MQTT_USER")
	password := os.Getenv("MQTT_PASSWORD")
	clientID := os.Getenv("MQTT_CLIENT_ID")
	broker := os.Getenv("MQTT_BROKER")

	if user == "" || password == "" || clientID == "" || broker == "" {
		log.Fatal("Variáveis de ambiente MQTT faltando. Verifique seu arquivo .env")
	}

	brokerURL := formatBrokerURL(broker)

	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetUsername(user).
		SetPassword(password).
		SetAutoReconnect(true).
		SetTLSConfig(&tls.Config{InsecureSkipVerify: true})

	opts.OnConnect = func(c mqtt.Client) {
		log.Println("✅ Conectado ao HiveMQ Cloud!")
		if token := c.Subscribe(TopicAccessRequest, 1, m.onAccessRequest); token.Wait() && token.Error() != nil {
			log.Printf("❌ Erro ao subscrever ao tópico %s: %v", TopicAccessRequest, token.Error())
			return
		}
		log.Printf("📡 Inscrito no tópico: %s", TopicAccessRequest)
	}

	m.client = mqtt.NewClient(opts)
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Erro ao conectar no broker (%s): %v", brokerURL, token.Error())
	}
}

func (m *MQTTBrokerService) onAccessRequest(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("📨 [MQTT RX] Tópico: %s | Payload: %s\n", msg.Topic(), string(msg.Payload()))

	var req AccessRequestPayload
	if err := json.Unmarshal(msg.Payload(), &req); err != nil {
		log.Printf("⚠️ Erro ao decodificar JSON da solicitação: %v", err)
		return
	}

	uid := strings.ToUpper(strings.TrimSpace(req.UID))
	if uid == "" {
		log.Println("⚠️ Solicitação recebida sem UID válido")
		return
	}

	authorized := false
	status := "NEGADO"

	if m.cardService != nil {
		card, err := m.cardService.GetByUID(uid)
		if err != nil {
			log.Printf("🔒 Cartão UID %s não encontrado: %v", uid, err)
		} else if !card.Active {
			log.Printf("🔒 Cartão UID %s encontrado (ID=%d), porém está INATIVO", uid, card.ID)
			if m.accessLogService != nil {
				if _, logErr := m.accessLogService.Create(card.ID, false); logErr != nil {
					log.Printf("⚠️ Erro ao salvar log de acesso negado: %v", logErr)
				}
			}
		} else {
			authorized = true
			status = "LIBERADO"
			log.Printf("🔓 Acesso LIBERADO para o cartão UID %s (ID=%d, UserID=%d)", uid, card.ID, card.UserID)
			if m.accessLogService != nil {
				if _, logErr := m.accessLogService.Create(card.ID, true); logErr != nil {
					log.Printf("⚠️ Erro ao salvar log de acesso autorizado: %v", logErr)
				}
			}
		}
	} else {
		log.Println("⚠️ CardService não configurado no broker MQTT")
	}

	_ = authorized
	m.publishResponse(uid, status)
}

func (m *MQTTBrokerService) publishResponse(uid, status string) {
	resp := AccessResponsePayload{
		Status: status,
		UID:    uid,
	}

	payloadBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("❌ Erro ao serializar resposta MQTT: %v", err)
		return
	}

	token := m.client.Publish(TopicAccessResponse, 1, false, payloadBytes)
	token.Wait()
	if token.Error() != nil {
		log.Printf("❌ Erro ao publicar resposta em %s: %v", TopicAccessResponse, token.Error())
		return
	}

	log.Printf("📤 [MQTT TX] Resposta publicada em %s -> %s", TopicAccessResponse, string(payloadBytes))
}

