package models

import "golang.org/x/crypto/bcrypt"

type User struct {
	ID       int     `json:"-"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Password string  `json:"-"`
	QRSecret *string `json:"-"`
}

func (user *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(bytes)

	return nil
}

func (user *User) CheckPassword(password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return err
	}
	return nil
}
