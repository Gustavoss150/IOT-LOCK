package services

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"iot-lock/entities"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

const (
	// Tópicos de Controle de Acesso da Porta
	TopicAccessRequest  = "smartlock/acesso/solicitacao" // ESP -> Backend
	TopicAccessResponse = "smartlock/acesso/resposta"    // Backend -> ESP

	// Tópicos de Cadastro de Cartão
	TopicEnrollmentCommand  = "smartlock/cadastro/comando"     // Backend -> ESP (Ativa/Cancela modo cadastro)
	TopicEnrollmentRequest  = "smartlock/cadastro/solicitacao" // ESP -> Backend (Envia UID lido para cadastro)
	TopicEnrollmentResponse = "smartlock/cadastro/resposta"    // Backend -> ESP (Confirma leitura do cadastro)
)

type AccessRequestPayload struct {
	DeviceID string `json:"device_id"`
	UID      string `json:"uid"`
}

type AccessResponsePayload struct {
	Status string `json:"status"`
	UID    string `json:"uid"`
}

type EnrollmentStatus struct {
	Active       bool           `json:"active"`
	UserID       uint           `json:"user_id"`
	Card         *entities.Card `json:"card,omitempty"`
	ConflictCard *entities.Card `json:"conflict_card,omitempty"`
	Error        string         `json:"error,omitempty"`
}

type MQTTBrokerService struct {
	client           mqtt.Client
	cardService      *CardService
	accessLogService *AccessLogService

	mu                  sync.Mutex
	enrollmentActive    bool
	enrollmentUserID    uint
	enrollmentExpiresAt time.Time
	lastEnrolledCard    *entities.Card
	conflictCard        *entities.Card
	lastEnrollmentErr   string
}

func NewMQTTBrokerService(cardService *CardService, accessLogService *AccessLogService) *MQTTBrokerService {
	return &MQTTBrokerService{
		cardService:      cardService,
		accessLogService: accessLogService,
	}
}

func (m *MQTTBrokerService) StartEnrollment(userID uint) {
	m.mu.Lock()
	m.enrollmentActive = true
	m.enrollmentUserID = userID
	m.enrollmentExpiresAt = time.Now().Add(60 * time.Second)
	m.lastEnrolledCard = nil
	m.conflictCard = nil
	m.lastEnrollmentErr = ""
	m.mu.Unlock()

	log.Printf("🪪 [MODO CADASTRO] Aguardando leitura de cartão no tópico %s para UserID=%d (60s)...", TopicEnrollmentRequest, userID)
	m.publishToTopic(TopicEnrollmentCommand, []byte(`{"modo":"ATIVAR"}`))
}

func (m *MQTTBrokerService) CancelEnrollment() {
	m.mu.Lock()
	m.enrollmentActive = false
	m.enrollmentUserID = 0
	m.conflictCard = nil
	m.mu.Unlock()

	log.Println("🪪 [MODO CADASTRO] Cancelado.")
	m.publishToTopic(TopicEnrollmentCommand, []byte(`{"modo":"CANCELAR"}`))
}

func (m *MQTTBrokerService) GetEnrollmentStatus() EnrollmentStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.enrollmentActive && time.Now().After(m.enrollmentExpiresAt) {
		m.enrollmentActive = false
		m.lastEnrollmentErr = "Tempo limite para aproximar o cartão expirou."
	}

	return EnrollmentStatus{
		Active:       m.enrollmentActive,
		UserID:       m.enrollmentUserID,
		Card:         m.lastEnrolledCard,
		ConflictCard: m.conflictCard,
		Error:        m.lastEnrollmentErr,
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
		log.Printf("📡 Inscrito no tópico de acesso: %s", TopicAccessRequest)

		if token := c.Subscribe(TopicEnrollmentRequest, 1, m.onEnrollmentRequest); token.Wait() && token.Error() != nil {
			log.Printf("❌ Erro ao subscrever ao tópico %s: %v", TopicEnrollmentRequest, token.Error())
			return
		}
		log.Printf("📡 Inscrito no tópico de cadastro: %s", TopicEnrollmentRequest)
	}

	m.client = mqtt.NewClient(opts)
	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Erro ao conectar no broker (%s): %v", brokerURL, token.Error())
	}
}

func (m *MQTTBrokerService) onEnrollmentRequest(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("📨 [MQTT RX - CADASTRO] Tópico: %s | Payload: %s\n", msg.Topic(), string(msg.Payload()))

	var req AccessRequestPayload
	if err := json.Unmarshal(msg.Payload(), &req); err != nil {
		log.Printf("⚠️ Erro ao decodificar JSON de cadastro: %v", err)
		return
	}

	uid := strings.ToUpper(strings.TrimSpace(req.UID))
	if uid == "" {
		log.Println("⚠️ Solicitação de cadastro recebida sem UID válido")
		return
	}

	m.processEnrollmentUID(uid, TopicEnrollmentResponse)
}

