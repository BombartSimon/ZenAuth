package adapters

import (
	"context"
	"database/sql"
	"fmt"
)

// ExternalRoleConfig contient la configuration pour le manager externe
type ExternalRoleConfig struct {
	// Tables
	RoleTable      string
	GroupTable     string
	UserRoleTable  string
	GroupRoleTable string
	UserGroupTable string

	// Colonnes rôles
	RoleIDCol   string
	RoleNameCol string
	RoleDescCol string

	// Colonnes groupes
	GroupIDCol   string
	GroupNameCol string
	GroupDescCol string

	// Colonnes relations
	UserRoleUserCol   string
	UserRoleRoleCol   string
	GroupRoleGroupCol string
	GroupRoleRoleCol  string
	UserGroupUserCol  string
	UserGroupGroupCol string
}

// ExternalManager implémente Manager pour les providers SQL externes
type ExternalManager struct {
	db     *sql.DB
	config ExternalRoleConfig
}

// NewExternalManager crée un nouveau gestionnaire externe de rôles
func NewExternalManager(db *sql.DB, config ExternalRoleConfig) *ExternalManager {
	return &ExternalManager{
		db:     db,
		config: config,
	}
}

// GetAllRoles récupère tous les rôles disponibles
func (m *ExternalManager) GetAllRoles(ctx context.Context) ([]Role, error) {
	query := fmt.Sprintf(
		"SELECT %s, %s, %s FROM %s",
		m.config.RoleIDCol,
		m.config.RoleNameCol,
		m.config.RoleDescCol,
		m.config.RoleTable,
	)

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var desc sql.NullString // Pour gérer les descriptions NULL

		if err := rows.Scan(&role.ID, &role.Name, &desc); err != nil {
			return nil, err
		}

		if desc.Valid {
			role.Description = desc.String
		}

		roles = append(roles, role)
	}

	return roles, nil
}

// GetUserRoles récupère tous les rôles d'un utilisateur
func (m *ExternalManager) GetUserRoles(ctx context.Context, userID string) ([]Role, error) {
	// Rôles directs
	directRoles, err := m.GetUserDirectRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Rôles via groupes
	query := fmt.Sprintf(`
        SELECT DISTINCT r.%s, r.%s, r.%s
        FROM %s r
        JOIN %s gr ON r.%s = gr.%s
        JOIN %s ug ON gr.%s = ug.%s
        WHERE ug.%s = $1
    `,
		m.config.RoleIDCol, m.config.RoleNameCol, m.config.RoleDescCol,
		m.config.RoleTable,
		m.config.GroupRoleTable, m.config.RoleIDCol, m.config.GroupRoleRoleCol,
		m.config.UserGroupTable, m.config.GroupRoleGroupCol, m.config.UserGroupGroupCol,
		m.config.UserGroupUserCol,
	)

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

	// Ajouter ensuite les rôles via groupes
	for rows.Next() {
		var role Role
		var desc sql.NullString

		if err := rows.Scan(&role.ID, &role.Name, &desc); err != nil {
			return nil, err
		}

		if desc.Valid {
			role.Description = desc.String
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

// HasRole vérifie si un utilisateur a un rôle
func (m *ExternalManager) HasRole(ctx context.Context, userID string, roleID string) (bool, error) {
	// Vérifier rôle direct
	query1 := fmt.Sprintf(
		"SELECT 1 FROM %s WHERE %s = $1 AND %s = $2 LIMIT 1",
		m.config.UserRoleTable,
		m.config.UserRoleUserCol,
		m.config.UserRoleRoleCol,
	)

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
	query2 := fmt.Sprintf(`
        SELECT 1 FROM %s gr
        JOIN %s ug ON gr.%s = ug.%s
        WHERE ug.%s = $1 AND gr.%s = $2
        LIMIT 1
    `,
		m.config.GroupRoleTable,
		m.config.UserGroupTable, m.config.GroupRoleGroupCol, m.config.UserGroupGroupCol,
		m.config.UserGroupUserCol, m.config.GroupRoleRoleCol,
	)

	row = m.db.QueryRowContext(ctx, query2, userID, roleID)
	err = row.Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	return err == nil, nil
}

// Implémentation des autres méthodes de l'interface Manager...
// (Par souci de concision, je ne mets pas toutes les méthodes ici,
// mais vous aurez besoin de les implémenter toutes pour satisfaire l'interface)

// GetUserDirectRoles récupère les rôles directs d'un utilisateur
func (m *ExternalManager) GetUserDirectRoles(ctx context.Context, userID string) ([]Role, error) {
	query := fmt.Sprintf(`
        SELECT r.%s, r.%s, r.%s
        FROM %s r
        JOIN %s ur ON r.%s = ur.%s
        WHERE ur.%s = $1
    `,
		m.config.RoleIDCol, m.config.RoleNameCol, m.config.RoleDescCol,
		m.config.RoleTable,
		m.config.UserRoleTable, m.config.RoleIDCol, m.config.UserRoleRoleCol,
		m.config.UserRoleUserCol,
	)

	rows, err := m.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var desc sql.NullString

		if err := rows.Scan(&role.ID, &role.Name, &desc); err != nil {
			return nil, err
		}

		if desc.Valid {
			role.Description = desc.String
		}

		roles = append(roles, role)
	}

	return roles, nil
}

// AssignRoleToUser assigne un rôle à un utilisateur
func (m *ExternalManager) AssignRoleToUser(ctx context.Context, userID string, roleID string) error {
	query := fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		m.config.UserRoleTable,
		m.config.UserRoleUserCol,
		m.config.UserRoleRoleCol,
	)

	_, err := m.db.ExecContext(ctx, query, userID, roleID)
	return err
}

// RemoveRoleFromUser supprime un rôle d'un utilisateur
func (m *ExternalManager) RemoveRoleFromUser(ctx context.Context, userID string, roleID string) error {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1 AND %s = $2",
		m.config.UserRoleTable,
		m.config.UserRoleUserCol,
		m.config.UserRoleRoleCol,
	)

	_, err := m.db.ExecContext(ctx, query, userID, roleID)
	return err
}

