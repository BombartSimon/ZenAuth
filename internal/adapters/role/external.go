package adapters

import (
	"context"
	"database/sql"
	"fmt"
)

type ExternalRoleConfig struct {
	// Tables
	RoleTable      string
	GroupTable     string
	UserRoleTable  string
	GroupRoleTable string
	UserGroupTable string

	RoleIDCol   string
	RoleNameCol string
	RoleDescCol string

	GroupIDCol   string
	GroupNameCol string
	GroupDescCol string

	UserRoleUserCol   string
	UserRoleRoleCol   string
	GroupRoleGroupCol string
	GroupRoleRoleCol  string
	UserGroupUserCol  string
	UserGroupGroupCol string
}

// Implements the Manager interface
type ExternalManager struct {
	db     *sql.DB
	config ExternalRoleConfig
}

func NewExternalManager(db *sql.DB, config ExternalRoleConfig) *ExternalManager {
	return &ExternalManager{
		db:     db,
		config: config,
	}
}

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

func (m *ExternalManager) GetUserRoles(ctx context.Context, userID string) ([]Role, error) {

	directRoles, err := m.GetUserDirectRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

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

	roleMap := make(map[string]Role)

	for _, role := range directRoles {
		roleMap[role.ID] = role
	}

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

	result := make([]Role, 0, len(roleMap))
	for _, role := range roleMap {
		result = append(result, role)
	}

	return result, nil
}
func (m *ExternalManager) HasRole(ctx context.Context, userID string, roleID string) (bool, error) {
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

func (m *ExternalManager) DeleteRole(ctx context.Context, id string) error {

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

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
		Roles:       []Role{},
	}, nil
}

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

func (m *ExternalManager) DeleteGroup(ctx context.Context, id string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

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

func (m *ExternalManager) AddRoleToGroup(ctx context.Context, groupID string, roleID string) error {
	return m.AssignRoleToGroup(ctx, groupID, roleID)
}