func (m *MQTTBrokerService) processEnrollmentUID(uid string, responseTopic string) {
	m.mu.Lock()
	isEnrolling := m.enrollmentActive && time.Now().Before(m.enrollmentExpiresAt)
	enrollUserID := m.enrollmentUserID
	if isEnrolling {
		m.enrollmentActive = false
	}
	m.mu.Unlock()

	if !isEnrolling || m.cardService == nil {
		log.Printf("⚠️ Leitura de cadastro ignorada (modo cadastro inativo ou expirado para UID %s)", uid)
		m.publishResponseToTopic(responseTopic, uid, "EXPIRADO")
		return
	}

	log.Printf("🪪 [MODO CADASTRO] Cartão UID %s recebido! Verificando vínculo para UserID=%d...", uid, enrollUserID)
	existingCard, err := m.cardService.GetByUID(uid)
	if err == nil && existingCard != nil {
		if existingCard.UserID != enrollUserID {
			m.mu.Lock()
			m.conflictCard = existingCard
			m.lastEnrolledCard = nil
			m.lastEnrollmentErr = ""
			m.mu.Unlock()
			log.Printf("⚠️ [MODO CADASTRO] Cartão UID %s já pertence ao UserID=%d. Aguardando confirmação para transferir para UserID=%d.", uid, existingCard.UserID, enrollUserID)
			m.publishResponseToTopic(responseTopic, uid, "CONFLITO")
			return
		}

		existingCard.Active = true
		card, updateErr := m.cardService.Update(existingCard)
		m.mu.Lock()
		if updateErr != nil {
			m.lastEnrollmentErr = updateErr.Error()
		} else {
			m.lastEnrolledCard = card
			m.lastEnrollmentErr = ""
		}
		m.mu.Unlock()
		m.publishResponseToTopic(responseTopic, uid, "CADASTRADO")
		return
	}

	card, createErr := m.cardService.Create(uid, enrollUserID)
	m.mu.Lock()
	if createErr != nil {
		m.lastEnrollmentErr = createErr.Error()
		log.Printf("❌ [MODO CADASTRO] Erro ao cadastrar cartão %s: %v", uid, createErr)
	} else {
		m.lastEnrolledCard = card
		m.lastEnrollmentErr = ""
		log.Printf("✅ [MODO CADASTRO] Cartão UID %s cadastrado com sucesso para UserID=%d!", uid, enrollUserID)
	}
	m.mu.Unlock()

	if createErr == nil {
		m.publishResponseToTopic(responseTopic, uid, "CADASTRADO")
	} else {
		m.publishResponseToTopic(responseTopic, uid, "ERRO")
	}
}

func (m *MQTTBrokerService) onAccessRequest(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("📨 [MQTT RX - ACESSO] Tópico: %s | Payload: %s\n", msg.Topic(), string(msg.Payload()))

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

	// Fallback: caso a ESP8266 ainda esteja com o firmware antigo enviando no tópico de acesso durante o modo cadastro
	m.mu.Lock()
	isEnrolling := m.enrollmentActive && time.Now().Before(m.enrollmentExpiresAt)
	m.mu.Unlock()
	if isEnrolling {
		m.processEnrollmentUID(uid, TopicAccessResponse)
		return
	}

	authorized := false
	status := "NEGADO"

	if m.cardService != nil {
		card, err := m.cardService.GetByUID(uid)
		if err != nil {
			log.Printf("🔒 Cartão UID %s não encontrado: %v", uid, err)
			if m.accessLogService != nil {
				if _, logErr := m.accessLogService.Create(nil, uid, false); logErr != nil {
					log.Printf("⚠️ Erro ao salvar log de cartão não cadastrado: %v", logErr)
				}
			}
		} else if !card.Active {
			log.Printf("🔒 Cartão UID %s encontrado (ID=%d), porém está INATIVO", uid, card.ID)
			if m.accessLogService != nil {
				if _, logErr := m.accessLogService.Create(&card.ID, uid, false); logErr != nil {
					log.Printf("⚠️ Erro ao salvar log de acesso negado: %v", logErr)
				}
			}
		} else {
			authorized = true
			status = "LIBERADO"
			log.Printf("🔓 Acesso LIBERADO para o cartão UID %s (ID=%d, UserID=%d)", uid, card.ID, card.UserID)
			if m.accessLogService != nil {
				if _, logErr := m.accessLogService.Create(&card.ID, uid, true); logErr != nil {
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
	m.publishResponseToTopic(TopicAccessResponse, uid, status)
}

func (m *MQTTBrokerService) publishResponseToTopic(topic, uid, status string) {
	resp := AccessResponsePayload{
		Status: status,
		UID:    uid,
	}

	payloadBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("❌ Erro ao serializar resposta MQTT: %v", err)
		return
	}

	m.publishToTopic(topic, payloadBytes)
}

func (m *MQTTBrokerService) publishToTopic(topic string, payloadBytes []byte) {
	if m.client == nil || !m.client.IsConnected() {
		log.Printf("⚠️ Cliente MQTT não conectado ao tentar publicar em %s", topic)
		return
	}

	token := m.client.Publish(topic, 1, false, payloadBytes)
	token.Wait()
	if token.Error() != nil {
		log.Printf("❌ Erro ao publicar em %s: %v", topic, token.Error())
		return
	}

	log.Printf("📤 [MQTT TX] Publicado em %s -> %s", topic, string(payloadBytes))
}

