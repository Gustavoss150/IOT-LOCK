package repoInterfaces

import "iot-lock/entities"

type AccessLogRepository interface {
	Create(log *entities.AccessLog) error
	FindByID(id uint) (*entities.AccessLog, error)
	FindRecent(limit, offset int) ([]*entities.AccessLog, error)
	FindByCardID(cardID uint, limit, offset int) ([]*entities.AccessLog, error)
}
