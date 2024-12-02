package mocks

import (
	"time"

	"snippetbox.hichammou/internal/models"
)

type UserModel struct{}

func (m *UserModel) Insert(name, email, password string) error {
	switch email {
	case "hicham@gmail.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	if email == "hicham@gmail.com" && password == "1234" {
		return 1, nil
	}
	return 0, models.ErrInvalideCredentials
}

func (m *UserModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (m *UserModel) Get(id int) (models.User, error) {
	if id == 1 {
		u := models.User{
			ID:      1,
			Name:    "Hicham",
			Email:   "hicham@example.com",
			Created: time.Now(),
		}

		return u, nil
	}

	return models.User{}, models.ErrNoRecord
}