// CreateRole crée un nouveau rôle
func (m *ExternalManager) CreateRole(ctx context.Context, name string, description string) (*Role, error) {
	query := fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES ($1, $2) RETURNING %s",
		m.config.RoleTable,
		m.config.RoleNameCol,
		m.config.RoleDescCol,
		m.config.RoleIDCol,
	)

	var id string
	err := m.db.QueryRowContext(ctx, query, name, description).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &Role{
		ID:          id,
		Name:        name,
		Description: description,
	}, nil
}

// GetRole récupère un rôle par son ID
func (m *ExternalManager) GetRole(ctx context.Context, id string) (Role, error) {
	query := fmt.Sprintf(
		"SELECT %s, %s, %s FROM %s WHERE %s = $1",
		m.config.RoleIDCol,
		m.config.RoleNameCol,
		m.config.RoleDescCol,
		m.config.RoleTable,
		m.config.RoleIDCol,
	)

	var role Role
	var desc sql.NullString
	err := m.db.QueryRowContext(ctx, query, id).Scan(&role.ID, &role.Name, &desc)
	if err != nil {
		return Role{}, err
	}

	if desc.Valid {
		role.Description = desc.String
	}

	return role, nil
}

// UpdateRole met à jour un rôle existant
func (m *ExternalManager) UpdateRole(ctx context.Context, id string, name string, description string) error {
	query := fmt.Sprintf(
		"UPDATE %s SET %s = $1, %s = $2 WHERE %s = $3",
		m.config.RoleTable,
		m.config.RoleNameCol,
		m.config.RoleDescCol,
		m.config.RoleIDCol,
	)

	_, err := m.db.ExecContext(ctx, query, name, description, id)
	return err
}

