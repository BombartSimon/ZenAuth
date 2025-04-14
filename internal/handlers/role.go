package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	rProviders "zenauth/internal/adapters/role"

	"github.com/gorilla/mux"
)

// AdminRolesHandler handles requests to the /admin/roles endpoint
func AdminRolesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listRoles(w, r)
	case http.MethodPost:
		createRole(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminRoleHandler handles requests to the /admin/roles/{id} endpoint
func AdminRoleHandler(w http.ResponseWriter, r *http.Request) {
	// Extract role ID from mux vars
	vars := mux.Vars(r)
	id := vars["id"]

	switch r.Method {
	case http.MethodGet:
		getRole(w, r, id)
	case http.MethodPut:
		updateRole(w, r, id)
	case http.MethodDelete:
		deleteRole(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminGroupsHandler handles requests to the /admin/groups endpoint
func AdminGroupsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listGroups(w, r)
	case http.MethodPost:
		createGroup(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminGroupHandler handles requests to the /admin/groups/{id} endpoint
func AdminGroupHandler(w http.ResponseWriter, r *http.Request) {
	// Extract group ID from mux vars
	vars := mux.Vars(r)
	id := vars["id"]

	switch r.Method {
	case http.MethodGet:
		getGroup(w, r, id)
	case http.MethodPut:
		updateGroup(w, r, id)
	case http.MethodDelete:
		deleteGroup(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminUserRolesHandler handles requests to the /admin/users-roles endpoint
func AdminUserRolesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getUserRoles(w, r)
	case http.MethodPost:
		assignRoleToUser(w, r)
	case http.MethodDelete:
		removeRoleFromUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// AdminUserGroupsHandler handles requests to the /admin/users-groups endpoint
func AdminUserGroupsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getUserGroups(w, r)
	case http.MethodPost:
		assignUserToGroup(w, r)
	case http.MethodDelete:
		removeUserFromGroup(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Role Management Functions

func listRoles(w http.ResponseWriter, r *http.Request) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	roles, err := rProviders.CurrentManager.GetAllRoles(ctx)
	if err != nil {
		http.Error(w, "Failed to retrieve roles: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

func createRole(w http.ResponseWriter, r *http.Request) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	var data struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.Name == "" {
		http.Error(w, "Role name is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	newRole, err := rProviders.CurrentManager.CreateRole(ctx, data.Name, data.Description)
	if err != nil {
		http.Error(w, "Failed to create role: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newRole)
}

func getRole(w http.ResponseWriter, r *http.Request, id string) {
	// This would be implemented to get a specific role by ID
	// Since it's not a core part of your immediate needs, I'll leave this as a stub
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func updateRole(w http.ResponseWriter, r *http.Request, id string) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	var data struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.Name == "" {
		http.Error(w, "Role name is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err := rProviders.CurrentManager.UpdateRole(ctx, id, data.Name, data.Description)
	if err != nil {
		http.Error(w, "Failed to update role: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Role updated successfully",
	})
}

func deleteRole(w http.ResponseWriter, r *http.Request, id string) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	err := rProviders.CurrentManager.DeleteRole(ctx, id)
	if err != nil {
		http.Error(w, "Failed to delete role: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Group Management Functions

func listGroups(w http.ResponseWriter, r *http.Request) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	groups, err := rProviders.CurrentManager.GetAllGroups(ctx)
	if err != nil {
		http.Error(w, "Failed to retrieve groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

func createGroup(w http.ResponseWriter, r *http.Request) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	var data struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.Name == "" {
		http.Error(w, "Group name is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	newGroup, err := rProviders.CurrentManager.CreateGroup(ctx, data.Name, data.Description)
	if err != nil {
		http.Error(w, "Failed to create group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newGroup)
}

func getGroup(w http.ResponseWriter, r *http.Request, id string) {
	// This would be implemented to get a specific group by ID
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func updateGroup(w http.ResponseWriter, r *http.Request, id string) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	var data struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.Name == "" {
		http.Error(w, "Group name is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err := rProviders.CurrentManager.UpdateGroup(ctx, id, data.Name, data.Description)
	if err != nil {
		http.Error(w, "Failed to update group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Group updated successfully",
	})
}

func deleteGroup(w http.ResponseWriter, r *http.Request, id string) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	err := rProviders.CurrentManager.DeleteGroup(ctx, id)
	if err != nil {
		http.Error(w, "Failed to delete group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// User-Role Management Functions

func getUserRoles(w http.ResponseWriter, r *http.Request) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id parameter is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	roles, err := rProviders.CurrentManager.GetUserRoles(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to get user roles: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

func assignRoleToUser(w http.ResponseWriter, r *http.Request) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	var data struct {
		UserID string `json:"user_id"`
		RoleID string `json:"role_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.UserID == "" || data.RoleID == "" {
		http.Error(w, "Both user_id and role_id are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err := rProviders.CurrentManager.AssignRoleToUser(ctx, data.UserID, data.RoleID)
	if err != nil {
		http.Error(w, "Failed to assign role to user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func removeRoleFromUser(w http.ResponseWriter, r *http.Request) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	userID := r.URL.Query().Get("user_id")
	roleID := r.URL.Query().Get("role_id")

	if userID == "" || roleID == "" {
		http.Error(w, "Both user_id and role_id parameters are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err := rProviders.CurrentManager.RemoveRoleFromUser(ctx, userID, roleID)
	if err != nil {
		http.Error(w, "Failed to remove role from user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// User-Group Management Functions

func getUserGroups(w http.ResponseWriter, r *http.Request) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id parameter is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	groups, err := rProviders.CurrentManager.GetUserGroups(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to get user groups: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

func assignUserToGroup(w http.ResponseWriter, r *http.Request) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	var data struct {
		UserID  string `json:"user_id"`
		GroupID string `json:"group_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if data.UserID == "" || data.GroupID == "" {
		http.Error(w, "Both user_id and group_id are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err := rProviders.CurrentManager.AssignUserToGroup(ctx, data.UserID, data.GroupID)
	if err != nil {
		http.Error(w, "Failed to assign user to group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func removeUserFromGroup(w http.ResponseWriter, r *http.Request) {
	if rProviders.CurrentManager == nil {
		http.Error(w, "Role manager not initialized", http.StatusInternalServerError)
		return
	}

	userID := r.URL.Query().Get("user_id")
	groupID := r.URL.Query().Get("group_id")

	if userID == "" || groupID == "" {
		http.Error(w, "Both user_id and group_id parameters are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err := rProviders.CurrentManager.RemoveUserFromGroup(ctx, userID, groupID)
	if err != nil {
		http.Error(w, "Failed to remove user from group: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
