package handlers

import (
	"encoding/json"
	"net/http"
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
