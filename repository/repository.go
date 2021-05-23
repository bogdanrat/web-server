package repository

import "github.com/bogdanrat/web-server/models"

type DatabaseRepository interface {
	GetAllUsers() ([]*models.User, error)
	GetUserByEmail(string) (*models.User, error)
	InsertUser(user *models.User) error
	UpdateUserQRSecret(email, secret string) error
}