// DeleteRole supprime un rôle
func (m *ExternalManager) DeleteRole(ctx context.Context, id string) error {
	// Commencer une transaction pour s'assurer que toutes les suppressions sont atomiques
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Nettoyer d'abord les relations
	// Supprimer des user_role
	query1 := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1",
		m.config.UserRoleTable,
		m.config.UserRoleRoleCol,
	)
	_, err = tx.ExecContext(ctx, query1, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Supprimer des group_role
	query2 := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1",
		m.config.GroupRoleTable,
		m.config.GroupRoleRoleCol,
	)
	_, err = tx.ExecContext(ctx, query2, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Enfin, supprimer le rôle lui-même
	query3 := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1",
		m.config.RoleTable,
		m.config.RoleIDCol,
	)
	_, err = tx.ExecContext(ctx, query3, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetAllGroups récupère tous les groupes
func (m *ExternalManager) GetAllGroups(ctx context.Context) ([]Group, error) {
	query := fmt.Sprintf(
		"SELECT %s, %s, %s FROM %s",
		m.config.GroupIDCol,
		m.config.GroupNameCol,
		m.config.GroupDescCol,
		m.config.GroupTable,
	)

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var group Group
		var desc sql.NullString

		if err := rows.Scan(&group.ID, &group.Name, &desc); err != nil {
			return nil, err
		}

		if desc.Valid {
			group.Description = desc.String
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// CreateGroup crée un nouveau groupe
func (m *ExternalManager) CreateGroup(ctx context.Context, name string, description string) (*Group, error) {
	query := fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES ($1, $2) RETURNING %s",
		m.config.GroupTable,
		m.config.GroupNameCol,
		m.config.GroupDescCol,
		m.config.GroupIDCol,
	)

	var id string
	err := m.db.QueryRowContext(ctx, query, name, description).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &Group{
		ID:          id,
		Name:        name,
		Description: description,
		Roles:       []Role{}, // Initialiser avec un slice vide
	}, nil
}

// GetGroup récupère un groupe par son ID
func (m *ExternalManager) GetGroup(ctx context.Context, id string) (Group, error) {
	query := fmt.Sprintf(
		"SELECT %s, %s, %s FROM %s WHERE %s = $1",
		m.config.GroupIDCol,
		m.config.GroupNameCol,
		m.config.GroupDescCol,
		m.config.GroupTable,
		m.config.GroupIDCol,
	)

	var group Group
	var desc sql.NullString
	err := m.db.QueryRowContext(ctx, query, id).Scan(&group.ID, &group.Name, &desc)
	if err != nil {
		return Group{}, err
	}

	if desc.Valid {
		group.Description = desc.String
	}

	return group, nil
}

// UpdateGroup met à jour un groupe existant
func (m *ExternalManager) UpdateGroup(ctx context.Context, id string, name string, description string) error {
	query := fmt.Sprintf(
		"UPDATE %s SET %s = $1, %s = $2 WHERE %s = $3",
		m.config.GroupTable,
		m.config.GroupNameCol,
		m.config.GroupDescCol,
		m.config.GroupIDCol,
	)

	_, err := m.db.ExecContext(ctx, query, name, description, id)
	return err
}

// DeleteGroup supprime un groupe
func (m *ExternalManager) DeleteGroup(ctx context.Context, id string) error {
	// Commencer une transaction pour s'assurer que toutes les suppressions sont atomiques
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Nettoyer d'abord les relations
	// Supprimer des user_group
	query1 := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1",
		m.config.UserGroupTable,
		m.config.UserGroupGroupCol,
	)
	_, err = tx.ExecContext(ctx, query1, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Supprimer des group_role
	query2 := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1",
		m.config.GroupRoleTable,
		m.config.GroupRoleGroupCol,
	)
	_, err = tx.ExecContext(ctx, query2, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Enfin, supprimer le groupe lui-même
	query3 := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1",
		m.config.GroupTable,
		m.config.GroupIDCol,
	)
	_, err = tx.ExecContext(ctx, query3, id)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetUserGroups récupère tous les groupes d'un utilisateur
func (m *ExternalManager) GetUserGroups(ctx context.Context, userID string) ([]Group, error) {
	query := fmt.Sprintf(`
        SELECT g.%s, g.%s, g.%s
        FROM %s g
        JOIN %s ug ON g.%s = ug.%s
        WHERE ug.%s = $1
    `,
		m.config.GroupIDCol, m.config.GroupNameCol, m.config.GroupDescCol,
		m.config.GroupTable,
		m.config.UserGroupTable, m.config.GroupIDCol, m.config.UserGroupGroupCol,
		m.config.UserGroupUserCol,
	)

	rows, err := m.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var group Group
		var desc sql.NullString

		if err := rows.Scan(&group.ID, &group.Name, &desc); err != nil {
			return nil, err
		}

		if desc.Valid {
			group.Description = desc.String
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// AssignUserToGroup ajoute un utilisateur à un groupe
func (m *ExternalManager) AssignUserToGroup(ctx context.Context, userID string, groupID string) error {
	query := fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		m.config.UserGroupTable,
		m.config.UserGroupUserCol,
		m.config.UserGroupGroupCol,
	)

	_, err := m.db.ExecContext(ctx, query, userID, groupID)
	return err
}

// RemoveUserFromGroup retire un utilisateur d'un groupe
func (m *ExternalManager) RemoveUserFromGroup(ctx context.Context, userID string, groupID string) error {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1 AND %s = $2",
		m.config.UserGroupTable,
		m.config.UserGroupUserCol,
		m.config.UserGroupGroupCol,
	)

	_, err := m.db.ExecContext(ctx, query, userID, groupID)
	return err
}

// GetGroupRoles récupère tous les rôles d'un groupe
func (m *ExternalManager) GetGroupRoles(ctx context.Context, groupID string) ([]Role, error) {
	query := fmt.Sprintf(`
        SELECT r.%s, r.%s, r.%s
        FROM %s r
        JOIN %s gr ON r.%s = gr.%s
        WHERE gr.%s = $1
    `,
		m.config.RoleIDCol, m.config.RoleNameCol, m.config.RoleDescCol,
		m.config.RoleTable,
		m.config.GroupRoleTable, m.config.RoleIDCol, m.config.GroupRoleRoleCol,
		m.config.GroupRoleGroupCol,
	)

	rows, err := m.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		var desc sql.NullString

		if err := rows.Scan(&role.ID, &role.Name, &desc); err != nil {
			return nil, err
		}

		if desc.Valid {
			role.Description = desc.String
		}

		roles = append(roles, role)
	}

	return roles, nil
}

// AssignRoleToGroup ajoute un rôle à un groupe
func (m *ExternalManager) AssignRoleToGroup(ctx context.Context, groupID string, roleID string) error {
	query := fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		m.config.GroupRoleTable,
		m.config.GroupRoleGroupCol,
		m.config.GroupRoleRoleCol,
	)

	_, err := m.db.ExecContext(ctx, query, groupID, roleID)
	return err
}

