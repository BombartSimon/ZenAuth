package repositories

import (
	"database/sql"
	"zenauth/internal/models"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func InitPostgres(connStr string) error {
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	return db.Ping()
}

func GetDB() *sql.DB {
	return db
}

func GetUserByUsername(username string) (*models.User, error) {
	row := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = $1", username)
	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func VerifyPassword(hashed string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}

func StoreAuthCode(code *models.AuthCode) error {
	_, err := db.Exec(`INSERT INTO auth_codes (code, client_id, redirect_uri, user_id, code_challenge, code_challenge_method, expires_at, scope)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		code.Code, code.ClientID, code.RedirectURI, code.UserID,
		code.CodeChallenge, code.CodeChallengeMethod, code.ExpiresAt, code.Scope) // Ajouter code.Scope
	return err
}

func GetAuthCode(code string) (*models.AuthCode, error) {
	row := db.QueryRow(`SELECT code, client_id, redirect_uri, user_id, code_challenge, code_challenge_method, expires_at, scope
		FROM auth_codes WHERE code = $1`, code)

	var ac models.AuthCode
	err := row.Scan(&ac.Code, &ac.ClientID, &ac.RedirectURI, &ac.UserID, &ac.CodeChallenge, &ac.CodeChallengeMethod, &ac.ExpiresAt, &ac.Scope) // Ajouter ac.Scope
	if err != nil {
		return nil, err
	}
	return &ac, nil
}

func DeleteAuthCode(code string) error {
	_, err := db.Exec(`DELETE FROM auth_codes WHERE code = $1`, code)
	return err
}

func GetClientByID(id string) (*models.Client, error) {
	row := db.QueryRow(`SELECT id, secret, name, redirect_uris FROM clients WHERE id = $1`, id)

	var c models.Client
	err := row.Scan(&c.ID, &c.Secret, &c.Name, pq.Array(&c.RedirectURIs))
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func StoreRefreshToken(token string, clientID string, userID *string) error {
	_, err := db.Exec(`
		INSERT INTO refresh_tokens (token, client_id, user_id)
		VALUES ($1, $2, $3)`, token, clientID, userID)
	return err
}

func GetRefreshToken(token string) (string, *string, error) {
	row := db.QueryRow(`SELECT client_id, user_id FROM refresh_tokens WHERE token = $1`, token)
	var clientID string
	var userID *string
	err := row.Scan(&clientID, &userID)
	if err != nil {
		return "", nil, err
	}
	return clientID, userID, nil
}
