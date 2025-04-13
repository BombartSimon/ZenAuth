package adapters

import (
	"context"
)

type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Group struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Roles       []Role `json:"roles"`
}

type RoleSource interface {
	GetAllRoles(ctx context.Context) ([]Role, error)

	GetUserRoles(ctx context.Context, userID string) ([]Role, error)

	HasRole(ctx context.Context, userID string, roleID string) (bool, error)

	AssignRoleToUser(ctx context.Context, userID string, roleID string) error
	RemoveRoleFromUser(ctx context.Context, userID string, roleID string) error
	GetUserDirectRoles(ctx context.Context, userID string) ([]Role, error)

	GetAllGroups(ctx context.Context) ([]Group, error)
	GetUserGroups(ctx context.Context, userID string) ([]Group, error)
	AssignUserToGroup(ctx context.Context, userID string, groupID string) error
	RemoveUserFromGroup(ctx context.Context, userID string, groupID string) error

	CreateRole(ctx context.Context, name string, description string) (*Role, error)
	UpdateRole(ctx context.Context, id string, name string, description string) error
	DeleteRole(ctx context.Context, id string) error

	CreateGroup(ctx context.Context, name string, description string) (*Group, error)
	UpdateGroup(ctx context.Context, id string, name string, description string) error
	DeleteGroup(ctx context.Context, id string) error
	AddRoleToGroup(ctx context.Context, groupID string, roleID string) error
	RemoveRoleFromGroup(ctx context.Context, groupID string, roleID string) error
}

var CurrentManager RoleSource

type Type string

const (
	TypeLocal Type = "local"

	TypeExternal Type = "external"
)
