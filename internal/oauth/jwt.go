package oauth

import (
	"context"
	"time"
	"zenauth/config"
	rProviders "zenauth/internal/adapters/role"

	"github.com/golang-jwt/jwt"
)

// GenerateAccessToken crée un nouveau JWT token d'accès
func GenerateAccessToken(subject string, scope string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   subject,
		"aud":   "zenauth",
		"scope": scope,
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}

	// Inclure les rôles dans le JWT si configuré
	if config.App.RoleManager.IncludeRolesInJWT && rProviders.CurrentManager != nil {
		ctx := context.Background()

		// Récupérer les rôles de l'utilisateur
		roles, err := rProviders.CurrentManager.GetUserRoles(ctx, subject)
		if err == nil && len(roles) > 0 {
			// // Extraire uniquement les IDs des rôles pour le JWT
			// roleIDs := make([]string, len(roles))
			// for i, r := range roles {
			// 	roleIDs[i] = r.ID
			// }

			// claims["roles"] = roleIDs

			roleNames := make([]string, len(roles))
			for i, r := range roles {
				roleNames[i] = r.Name
			}
			claims["roles"] = roleNames
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.App.JWTSecret))
}

func ValidateAccessToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.App.JWTSecret), nil
	})
}
