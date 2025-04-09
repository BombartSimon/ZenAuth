package adapters

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// LocalManager implémente Manager pour stocker les rôles dans la base ZenAuth
type LocalManager struct {
	db *sql.DB
}

// NewLocalManager crée un nouveau gestionnaire local de rôles
func NewLocalManager(db *sql.DB) (*LocalManager, error) {
	manager := &LocalManager{db: db}
	if err := manager.initTables(); err != nil {
		return nil, err
	}
	return manager, nil
}

// initTables initialise les tables nécessaires si elles n'existent pas
func (m *LocalManager) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS roles (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE TABLE IF NOT EXISTS groups (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        )`,
		`CREATE TABLE IF NOT EXISTS group_roles (
            group_id TEXT NOT NULL,
            role_id TEXT NOT NULL,
            PRIMARY KEY (group_id, role_id),
            FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE,
            FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE
        )`,
		`CREATE TABLE IF NOT EXISTS user_roles (
            user_id TEXT NOT NULL,
            role_id TEXT NOT NULL,
            PRIMARY KEY (user_id, role_id),
            FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE
        )`,
		`CREATE TABLE IF NOT EXISTS user_groups (
            user_id TEXT NOT NULL,
            group_id TEXT NOT NULL,
            PRIMARY KEY (user_id, group_id),
            FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE
        )`,
		`CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_groups_user_id ON user_groups (user_id)`,
	}

	for _, query := range queries {
		if _, err := m.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

// GetAllRoles récupère tous les rôles disponibles
func (m *LocalManager) GetAllRoles(ctx context.Context) ([]Role, error) {
	query := `SELECT id, name, description FROM roles`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetUserRoles récupère tous les rôles d'un utilisateur (directs + via groupes)
func (m *LocalManager) GetUserRoles(ctx context.Context, userID string) ([]Role, error) {
	// Récupérer les rôles directs
	directRoles, err := m.GetUserDirectRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Récupérer les rôles via groupes
	query := `
        SELECT DISTINCT r.id, r.name, r.description
        FROM roles r
        JOIN group_roles gr ON r.id = gr.role_id
        JOIN user_groups ug ON gr.group_id = ug.group_id
        WHERE ug.user_id = $1
    `

	rows, err := m.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map pour éliminer les doublons
	roleMap := make(map[string]Role)

	// Ajouter d'abord les rôles directs
	for _, role := range directRoles {
		roleMap[role.ID] = role
	}

	// Ajouter les rôles via groupes
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roleMap[role.ID] = role
	}

	// Convertir la map en slice
	result := make([]Role, 0, len(roleMap))
	for _, role := range roleMap {
		result = append(result, role)
	}

	return result, nil
}

// HasRole vérifie si un utilisateur a un rôle spécifique
func (m *LocalManager) HasRole(ctx context.Context, userID string, roleID string) (bool, error) {
	// Vérifier rôle direct
	query1 := `SELECT 1 FROM user_roles WHERE user_id = $1 AND role_id = $2 LIMIT 1`
	row := m.db.QueryRowContext(ctx, query1, userID, roleID)
	var exists int
	err := row.Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if err == nil {
		return true, nil
	}

	// Vérifier rôle via groupe
	query2 := `
        SELECT 1 FROM group_roles gr
        JOIN user_groups ug ON gr.group_id = ug.group_id
        WHERE ug.user_id = $1 AND gr.role_id = $2
        LIMIT 1
    `
	row = m.db.QueryRowContext(ctx, query2, userID, roleID)
	err = row.Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return err == nil, nil
}

// AssignRoleToUser assigne un rôle directement à un utilisateur
func (m *LocalManager) AssignRoleToUser(ctx context.Context, userID string, roleID string) error {
	// Vérifier que le rôle existe
	var exists bool
	err := m.db.QueryRowContext(ctx, "SELECT 1 FROM roles WHERE id = $1", roleID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("role not found")
		}
		return err
	}

	// Insérer la relation user-role
	_, err = m.db.ExecContext(ctx,
		"INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		userID, roleID)

	return err
}

// RemoveRoleFromUser supprime un rôle d'un utilisateur
func (m *LocalManager) RemoveRoleFromUser(ctx context.Context, userID string, roleID string) error {
	_, err := m.db.ExecContext(ctx,
		"DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2",
		userID, roleID)

	return err
}

// GetUserDirectRoles récupère les rôles assignés directement à un utilisateur
func (m *LocalManager) GetUserDirectRoles(ctx context.Context, userID string) ([]Role, error) {
	query := `
        SELECT r.id, r.name, r.description
        FROM roles r
        JOIN user_roles ur ON r.id = ur.role_id
        WHERE ur.user_id = $1
    `

	rows, err := m.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetAllGroups récupère tous les groupes avec leurs rôles
func (m *LocalManager) GetAllGroups(ctx context.Context) ([]Group, error) {
	// D'abord récupérer tous les groupes
	query := `SELECT id, name, description FROM groups`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []Group
	groupMap := make(map[string]*Group)

	for rows.Next() {
		var group Group
		if err := rows.Scan(&group.ID, &group.Name, &group.Description); err != nil {
			return nil, err
		}
		groups = append(groups, group)
		groupMap[group.ID] = &groups[len(groups)-1]
	}

	// Ensuite récupérer les rôles pour chaque groupe
	for _, group := range groups {
		roles, err := m.getGroupRoles(ctx, group.ID)
		if err != nil {
			return nil, err
		}
		groupMap[group.ID].Roles = roles
	}

	return groups, nil
}

// Fonction helper pour récupérer les rôles d'un groupe
func (m *LocalManager) getGroupRoles(ctx context.Context, groupID string) ([]Role, error) {
	query := `
        SELECT r.id, r.name, r.description
        FROM roles r
        JOIN group_roles gr ON r.id = gr.role_id
        WHERE gr.group_id = $1
    `

	rows, err := m.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// GetUserGroups récupère les groupes auxquels appartient un utilisateur
func (m *LocalManager) GetUserGroups(ctx context.Context, userID string) ([]Group, error) {
	query := `
        SELECT g.id, g.name, g.description
        FROM groups g
        JOIN user_groups ug ON g.id = ug.group_id
        WHERE ug.user_id = $1
    `

	rows, err := m.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []Group
	groupMap := make(map[string]*Group)

	for rows.Next() {
		var group Group
		if err := rows.Scan(&group.ID, &group.Name, &group.Description); err != nil {
			return nil, err
		}
		groups = append(groups, group)
		groupMap[group.ID] = &groups[len(groups)-1]
	}

	// Récupérer les rôles pour chaque groupe
	for _, group := range groups {
		roles, err := m.getGroupRoles(ctx, group.ID)
		if err != nil {
			return nil, err
		}
		groupMap[group.ID].Roles = roles
	}

	return groups, nil
}

// AssignUserToGroup assigne un utilisateur à un groupe
func (m *LocalManager) AssignUserToGroup(ctx context.Context, userID string, groupID string) error {
	// Vérifier que le groupe existe
	var exists bool
	err := m.db.QueryRowContext(ctx, "SELECT 1 FROM groups WHERE id = $1", groupID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("group not found")
		}
		return err
	}

	// Insérer la relation user-group
	_, err = m.db.ExecContext(ctx,
		"INSERT INTO user_groups (user_id, group_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		userID, groupID)

	return err
}

// RemoveUserFromGroup supprime un utilisateur d'un groupe
func (m *LocalManager) RemoveUserFromGroup(ctx context.Context, userID string, groupID string) error {
	_, err := m.db.ExecContext(ctx,
		"DELETE FROM user_groups WHERE user_id = $1 AND group_id = $2",
		userID, groupID)

	return err
}

// Fonctions CRUD pour les rôles

// CreateRole crée un nouveau rôle
func (m *LocalManager) CreateRole(ctx context.Context, name string, description string) (*Role, error) {
	id := uuid.New().String()

	_, err := m.db.ExecContext(ctx,
		"INSERT INTO roles (id, name, description) VALUES ($1, $2, $3)",
		id, name, description)
	if err != nil {
		return nil, err
	}

	return &Role{
		ID:          id,
		Name:        name,
		Description: description,
	}, nil
}

// UpdateRole met à jour un rôle existant
func (m *LocalManager) UpdateRole(ctx context.Context, id string, name string, description string) error {
	result, err := m.db.ExecContext(ctx,
		"UPDATE roles SET name = $1, description = $2 WHERE id = $3",
		name, description, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("role not found")
	}

	return nil
}

// DeleteRole supprime un rôle
func (m *LocalManager) DeleteRole(ctx context.Context, id string) error {
	// Les contraintes ON DELETE CASCADE supprimeront automatiquement
	// les entrées dans user_roles et group_roles
	result, err := m.db.ExecContext(ctx, "DELETE FROM roles WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("role not found")
	}

	return nil
}

// Fonctions CRUD pour les groupes

// CreateGroup crée un nouveau groupe
func (m *LocalManager) CreateGroup(ctx context.Context, name string, description string) (*Group, error) {
	id := uuid.New().String()

	_, err := m.db.ExecContext(ctx,
		"INSERT INTO groups (id, name, description) VALUES ($1, $2, $3)",
		id, name, description)
	if err != nil {
		return nil, err
	}

	return &Group{
		ID:          id,
		Name:        name,
		Description: description,
		Roles:       []Role{},
	}, nil
}

// UpdateGroup met à jour un groupe existant
func (m *LocalManager) UpdateGroup(ctx context.Context, id string, name string, description string) error {
	result, err := m.db.ExecContext(ctx,
		"UPDATE groups SET name = $1, description = $2 WHERE id = $3",
		name, description, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("group not found")
	}

	return nil
}

// DeleteGroup supprime un groupe
func (m *LocalManager) DeleteGroup(ctx context.Context, id string) error {
	// Les contraintes ON DELETE CASCADE supprimeront automatiquement
	// les entrées dans user_groups et group_roles
	result, err := m.db.ExecContext(ctx, "DELETE FROM groups WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("group not found")
	}

	return nil
}

// AddRoleToGroup ajoute un rôle à un groupe
func (m *LocalManager) AddRoleToGroup(ctx context.Context, groupID string, roleID string) error {
	// Vérifier que le rôle et le groupe existent
	var exists bool

	err := m.db.QueryRowContext(ctx, "SELECT 1 FROM roles WHERE id = $1", roleID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("role not found")
		}
		return err
	}

	err = m.db.QueryRowContext(ctx, "SELECT 1 FROM groups WHERE id = $1", groupID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("group not found")
		}
		return err
	}

	// Ajouter le rôle au groupe
	_, err = m.db.ExecContext(ctx,
		"INSERT INTO group_roles (group_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		groupID, roleID)

	return err
}

// RemoveRoleFromGroup supprime un rôle d'un groupe
func (m *LocalManager) RemoveRoleFromGroup(ctx context.Context, groupID string, roleID string) error {
	_, err := m.db.ExecContext(ctx,
		"DELETE FROM group_roles WHERE group_id = $1 AND role_id = $2",
		groupID, roleID)

	return err
}
