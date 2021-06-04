package db

import "github.com/bogdanrat/web-server/contracts/models"

type UsersDatabase interface {
	GetAllUsers() ([]*models.User, error)
	GetUserByEmail(string) (*models.User, error)
	InsertUser(user *models.User) error
	UpdateUserQRSecret(email, secret string) error
}
