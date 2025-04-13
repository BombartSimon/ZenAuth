package adapters

import (
	"database/sql"
	"fmt"
	"zenauth/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type SQLUser struct {
	db                *sql.DB
	tableName         string
	idField           string
	usernameField     string
	passwordHashField string
	emailField        string
}

func NewSQLUser(db *sql.DB, tableName, idField, usernameField, passwordHashField, emailField string) *SQLUser {
	return &SQLUser{
		db:                db,
		tableName:         tableName,
		idField:           idField,
		usernameField:     usernameField,
		passwordHashField: passwordHashField,
		emailField:        emailField,
	}
}

func (p *SQLUser) GetUserByUsername(username string) (*models.User, error) {
	query := fmt.Sprintf(
		"SELECT %s, %s, %s, %s FROM %s WHERE %s = $1",
		p.idField, p.usernameField, p.passwordHashField, p.emailField,
		p.tableName, p.usernameField,
	)

	row := p.db.QueryRow(query, username)

	var user models.User
	var email sql.NullString

	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &email)
	if err != nil {
		return nil, err
	}

	if email.Valid {
		user.Email = email.String
	}

	return &user, nil
}

func (p *SQLUser) GetUserByEmail(email string) (*models.User, error) {
	query := fmt.Sprintf(
		"SELECT %s, %s, %s, %s FROM %s WHERE %s = $1",
		p.idField, p.usernameField, p.passwordHashField, p.emailField,
		p.tableName, p.emailField,
	)
	row := p.db.QueryRow(query, email)
	var user models.User
	var passwordHash sql.NullString
	err := row.Scan(&user.ID, &user.Username, &passwordHash, &user.Email)
	if err != nil {
		return nil, err
	}
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.String
	} else {
		user.PasswordHash = ""
	}
	return &user, nil
}

func (p *SQLUser) VerifyPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

func (p *SQLUser) GetAllUsers() ([]models.User, error) {
	query := fmt.Sprintf(
		"SELECT %s, %s, %s, %s FROM %s",
		p.idField, p.usernameField, p.passwordHashField, p.emailField,
		p.tableName,
	)

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var email sql.NullString
		var passwordHash sql.NullString

		if err := rows.Scan(&user.ID, &user.Username, &passwordHash, &email); err != nil {
			return nil, err
		}

		if email.Valid {
			user.Email = email.String
		}

		if passwordHash.Valid {
			user.PasswordHash = passwordHash.String
		} else {
			user.PasswordHash = ""
		}

		users = append(users, user)
	}

	return users, nil
}
