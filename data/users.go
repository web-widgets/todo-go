package data

import (
	"gorm.io/gorm"
)

func NewUsersDAO(db *gorm.DB) *UsersDAO {
	return &UsersDAO{db}
}

type UsersDAO struct {
	db *gorm.DB
}

func (m *UsersDAO) GetAll() ([]User, error) {
	users := make([]User, 0)
	err := m.db.Find(&users).Error
	return users, err
}
