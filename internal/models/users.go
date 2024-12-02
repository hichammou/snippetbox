package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	Get(id int) (User, error)
	UpdatePassword(id int, oldPassword, newPassword string) error
}

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created) VALUES (?, ?, ?, UTC_TIMESTAMP())`

	_, err = m.DB.Exec(stmt, name, email, string(hashedPassword))

	if err != nil {
		// If this returns an error, we use the errors.As() function to check whether the error has the type *mysql.MySQLError. If it does, the error will
		// be assigned to the mySQLError variable. We can then check whether or not the error relates to our user_uc_email (unique constraint on email column) key by checking if the error code
		// equals to 1062 and the content of the error message string. If it does we return an ErrDuplicatedEmail error.
		var mySQLError *mysql.MySQLError

		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}

		return err
	}

	return nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool

	stmt := `SELECT EXISTS (SELECT true FROM users WHERE id = ?)`
	err := m.DB.QueryRow(stmt, id).Scan(&exists)

	return exists, err
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := `SELECT id, hashed_password FROM users WHERE email = ?`

	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalideCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalideCredentials
		} else {
			return 0, err
		}
	}

	return id, nil
}

func (m *UserModel) Get(id int) (User, error) {
	var u User
	stmt := `SELECT id, name, email, created FROM users WHERE id = ?`

	err := m.DB.QueryRow(stmt, id).Scan(&u.ID, &u.Name, &u.Email, &u.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return u, ErrNoRecord
		}
		return u, err
	}

	return u, nil
}

func (m *UserModel) UpdatePassword(id int, oldPassword, newPassword string) error {
	var password string

	stmt := `SELECT hashed_password FROM users WHERE id = ?`

	err := m.DB.QueryRow(stmt, id).Scan(&password)

	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(oldPassword))
	if err != nil {
		return ErrInvalideCredentials
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)

	if err != nil {
		return err
	}

	stmt = `UPDATE users SET hashed_password = ? WHERE id = ?`

	_, err = m.DB.Exec(stmt, hashed, id)

	if err != nil {
		return err
	}

	// Means no error
	return nil
}