// RemoveRoleFromGroup retire un rôle d'un groupe
func (m *ExternalManager) RemoveRoleFromGroup(ctx context.Context, groupID string, roleID string) error {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1 AND %s = $2",
		m.config.GroupRoleTable,
		m.config.GroupRoleGroupCol,
		m.config.GroupRoleRoleCol,
	)

	_, err := m.db.ExecContext(ctx, query, groupID, roleID)
	return err
}

// GetRoleGroups récupère tous les groupes ayant un rôle spécifique
func (m *ExternalManager) GetRoleGroups(ctx context.Context, roleID string) ([]Group, error) {
	query := fmt.Sprintf(`
        SELECT g.%s, g.%s, g.%s
        FROM %s g
        JOIN %s gr ON g.%s = gr.%s
        WHERE gr.%s = $1
    `,
		m.config.GroupIDCol, m.config.GroupNameCol, m.config.GroupDescCol,
		m.config.GroupTable,
		m.config.GroupRoleTable, m.config.GroupIDCol, m.config.GroupRoleGroupCol,
		m.config.GroupRoleRoleCol,
	)

	rows, err := m.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var group Group
		var desc sql.NullString

		if err := rows.Scan(&group.ID, &group.Name, &desc); err != nil {
			return nil, err
		}

		if desc.Valid {
			group.Description = desc.String
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// GetRoleUsers récupère tous les utilisateurs ayant un rôle spécifique (directement)
func (m *ExternalManager) GetRoleUsers(ctx context.Context, roleID string) ([]string, error) {
	query := fmt.Sprintf(`
        SELECT %s
        FROM %s
        WHERE %s = $1
    `,
		m.config.UserRoleUserCol,
		m.config.UserRoleTable,
		m.config.UserRoleRoleCol,
	)

	rows, err := m.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// GetGroupUsers récupère tous les utilisateurs appartenant à un groupe
func (m *ExternalManager) GetGroupUsers(ctx context.Context, groupID string) ([]string, error) {
	query := fmt.Sprintf(`
        SELECT %s
        FROM %s
        WHERE %s = $1
    `,
		m.config.UserGroupUserCol,
		m.config.UserGroupTable,
		m.config.UserGroupGroupCol,
	)

	rows, err := m.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// AddRoleToGroup ajoute un rôle à un groupe (alias pour AssignRoleToGroup)
func (m *ExternalManager) AddRoleToGroup(ctx context.Context, groupID string, roleID string) error {
	return m.AssignRoleToGroup(ctx, groupID, roleID)
}
