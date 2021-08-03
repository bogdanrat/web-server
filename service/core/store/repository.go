package store

import "github.com/bogdanrat/web-server/contracts/models"

type DatabaseRepository interface {
	GetAllUsers() ([]*models.User, error)
	GetUserByEmail(string) (*models.User, error)
	InsertUser(user *models.User) error
	UpdateUserQRSecret(email, secret string) error
}
