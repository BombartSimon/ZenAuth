package role

import (
	"context"
)

// Role représente un rôle dans le système
type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Group représente un groupe contenant plusieurs rôles
type Group struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Roles       []Role `json:"roles"`
}

// Manager définit les opérations disponibles pour la gestion des rôles
type Manager interface {
	// Récupérer tous les rôles disponibles
	GetAllRoles(ctx context.Context) ([]Role, error)

	// Récupérer tous les rôles d'un utilisateur (directs + via groupes)
	GetUserRoles(ctx context.Context, userID string) ([]Role, error)

	// Vérifier si un utilisateur a un rôle spécifique
	HasRole(ctx context.Context, userID string, roleID string) (bool, error)

	// Rôles directs (sans passer par les groupes)
	AssignRoleToUser(ctx context.Context, userID string, roleID string) error
	RemoveRoleFromUser(ctx context.Context, userID string, roleID string) error
	GetUserDirectRoles(ctx context.Context, userID string) ([]Role, error)

	// Gestion des groupes
	GetAllGroups(ctx context.Context) ([]Group, error)
	GetUserGroups(ctx context.Context, userID string) ([]Group, error)
	AssignUserToGroup(ctx context.Context, userID string, groupID string) error
	RemoveUserFromGroup(ctx context.Context, userID string, groupID string) error

	// CRUD rôles
	CreateRole(ctx context.Context, name string, description string) (*Role, error)
	UpdateRole(ctx context.Context, id string, name string, description string) error
	DeleteRole(ctx context.Context, id string) error

	// CRUD groupes
	CreateGroup(ctx context.Context, name string, description string) (*Group, error)
	UpdateGroup(ctx context.Context, id string, name string, description string) error
	DeleteGroup(ctx context.Context, id string) error
	AddRoleToGroup(ctx context.Context, groupID string, roleID string) error
	RemoveRoleFromGroup(ctx context.Context, groupID string, roleID string) error
}

// CurrentManager contient le gestionnaire de rôles actif
var CurrentManager Manager

// Type définit le type de gestionnaire de rôles (local ou externe)
type Type string

const (
	// TypeLocal indique que les rôles sont gérés localement par ZenAuth
	TypeLocal Type = "local"

	// TypeExternal indique que les rôles sont gérés via un provider externe
	TypeExternal Type = "external"
)
