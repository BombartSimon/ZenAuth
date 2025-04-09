package models

type Token struct {
	AccessToken  string
	RefreshToken string
	Expiry       int64
	ClientID     string
	UserID       string
}

type RefreshToken struct {
	Token    string
	ClientID string
	UserID   *string
	IssuedAt string
}
