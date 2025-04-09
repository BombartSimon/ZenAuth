package models

import "time"

type AuthCode struct {
	Code                string
	ClientID            string
	RedirectURI         string
	UserID              string
	CodeChallenge       string
	CodeChallengeMethod string
	ExpiresAt           time.Time
	Scope               string
}
