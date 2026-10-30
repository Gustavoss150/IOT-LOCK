package repoInterfaces

import "iot-lock/entities"

// AccessLogRepository expõe operações de criação e consulta para logs imutáveis.
type AccessLogRepository interface {
	Create(log *entities.AccessLog) error
	FindByID(id uint) (*entities.AccessLog, error)
	FindRecent(limit, offset int) ([]*entities.AccessLog, error)
	FindByCardID(cardID uint, limit, offset int) ([]*entities.AccessLog, error)
}