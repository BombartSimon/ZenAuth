package handlers

import (
	"net/http"
	"zenauth/oauth"
)

var flows []oauth.OAuthFlow

func RegisterFlows(f []oauth.OAuthFlow) {
	flows = f
}

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")
	for _, flow := range flows {
		if flow.Supports(grantType) {
			flow.HandleTokenRequest(w, r)
			return
		}
	}
	http.Error(w, "unsupported_grant_type", http.StatusBadRequest)
}
