package middlewares

import (
	"net/http"
	"time"
	"zenauth/config"

	"github.com/golang-jwt/jwt"
)

// AdminAuthMiddleware checks for a valid JWT token in cookies for admin routes
func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip middleware for login and logout pages
		if r.URL.Path == "/admin/login" || r.URL.Path == "/admin/login/submit" || r.URL.Path == "/admin/logout" {
			next.ServeHTTP(w, r)
			return
		}

		// Get token from cookie
		cookie, err := r.Cookie("admin_token")
		if err != nil {
			// Redirect to login page if no token found
			http.Redirect(w, r, "/admin/login?error=Please+login+to+continue", http.StatusSeeOther)
			return
		}

		// Parse and verify the token
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			// Make sure the signing method is what we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}

			return []byte(config.App.Admin.JWTSecret), nil
		})

		// Handle token validation errors
		if err != nil || !token.Valid {
			// Clear the invalid cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "admin_token",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
			})

			http.Redirect(w, r, "/admin/login?error=Session+expired", http.StatusSeeOther)
			return
		}

		// Check admin claim in the token
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// Check if token has admin claim
			if admin, ok := claims["admin"].(bool); !ok || !admin {
				http.Redirect(w, r, "/admin/login?error=Not+authorized", http.StatusSeeOther)
				return
			}

			// Check if token has expired
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					http.Redirect(w, r, "/admin/login?error=Session+expired", http.StatusSeeOther)
					return
				}
			}

			// Optionally add user info to request context for use in handlers
			// ctx := context.WithValue(r.Context(), "user", claims)
			// r = r.WithContext(ctx)
		} else {
			http.Redirect(w, r, "/admin/login?error=Invalid+session", http.StatusSeeOther)
			return
		}

		// Continue to the next handler
		next.ServeHTTP(w, r)
	})
}
