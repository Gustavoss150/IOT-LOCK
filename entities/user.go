package entities

type User struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
	CPF  int64  `json:"cpf"`
}
