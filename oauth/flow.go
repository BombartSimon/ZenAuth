package oauth

import "net/http"

type OAuthFlow interface {
	Supports(grantType string) bool
	HandleTokenRequest(w http.ResponseWriter, r *http.Request)
}
