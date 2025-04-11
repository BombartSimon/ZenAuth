package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	sProviders "zenauth/internal/adapters/sessions"
	uProviders "zenauth/internal/adapters/users"
	"zenauth/internal/models"
	"zenauth/internal/repositories"
)

// AdminUsersHandler handles requests to the /admin/users endpoint
func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listUsers(w, r)
	case http.MethodPost:
		createUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminUserHandler handles requests to the /admin/users/{id} endpoint
func AdminUserHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from path
	id := r.URL.Path[len("/admin/users/"):]

	switch r.Method {
	case http.MethodGet:
		getUser(w, r, id)
	case http.MethodPut:
		updateUser(w, r, id)
	case http.MethodDelete:
		deleteUser(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminBlockedUsersHandler handles requests to the /admin/blocked-users endpoint
func AdminBlockedUsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listBlockedUsers(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listBlockedUsers returns all currently blocked users
func listBlockedUsers(w http.ResponseWriter, r *http.Request) {
	if !sProviders.IsLimiterEnabled() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	blockedUsers, err := sProviders.GetBlockedIdentifiers()
	if err != nil {
		http.Error(w, "Failed to retrieve blocked users", http.StatusInternalServerError)
		return
	}

	type BlockedUser struct {
		Identifier string `json:"identifier"`
		Type       string `json:"type"`
		BlockedFor string `json:"blocked_for"`
	}

	result := make([]BlockedUser, 0)

	// Convertir les identifiants en objets structurés
	for _, id := range blockedUsers {
		var userType string
		var identifier string

		if strings.HasPrefix(id, "user:") {
			userType = "user"
			identifier = strings.TrimPrefix(id, "user:")
		} else {
			userType = "ip"
			identifier = id
		}

		// Récupérer le temps restant de blocage
		remaining, _ := sProviders.GetRemainingBlockTime(id)

		result = append(result, BlockedUser{
			Identifier: identifier,
			Type:       userType,
			BlockedFor: remaining,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// AdminUnblockUserHandler handles requests to unblock a user or IP
func AdminUnblockUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Identifier string `json:"identifier"`
		Type       string `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var fullIdentifier string
	if data.Type == "user" {
		fullIdentifier = "user:" + data.Identifier
	} else {
		fullIdentifier = data.Identifier
	}

	if err := sProviders.ResetLoginAttempts(fullIdentifier); err != nil {
		http.Error(w, "Failed to unblock user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully unblocked " + data.Type + " " + data.Identifier,
	})
}

// AdminClientsHandler handles requests to the /admin/clients endpoint
func AdminClientsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listClients(w, r)
	case http.MethodPost:
		createClient(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminClientHandler handles requests to the /admin/clients/{id} endpoint
func AdminClientHandler(w http.ResponseWriter, r *http.Request) {
	// Extract client ID from path
	id := r.URL.Path[len("/admin/clients/"):]

	switch r.Method {
	case http.MethodGet:
		getClient(w, r, id)
	case http.MethodPut:
		updateClient(w, r, id)
	case http.MethodDelete:
		deleteClient(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// User Management
func listUsers(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	useProvider := false

	// check if provider have external
	if query.Get("provider") == "external" {
		useProvider = true
	}

	var users []models.User
	var err error

	if useProvider {
		// Utiliser le provider externe
		users, err = uProviders.CurrentUserProvider.GetAllUsers()
	} else {
		// Utiliser la base de données locale par défaut
		users, err = repositories.GetAllUsers()
	}

	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	// Don't return password hashes in the response
	type SafeUser struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email,omitempty"`
		Source   string `json:"source"`
	}

	safeUsers := make([]SafeUser, len(users))
	for i, user := range users {
		source := "local"
		if useProvider {
			source = "external"
		}
		safeUsers[i] = SafeUser{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Source:   source,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(safeUsers)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.Username == "" || data.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	user, err := repositories.CreateUser(data.Username, data.Password)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":       user.ID,
		"username": user.Username,
	})
}

func getUser(w http.ResponseWriter, r *http.Request, id string) {
	user, err := repositories.GetUserByID(id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":       user.ID,
		"username": user.Username,
	})
}

func updateUser(w http.ResponseWriter, r *http.Request, id string) {
	var data struct {
		Username string  `json:"username"`
		Password *string `json:"password,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if err := repositories.UpdateUser(id, data.Username, data.Password); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User updated successfully",
	})
}

func deleteUser(w http.ResponseWriter, r *http.Request, id string) {
	if err := repositories.DeleteUser(id); err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Client Management
func listClients(w http.ResponseWriter, r *http.Request) {
	clients, err := repositories.GetAllClients()
	if err != nil {
		http.Error(w, "Failed to retrieve clients", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

func createClient(w http.ResponseWriter, r *http.Request) {
	var data struct {
		ID           string   `json:"id"`
		Name         string   `json:"name"`
		Secret       string   `json:"secret"`
		RedirectURIs []string `json:"redirect_uris"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.ID == "" || data.Name == "" || data.Secret == "" {
		http.Error(w, "Client ID, name, and secret are required", http.StatusBadRequest)
		return
	}

	client, err := repositories.CreateClient(data.ID, data.Secret, data.Name, data.RedirectURIs)
	if err != nil {
		http.Error(w, "Failed to create client", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(client)
}

func getClient(w http.ResponseWriter, r *http.Request, id string) {
	client, err := repositories.GetClientByID(id)
	if err != nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}

func updateClient(w http.ResponseWriter, r *http.Request, id string) {
	var data struct {
		Name         string   `json:"name"`
		Secret       *string  `json:"secret,omitempty"`
		RedirectURIs []string `json:"redirect_uris"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.Name == "" {
		http.Error(w, "Client name is required", http.StatusBadRequest)
		return
	}

	if err := repositories.UpdateClient(id, data.Name, data.Secret, data.RedirectURIs); err != nil {
		http.Error(w, "Failed to update client", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Client updated successfully",
	})
}

func deleteClient(w http.ResponseWriter, r *http.Request, id string) {
	if err := repositories.DeleteClient(id); err != nil {
		http.Error(w, "Failed to delete client", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListAuthProviders returns all configured authentication providers
func ListAuthProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := repositories.GetAllAuthProviders()
	if err != nil {
		http.Error(w, "Failed to retrieve authentication providers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(providers)
}

func CreateAuthProvider(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Name         string `json:"name"`
		Type         string `json:"type"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		TenantID     string `json:"tenant_id"` // Added tenant_id
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.Name == "" || data.Type == "" || data.ClientID == "" || data.ClientSecret == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Validate provider type
	providerType := models.AuthProviderType(data.Type)
	if providerType != models.GoogleProvider &&
		providerType != models.MicrosoftProvider &&
		providerType != models.GitHubProvider {
		http.Error(w, "Invalid provider type", http.StatusBadRequest)
		return
	}

	provider, err := repositories.CreateAuthProvider(data.Name, providerType, data.ClientID, data.ClientSecret, data.TenantID)
	if err != nil {
		http.Error(w, "Failed to create authentication provider", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(provider)
}

// GetAuthProvider retrieves an authentication provider by ID
func GetAuthProvider(w http.ResponseWriter, r *http.Request) {
	// Extract ID from the path
	path := strings.TrimPrefix(r.URL.Path, "/admin/auth-providers/")
	id := strings.TrimSuffix(path, "/")

	provider, err := repositories.GetAuthProviderByID(id)
	if err != nil {
		http.Error(w, "Authentication provider not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(provider)
}

// UpdateAuthProvider updates an existing authentication provider
func UpdateAuthProvider(w http.ResponseWriter, r *http.Request) {
	// Extract ID from the path
	path := strings.TrimPrefix(r.URL.Path, "/admin/auth-providers/")
	id := strings.TrimSuffix(path, "/")

	var data struct {
		Name         string `json:"name"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"` // Optional for updates
		TenantID     string `json:"tenant_id"`     // Added tenant_id
		Enabled      bool   `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.Name == "" || data.ClientID == "" {
		http.Error(w, "Name and client_id are required", http.StatusBadRequest)
		return
	}

	err := repositories.UpdateAuthProvider(id, data.Name, data.ClientID, data.ClientSecret, data.TenantID, data.Enabled)
	if err != nil {
		http.Error(w, "Failed to update authentication provider", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteAuthProvider removes an authentication provider
func DeleteAuthProvider(w http.ResponseWriter, r *http.Request) {
	// Extract ID from the path
	path := strings.TrimPrefix(r.URL.Path, "/admin/auth-providers/")
	id := strings.TrimSuffix(path, "/")

	err := repositories.DeleteAuthProvider(id)
	if err != nil {
		http.Error(w, "Failed to delete authentication provider", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
