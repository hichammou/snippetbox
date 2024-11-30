package models

import "errors"

var (
	ErrNoRecord = errors.New("models: no matching record found")

	ErrInvalideCredentials = errors.New("models: invalide credentials")

	ErrDuplicateEmail = errors.New("models: duplicate email")
)
